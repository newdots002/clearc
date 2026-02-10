package analyzer

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

// DirType represents the type/category of a directory
type DirType string

const (
	DirTypeSystem      DirType = "system"       // System files, don't delete
	DirTypeApplication DirType = "application"  // Installed applications
	DirTypeCache       DirType = "cache"        // Cache files, safe to delete
	DirTypeTemp        DirType = "temp"         // Temporary files, safe to delete
	DirTypeDevCache    DirType = "dev_cache"    // Development cache (node_modules, etc)
	DirTypeBuildOutput DirType = "build_output" // Build outputs (dist, build, target)
	DirTypeUserData    DirType = "user_data"    // User documents/data
	DirTypeDownloads   DirType = "downloads"    // Downloads folder
	DirTypeLogs        DirType = "logs"         // Log files
	DirTypeBackup      DirType = "backup"       // Backup files
	DirTypeUnknown     DirType = "unknown"      // Unknown type
)

// CleanRecommendation represents the cleaning recommendation level
type CleanRecommendation string

const (
	RecommendSafe    CleanRecommendation = "safe"    // Safe to delete
	RecommendCaution CleanRecommendation = "caution" // Delete with caution
	RecommendKeep    CleanRecommendation = "keep"    // Should keep
	RecommendNever   CleanRecommendation = "never"   // Never delete (system)
)

// FileNode represents a file or directory in the tree
type FileNode struct {
	Name           string              `json:"name"`
	Path           string              `json:"path"`
	Size           int64               `json:"size"`
	IsDir          bool                `json:"isDir"`
	Children       []*FileNode         `json:"children,omitempty"`
	IsProtected    bool                `json:"isProtected"` // White-listed system directories
	FileCount      int                 `json:"fileCount"`
	DirType        DirType             `json:"dirType"`        // Directory type/category
	Recommendation CleanRecommendation `json:"recommendation"` // Cleaning recommendation
	Description    string              `json:"description"`    // Description of what this directory contains
}

// Analyzer handles disk analysis
type Analyzer struct {
	mu            sync.RWMutex
	progress      int
	status        string
	whitelist     map[string]bool
	whitelistDirs []string
	minSizeBytes  int64 // Minimum file size to display (default 0)
	cache         *DirCache
}

// New creates a new Analyzer
func New() *Analyzer {
	a := &Analyzer{
		whitelist: make(map[string]bool),
		cache:     NewDirCache(),
	}
	a.initWhitelist()
	return a
}

// SaveCache saves the directory cache to disk
func (a *Analyzer) SaveCache() error {
	return a.cache.Save()
}

// ClearCache clears all cached data
func (a *Analyzer) ClearCache() {
	a.cache.Clear()
}

// GetCacheStats returns cache statistics
func (a *Analyzer) GetCacheStats() (total int, valid int) {
	return a.cache.Stats()
}

// initWhitelist initializes the whitelist of protected directories
func (a *Analyzer) initWhitelist() {
	if runtime.GOOS == "windows" {
		// Windows system directories that should not be deleted
		a.whitelistDirs = []string{
			"Windows",
			"Program Files",
			"Program Files (x86)",
			"ProgramData",
			"$WINDOWS.~BT",
			"$Windows.~WS",
			"$WinREAgent",
			"$Recycle.Bin",
			"System Volume Information",
			"Recovery",
			"Boot",
			"EFI",
			"PerfLogs",
			"MSOCache",
			"Intel",
			"AMD",
			"NVIDIA",
			"Config.Msi",
			"Documents and Settings",
		}
	} else if runtime.GOOS == "darwin" {
		a.whitelistDirs = []string{
			"System",
			"Library",
			"Applications",
			"usr",
			"bin",
			"sbin",
			"private",
			"cores",
			"dev",
			"etc",
			"tmp",
			"var",
			".Spotlight-V100",
			".fseventsd",
			".Trashes",
		}
	} else {
		// Linux
		a.whitelistDirs = []string{
			"bin",
			"boot",
			"dev",
			"etc",
			"lib",
			"lib64",
			"proc",
			"root",
			"run",
			"sbin",
			"srv",
			"sys",
			"usr",
			"var",
			"lost+found",
		}
	}

	for _, dir := range a.whitelistDirs {
		a.whitelist[strings.ToLower(dir)] = true
	}
}

// identifyDirType identifies the type of a directory based on its name and path
func (a *Analyzer) identifyDirType(path string, name string) (DirType, CleanRecommendation, string) {
	nameLower := strings.ToLower(name)
	pathLower := strings.ToLower(path)

	// System directories
	if a.IsProtected(path) {
		return DirTypeSystem, RecommendNever, "系统核心目录，请勿删除"
	}

	// AppData specific patterns - these are common space hogs
	if strings.Contains(pathLower, "appdata\\local") {
		appDataLocalPatterns := map[string]struct {
			dirType DirType
			rec     CleanRecommendation
			desc    string
		}{
			"temp":                          {DirTypeTemp, RecommendSafe, "Windows 临时文件，可安全删除"},
			"microsoft\\windows\\inetcache": {DirTypeCache, RecommendSafe, "IE/Edge 网络缓存，可安全删除"},
			"microsoft\\windows\\explorer":  {DirTypeCache, RecommendCaution, "资源管理器缩略图缓存"},
			"microsoft\\windows\\wer":       {DirTypeLogs, RecommendSafe, "Windows 错误报告，可安全删除"},
			"crashdumps":                    {DirTypeLogs, RecommendSafe, "程序崩溃转储，可安全删除"},
			"d3dscache":                     {DirTypeCache, RecommendSafe, "DirectX 着色器缓存，可安全删除"},
			"pip":                           {DirTypeDevCache, RecommendSafe, "Python pip 缓存，可安全删除"},
			"yarn":                          {DirTypeDevCache, RecommendSafe, "Yarn 缓存，可安全删除"},
			"pnpm":                          {DirTypeDevCache, RecommendSafe, "pnpm 缓存，可安全删除"},
			"nuget":                         {DirTypeDevCache, RecommendSafe, "NuGet 包缓存，可安全删除"},
			"docker":                        {DirTypeDevCache, RecommendCaution, "Docker 数据，删除会丢失容器和镜像"},
			"packages":                      {DirTypeCache, RecommendCaution, "应用包缓存，可能影响应用启动速度"},
		}
		for pattern, info := range appDataLocalPatterns {
			if strings.Contains(pathLower, pattern) || nameLower == pattern {
				return info.dirType, info.rec, info.desc
			}
		}
		// Browser specific caches
		browserCaches := []string{"google\\chrome", "microsoft\\edge", "mozilla\\firefox", "opera", "brave"}
		for _, browser := range browserCaches {
			if strings.Contains(pathLower, browser) {
				if strings.Contains(nameLower, "cache") || nameLower == "code cache" || nameLower == "gpucache" {
					return DirTypeCache, RecommendSafe, "浏览器缓存，可安全删除"
				}
				if nameLower == "user data" || nameLower == "default" {
					return DirTypeUserData, RecommendCaution, "浏览器用户数据，包含历史记录和设置"
				}
			}
		}
	}

	if strings.Contains(pathLower, "appdata\\roaming") {
		appDataRoamingPatterns := map[string]struct {
			dirType DirType
			rec     CleanRecommendation
			desc    string
		}{
			"npm-cache": {DirTypeDevCache, RecommendSafe, "npm 全局缓存，可安全删除"},
			"npm":       {DirTypeDevCache, RecommendCaution, "npm 配置和全局包"},
			"cursor":    {DirTypeUserData, RecommendCaution, "Cursor 编辑器数据"},
			"code":      {DirTypeUserData, RecommendCaution, "VS Code 用户数据"},
			"discord":   {DirTypeCache, RecommendCaution, "Discord 缓存和数据"},
			"slack":     {DirTypeCache, RecommendCaution, "Slack 缓存和数据"},
			"tencent":   {DirTypeUserData, RecommendCaution, "腾讯软件数据（QQ/微信等）"},
			"wechat":    {DirTypeUserData, RecommendKeep, "微信聊天记录和文件"},
			"feishu":    {DirTypeUserData, RecommendCaution, "飞书数据"},
			"dingtalk":  {DirTypeUserData, RecommendCaution, "钉钉数据"},
		}
		for pattern, info := range appDataRoamingPatterns {
			if strings.Contains(pathLower, pattern) || nameLower == pattern {
				return info.dirType, info.rec, info.desc
			}
		}
	}

	// Development cache patterns
	devCachePatterns := map[string]string{
		"node_modules":  "Node.js 依赖包目录",
		".npm":          "npm 缓存目录",
		".yarn":         "Yarn 缓存目录",
		".pnpm-store":   "pnpm 存储目录",
		"go":            "Go 模块缓存",
		".cargo":        "Rust Cargo 缓存",
		".rustup":       "Rust 工具链",
		".gradle":       "Gradle 构建缓存",
		".m2":           "Maven 本地仓库",
		".nuget":        "NuGet 包缓存",
		".composer":     "PHP Composer 缓存",
		"__pycache__":   "Python 字节码缓存",
		".pytest_cache": "Pytest 缓存",
		".mypy_cache":   "MyPy 类型检查缓存",
		"venv":          "Python 虚拟环境",
		".venv":         "Python 虚拟环境",
		"vendor":        "依赖包目录",
		".bundle":       "Ruby Bundler 缓存",
		"pods":          "iOS CocoaPods 依赖",
		"pip":           "Python pip 缓存",
	}

	for pattern, desc := range devCachePatterns {
		if nameLower == pattern {
			return DirTypeDevCache, RecommendSafe, desc + "，可安全删除后重新安装"
		}
	}

	// Build output patterns
	buildPatterns := map[string]string{
		"dist":          "构建输出目录",
		"build":         "构建输出目录",
		"out":           "输出目录",
		"target":        "Rust/Java 构建目录",
		"bin":           "二进制输出目录",
		"obj":           ".NET 编译中间文件",
		".next":         "Next.js 构建缓存",
		".nuxt":         "Nuxt.js 构建缓存",
		".output":       "框架输出目录",
		".turbo":        "Turborepo 缓存",
		".parcel-cache": "Parcel 构建缓存",
		".webpack":      "Webpack 缓存",
		".vite":         "Vite 缓存",
	}

	for pattern, desc := range buildPatterns {
		if nameLower == pattern {
			return DirTypeBuildOutput, RecommendSafe, desc + "，可删除后重新构建"
		}
	}

	// Cache directories
	cachePatterns := map[string]string{
		"cache":       "缓存目录",
		"caches":      "缓存目录",
		".cache":      "应用缓存",
		"cacheddata":  "缓存数据",
		"code cache":  "代码缓存",
		"gpucache":    "GPU 缓存",
		"shadercache": "着色器缓存",
		"thumbnails":  "缩略图缓存",
		"thumbs":      "缩略图缓存",
	}

	for pattern, desc := range cachePatterns {
		if nameLower == pattern || strings.Contains(nameLower, "cache") {
			return DirTypeCache, RecommendSafe, desc + "，可安全删除"
		}
	}

	// Temp directories
	tempPatterns := map[string]string{
		"temp":     "临时文件目录",
		"tmp":      "临时文件目录",
		".tmp":     "临时文件目录",
		"tempdata": "临时数据",
	}

	for pattern, desc := range tempPatterns {
		if nameLower == pattern {
			return DirTypeTemp, RecommendSafe, desc + "，可安全删除"
		}
	}

	// Log directories
	logPatterns := map[string]string{
		"logs":        "日志目录",
		"log":         "日志目录",
		".logs":       "日志目录",
		"crashdumps":  "崩溃转储",
		"crashpad":    "崩溃报告",
		"diagnostics": "诊断数据",
	}

	for pattern, desc := range logPatterns {
		if nameLower == pattern {
			return DirTypeLogs, RecommendCaution, desc + "，可删除但可能影响问题排查"
		}
	}

	// Backup directories
	backupPatterns := map[string]string{
		"backup":   "备份目录",
		"backups":  "备份目录",
		".backup":  "备份目录",
		"old":      "旧版本文件",
		"archive":  "归档文件",
		"archives": "归档文件",
	}

	for pattern, desc := range backupPatterns {
		if nameLower == pattern {
			return DirTypeBackup, RecommendCaution, desc + "，删除前请确认不再需要"
		}
	}

	// User data directories
	userDataPatterns := map[string]string{
		"documents": "用户文档",
		"downloads": "下载文件",
		"desktop":   "桌面文件",
		"pictures":  "图片文件",
		"videos":    "视频文件",
		"music":     "音乐文件",
		"appdata":   "应用数据",
	}

	for pattern, desc := range userDataPatterns {
		if nameLower == pattern {
			if nameLower == "downloads" {
				return DirTypeDownloads, RecommendCaution, desc + "，可能包含已不需要的下载文件"
			}
			return DirTypeUserData, RecommendKeep, desc + "，包含用户个人数据"
		}
	}

	// Application directories
	if strings.Contains(pathLower, "program files") ||
		strings.Contains(pathLower, "appdata\\local\\programs") {
		return DirTypeApplication, RecommendKeep, "已安装的应用程序"
	}

	// Check for common application data paths
	if strings.Contains(pathLower, "appdata\\local") ||
		strings.Contains(pathLower, "appdata\\roaming") {
		// Check if it looks like cache
		if strings.Contains(nameLower, "cache") {
			return DirTypeCache, RecommendSafe, "应用缓存数据"
		}
		return DirTypeUserData, RecommendCaution, "应用数据目录"
	}

	// Default to unknown
	return DirTypeUnknown, RecommendCaution, "未识别的目录类型"
}

// IsProtected checks if a path is protected
func (a *Analyzer) IsProtected(path string) bool {
	name := strings.ToLower(filepath.Base(path))
	return a.whitelist[name]
}

// GetProgress returns current progress
func (a *Analyzer) GetProgress() int {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.progress
}

// GetStatus returns current status
func (a *Analyzer) GetStatus() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.status
}

func (a *Analyzer) setProgress(p int) {
	a.mu.Lock()
	a.progress = p
	a.mu.Unlock()
}

func (a *Analyzer) setStatus(s string) {
	a.mu.Lock()
	a.status = s
	a.mu.Unlock()
}

// SetMinSize sets the minimum file/folder size to display (in bytes)
func (a *Analyzer) SetMinSize(sizeBytes int64) {
	a.mu.Lock()
	a.minSizeBytes = sizeBytes
	a.mu.Unlock()
}

// GetMinSize returns the current minimum size filter
func (a *Analyzer) GetMinSize() int64 {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.minSizeBytes
}

// SizeUpdateCallback is called when a node's size is calculated
type SizeUpdateCallback func(path string, size int64, fileCount int)

// AnalyzeDrive analyzes a drive - returns immediately with directory list, sizes calculated async
func (a *Analyzer) AnalyzeDrive(drivePath string) (*FileNode, error) {
	a.setProgress(0)
	a.setStatus("正在读取目录...")

	root := &FileNode{
		Name:        filepath.Base(drivePath),
		Path:        drivePath,
		IsDir:       true,
		IsProtected: false,
		Children:    make([]*FileNode, 0),
	}

	// Read top-level directories - this is fast
	entries, err := os.ReadDir(drivePath)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		childPath := filepath.Join(drivePath, entry.Name())
		info, err := entry.Info()
		if err != nil {
			continue
		}

		name := entry.Name()
		child := &FileNode{
			Name:        name,
			Path:        childPath,
			IsDir:       entry.IsDir(),
			IsProtected: a.IsProtected(childPath),
			Size:        -1, // -1 indicates size not yet calculated
			FileCount:   0,
		}

		if entry.IsDir() {
			child.DirType, child.Recommendation, child.Description = a.identifyDirType(childPath, name)
			// Check cache for directory size
			if cached, ok := a.cache.Get(childPath); ok {
				child.Size = cached.Size
				child.FileCount = cached.FileCount
			}
		} else {
			child.Size = info.Size()
			child.FileCount = 1
			child.DirType = DirTypeUnknown
			child.Recommendation = RecommendCaution
			child.Description = "文件"
		}

		root.Children = append(root.Children, child)
	}

	a.setProgress(100)
	a.setStatus("目录列表已加载")

	return root, nil
}

// QuickScan scans important directories like Users, caches, temp folders
func (a *Analyzer) QuickScan() (*FileNode, error) {
	a.setProgress(0)
	a.setStatus("快速扫描重点目录...")

	root := &FileNode{
		Name:        "重点目录",
		Path:        "C:\\",
		IsDir:       true,
		IsProtected: false,
		Children:    make([]*FileNode, 0),
	}

	// Important directories to scan
	importantDirs := []struct {
		path        string
		description string
	}{
		{"C:\\Users", "用户数据目录"},
		{"C:\\Windows\\Temp", "Windows 临时文件"},
		{"C:\\Windows\\SoftwareDistribution", "Windows 更新缓存"},
		{"C:\\Windows\\Prefetch", "预读取缓存"},
		{"C:\\ProgramData\\Package Cache", "程序安装缓存"},
	}

	// Add user-specific directories
	userProfile := os.Getenv("USERPROFILE")
	if userProfile != "" {
		userDirs := []struct {
			subPath     string
			description string
		}{
			{"AppData\\Local\\Temp", "用户临时文件"},
			{"AppData\\Local\\Microsoft\\Windows\\INetCache", "IE缓存"},
			{"AppData\\Local\\Google\\Chrome\\User Data\\Default\\Cache", "Chrome缓存"},
			{"AppData\\Local\\Microsoft\\Edge\\User Data\\Default\\Cache", "Edge缓存"},
			{"AppData\\Roaming\\npm-cache", "npm缓存"},
			{"AppData\\Local\\yarn\\Cache", "Yarn缓存"},
			{".gradle", "Gradle缓存"},
			{".m2", "Maven缓存"},
			{".nuget", "NuGet缓存"},
			{"go\\pkg", "Go模块缓存"},
			{".cargo", "Rust Cargo缓存"},
			{"Downloads", "下载文件夹"},
		}
		for _, ud := range userDirs {
			importantDirs = append(importantDirs, struct {
				path        string
				description string
			}{filepath.Join(userProfile, ud.subPath), ud.description})
		}
	}

	for _, dir := range importantDirs {
		info, err := os.Stat(dir.path)
		if err != nil {
			continue // Skip non-existent directories
		}
		if !info.IsDir() {
			continue
		}

		name := filepath.Base(dir.path)
		dirType, recommendation, _ := a.identifyDirType(dir.path, name)

		child := &FileNode{
			Name:           name,
			Path:           dir.path,
			IsDir:          true,
			IsProtected:    a.IsProtected(dir.path),
			Size:           -1,
			FileCount:      0,
			DirType:        dirType,
			Recommendation: recommendation,
			Description:    dir.description,
		}

		// Check cache for directory size
		if cached, ok := a.cache.Get(dir.path); ok {
			child.Size = cached.Size
			child.FileCount = cached.FileCount
		}

		root.Children = append(root.Children, child)
	}

	a.setProgress(100)
	a.setStatus("重点目录列表已加载")

	return root, nil
}

// CalculateSizesAsync calculates sizes for all children asynchronously
// skipSystem: if true, skip system/protected directories
func (a *Analyzer) CalculateSizesAsync(nodes []*FileNode, skipSystem bool, callback SizeUpdateCallback) {
	a.setProgress(0)
	a.setStatus("正在计算文件夹大小...")

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 8) // Limit concurrent calculations

	// Filter nodes if skipSystem is true
	var nodesToProcess []*FileNode
	skippedCount := 0
	for _, node := range nodes {
		if skipSystem && (node.IsProtected || node.Recommendation == RecommendNever) {
			skippedCount++
			continue
		}
		nodesToProcess = append(nodesToProcess, node)
	}

	total := len(nodesToProcess)
	if total == 0 {
		a.setProgress(100)
		a.setStatus(fmt.Sprintf("分析完成 (跳过系统目录: %d)", skippedCount))
		return
	}

	completed := 0
	cachedCount := 0
	var mu sync.Mutex

	for _, node := range nodesToProcess {
		if !node.IsDir {
			mu.Lock()
			completed++
			a.setProgress((completed * 100) / total)
			mu.Unlock()
			continue
		}

		// If already has size from cache (loaded during node creation), notify and skip
		if node.Size >= 0 {
			mu.Lock()
			cachedCount++
			completed++
			a.setProgress((completed * 100) / total)
			a.setStatus("从缓存加载: " + node.Name)
			mu.Unlock()
			if callback != nil {
				callback(node.Path, node.Size, node.FileCount)
			}
			continue
		}

		wg.Add(1)
		go func(n *FileNode) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// Use cache-aware calculation
			size, count, fromCache := a.calculateDirSizeWithCache(n.Path)
			n.Size = size
			n.FileCount = count

			mu.Lock()
			if fromCache {
				cachedCount++
				a.setStatus("从缓存加载: " + n.Name)
			} else {
				a.setStatus("正在计算: " + n.Name)
			}
			completed++
			a.setProgress((completed * 100) / total)
			mu.Unlock()

			// Notify callback
			if callback != nil {
				callback(n.Path, size, count)
			}
		}(node)
	}

	wg.Wait()

	// Save cache after calculation
	a.cache.Save()

	a.setProgress(100)
	statusMsg := "分析完成"
	if cachedCount > 0 || skippedCount > 0 {
		statusMsg = fmt.Sprintf("分析完成 (缓存: %d, 跳过: %d)", cachedCount, skippedCount)
	}
	a.setStatus(statusMsg)
}

// calculateDirSizeWithCache calculates directory size, using cache when available
func (a *Analyzer) calculateDirSizeWithCache(path string) (int64, int, bool) {
	// Check cache first
	if entry, ok := a.cache.Get(path); ok {
		return entry.Size, entry.FileCount, true // true = from cache
	}

	// Calculate size
	size, count := a.calculateDirSizeFullFast(path)

	// Store in cache
	a.cache.Set(path, size, count)

	return size, count, false // false = freshly calculated
}

// calculateDirSizeFullFast calculates full directory size with optimizations (no cache)
func (a *Analyzer) calculateDirSizeFullFast(path string) (int64, int) {
	minSize := a.GetMinSize()
	return a.calculateDirSizeSmart(path, minSize)
}

// calculateDirSizeSmart calculates directory size, skipping deep recursion for small directories
func (a *Analyzer) calculateDirSizeSmart(path string, minSize int64) (int64, int) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return 0, 0
	}

	var totalSize int64
	var totalCount int

	// First pass: count direct files
	var subDirs []os.DirEntry
	for _, entry := range entries {
		if entry.IsDir() {
			subDirs = append(subDirs, entry)
		} else {
			info, err := entry.Info()
			if err == nil {
				totalSize += info.Size()
				totalCount++
			}
		}
	}

	// If no subdirs, we're done
	if len(subDirs) == 0 {
		return totalSize, totalCount
	}

	// Process subdirectories with smart skipping
	type result struct {
		size  int64
		count int
	}
	results := make(chan result, len(subDirs))
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 8)

	for _, entry := range subDirs {
		wg.Add(1)
		go func(e os.DirEntry) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			childPath := filepath.Join(path, e.Name())

			// Quick estimate first - just count first level
			quickSize, quickCount := a.quickEstimate(childPath)

			// If quick estimate is already above minSize, do full calculation
			// If below minSize and we don't need precision, use estimate
			if minSize > 0 && quickSize < minSize/2 {
				// Small directory, use quick estimate (good enough for filtering)
				results <- result{quickSize, quickCount}
			} else {
				// Potentially large directory, calculate fully
				s, c := a.calculateDirSizeSmart(childPath, minSize)
				results <- result{s, c}
			}
		}(entry)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	for r := range results {
		totalSize += r.size
		totalCount += r.count
	}

	return totalSize, totalCount
}

// quickEstimate does a fast size estimate by only counting first level files
func (a *Analyzer) quickEstimate(path string) (int64, int) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return 0, 0
	}

	var size int64
	var count int
	var hasSubDirs bool

	for _, entry := range entries {
		if entry.IsDir() {
			hasSubDirs = true
			// Add a rough estimate for subdirs (assume average 10MB per subdir)
			size += 10 * 1024 * 1024
			count += 100
		} else {
			info, err := entry.Info()
			if err == nil {
				size += info.Size()
				count++
			}
		}
	}

	// If has subdirs, this is just an estimate
	if hasSubDirs {
		// Multiply by a factor to account for nested content
		size = size * 2
		count = count * 2
	}

	return size, count
}

// ExpandNode expands a node to show its children (fast - returns immediately, sizes calculated async)
func (a *Analyzer) ExpandNode(path string) (*FileNode, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	if !info.IsDir() {
		return nil, nil
	}

	name := filepath.Base(path)
	dirType, recommendation, description := a.identifyDirType(path, name)

	node := &FileNode{
		Name:           name,
		Path:           path,
		IsDir:          true,
		IsProtected:    a.IsProtected(path),
		Children:       make([]*FileNode, 0),
		DirType:        dirType,
		Recommendation: recommendation,
		Description:    description,
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	// Build children list, checking cache for sizes
	for _, entry := range entries {
		childPath := filepath.Join(path, entry.Name())
		childInfo, err := entry.Info()
		if err != nil {
			continue
		}

		childName := entry.Name()
		child := &FileNode{
			Name:        childName,
			Path:        childPath,
			IsDir:       entry.IsDir(),
			IsProtected: a.IsProtected(childPath),
			Size:        -1, // -1 indicates size not yet calculated
			FileCount:   0,
		}

		if entry.IsDir() {
			child.DirType, child.Recommendation, child.Description = a.identifyDirType(childPath, childName)
			// Check cache for directory size
			if cached, ok := a.cache.Get(childPath); ok {
				child.Size = cached.Size
				child.FileCount = cached.FileCount
			}
		} else {
			child.Size = childInfo.Size()
			child.FileCount = 1
			child.DirType = DirTypeUnknown
			child.Recommendation = RecommendCaution
			child.Description = "文件"
		}

		node.Children = append(node.Children, child)
	}

	return node, nil
}

// CalculateNodeSizesAsync calculates sizes for children of a specific path
func (a *Analyzer) CalculateNodeSizesAsync(children []*FileNode, callback SizeUpdateCallback) {
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 8)

	for _, child := range children {
		if !child.IsDir {
			continue
		}

		// If already has size from cache, just notify callback
		if child.Size >= 0 {
			if callback != nil {
				callback(child.Path, child.Size, child.FileCount)
			}
			continue
		}

		wg.Add(1)
		go func(n *FileNode) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// Use cache-aware calculation
			size, count, _ := a.calculateDirSizeWithCache(n.Path)
			n.Size = size
			n.FileCount = count

			if callback != nil {
				callback(n.Path, size, count)
			}
		}(child)
	}

	wg.Wait()

	// Save cache
	a.cache.Save()
}

// DeletePaths deletes the specified paths
func (a *Analyzer) DeletePaths(paths []string) (int64, int, []string) {
	var deletedSize int64
	var deletedCount int
	var errors []string

	for _, path := range paths {
		// Check if protected
		if a.IsProtected(path) {
			errors = append(errors, "受保护的目录: "+path)
			continue
		}

		// Get size before deletion
		info, err := os.Stat(path)
		if err != nil {
			errors = append(errors, "无法访问: "+path)
			continue
		}

		var size int64
		var count int
		if info.IsDir() {
			size, count = a.calculateDirSizeFullFast(path)
		} else {
			size = info.Size()
			count = 1
		}

		// Delete
		err = os.RemoveAll(path)
		if err != nil {
			errors = append(errors, "删除失败: "+path+" - "+err.Error())
			continue
		}

		// Invalidate cache for this path and parent directories
		a.cache.Invalidate(path)
		// Also invalidate parent directory since its size changed
		parentDir := filepath.Dir(path)
		a.cache.Invalidate(parentDir)

		deletedSize += size
		deletedCount += count
	}

	// Save cache after deletion
	a.cache.Save()

	return deletedSize, deletedCount, errors
}

// GetWhitelistDirs returns the list of protected directories
func (a *Analyzer) GetWhitelistDirs() []string {
	return a.whitelistDirs
}
