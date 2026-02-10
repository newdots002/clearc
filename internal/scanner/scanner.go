package scanner

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"clearc/internal/platform"
)

// Scanner handles file scanning operations
type Scanner struct {
	progress int
	status   string
	mu       sync.RWMutex
}

// New creates a new Scanner instance
func New() *Scanner {
	return &Scanner{}
}

// CategoryResult represents a category of junk files
type CategoryResult struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Size        int64    `json:"size"`
	FileCount   int      `json:"fileCount"`
	Color       string   `json:"color"`
	Selected    bool     `json:"selected"`
	Paths       []string `json:"-"`
}

// ScanResult represents the result of a scan
type ScanResult struct {
	Categories []CategoryResult `json:"categories"`
	TotalSize  int64            `json:"totalSize"`
	TotalFiles int              `json:"totalFiles"`
}

// GetProgress returns the current scan progress
func (s *Scanner) GetProgress() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.progress
}

// GetStatus returns the current scan status
func (s *Scanner) GetStatus() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.status
}

func (s *Scanner) setProgress(progress int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.progress = progress
}

func (s *Scanner) setStatus(status string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.status = status
}

// ScanAll scans all categories for junk files
func (s *Scanner) ScanAll() (*ScanResult, error) {
	s.setProgress(0)
	s.setStatus("正在初始化扫描...")

	result := &ScanResult{
		Categories: make([]CategoryResult, 0),
	}

	// Get dev cache directories
	devCaches := platform.GetDevCacheDirs()
	totalCategories := len(devCaches) + 2 // +2 for temp and browser cache
	currentCategory := 0

	// Scan temp directories
	s.setStatus("正在扫描临时文件...")
	tempResult := s.scanTempDirs()
	if tempResult.Size > 0 {
		result.Categories = append(result.Categories, tempResult)
		result.TotalSize += tempResult.Size
		result.TotalFiles += tempResult.FileCount
	}
	currentCategory++
	s.setProgress(currentCategory * 100 / totalCategories)

	// Scan browser cache
	s.setStatus("正在扫描浏览器缓存...")
	browserResult := s.scanBrowserCache()
	if browserResult.Size > 0 {
		result.Categories = append(result.Categories, browserResult)
		result.TotalSize += browserResult.Size
		result.TotalFiles += browserResult.FileCount
	}
	currentCategory++
	s.setProgress(currentCategory * 100 / totalCategories)

	// Scan dev caches
	for id, info := range devCaches {
		s.setStatus("正在扫描 " + info.Name + "...")
		catResult := s.scanDevCache(id, info)
		if catResult.Size > 0 {
			result.Categories = append(result.Categories, catResult)
			result.TotalSize += catResult.Size
			result.TotalFiles += catResult.FileCount
		}
		currentCategory++
		s.setProgress(currentCategory * 100 / totalCategories)
	}

	s.setProgress(100)
	s.setStatus("扫描完成")

	return result, nil
}

func (s *Scanner) scanTempDirs() CategoryResult {
	result := CategoryResult{
		ID:          "temp",
		Name:        "系统临时文件",
		Description: "Windows Temp 目录",
		Color:       "#22C55E",
		Selected:    true,
		Paths:       make([]string, 0),
	}

	tempDirs := platform.GetTempDirs()
	for _, dir := range tempDirs {
		if dir == "" {
			continue
		}
		size, count, paths := s.scanDirectory(dir, nil, 2)
		result.Size += size
		result.FileCount += count
		result.Paths = append(result.Paths, paths...)
	}

	return result
}

func (s *Scanner) scanBrowserCache() CategoryResult {
	result := CategoryResult{
		ID:          "browser",
		Name:        "浏览器缓存",
		Description: "Chrome / Edge / Firefox 缓存",
		Color:       "#EF4444",
		Selected:    false,
		Paths:       make([]string, 0),
	}

	cacheDirs := platform.GetCacheDirs()
	for _, dir := range cacheDirs {
		if dir == "" {
			continue
		}
		size, count, paths := s.scanDirectory(dir, nil, 2)
		result.Size += size
		result.FileCount += count
		result.Paths = append(result.Paths, paths...)
	}

	return result
}

func (s *Scanner) scanDevCache(id string, info platform.DevCacheInfo) CategoryResult {
	result := CategoryResult{
		ID:          id,
		Name:        info.Name,
		Description: info.Description,
		Color:       info.Color,
		Selected:    true,
		Paths:       make([]string, 0),
	}

	for _, searchPath := range info.SearchPaths {
		for _, pattern := range info.Patterns {
			s.setStatus("正在扫描 " + searchPath + " 中的 " + pattern + "...")
			matches := s.findMatchingDirs(searchPath, pattern, 5)
			for _, match := range matches {
				size, count, _ := s.scanDirectory(match, nil, 10)
				result.Size += size
				result.FileCount += count
				result.Paths = append(result.Paths, match)
			}
		}
	}

	return result
}

func (s *Scanner) findMatchingDirs(root string, pattern string, maxDepth int) []string {
	var matches []string

	if _, err := os.Stat(root); os.IsNotExist(err) {
		return matches
	}

	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// Calculate depth
		relPath, _ := filepath.Rel(root, path)
		depth := strings.Count(relPath, string(os.PathSeparator))
		if depth > maxDepth {
			return filepath.SkipDir
		}

		// Skip hidden directories (except the pattern itself)
		if info.IsDir() && strings.HasPrefix(info.Name(), ".") && info.Name() != pattern {
			// Allow scanning inside .npm, .yarn, etc.
			if !strings.HasPrefix(pattern, ".") {
				return filepath.SkipDir
			}
		}

		// Check if matches pattern
		if info.IsDir() && info.Name() == pattern {
			matches = append(matches, path)
			return filepath.SkipDir
		}

		// Check for glob patterns
		if strings.Contains(pattern, "*") {
			matched, _ := filepath.Match(pattern, info.Name())
			if matched {
				matches = append(matches, path)
			}
		}

		return nil
	})

	return matches
}

func (s *Scanner) scanDirectory(dir string, patterns []string, maxDepth int) (int64, int, []string) {
	var totalSize int64
	var fileCount int
	var paths []string

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return 0, 0, nil
	}

	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// Calculate depth
		relPath, _ := filepath.Rel(dir, path)
		depth := strings.Count(relPath, string(os.PathSeparator))
		if depth > maxDepth {
			return filepath.SkipDir
		}

		if !info.IsDir() {
			totalSize += info.Size()
			fileCount++
		}

		return nil
	})

	paths = append(paths, dir)
	return totalSize, fileCount, paths
}

// UnusedFile represents a file that hasn't been accessed for a long time
type UnusedFile struct {
	Path         string `json:"path"`
	Name         string `json:"name"`
	Size         int64  `json:"size"`
	LastAccessed int64  `json:"lastAccessed"`
	DaysUnused   int    `json:"daysUnused"`
	FileType     string `json:"fileType"`
}

// GetUnusedFiles returns files that haven't been accessed for specified days
func (s *Scanner) GetUnusedFiles(days int, minSize int64) ([]UnusedFile, error) {
	s.setProgress(0)
	s.setStatus("正在扫描未使用文件...")

	var unusedFiles []UnusedFile
	userHome := platform.GetUserHome()

	// Common user directories to scan
	scanDirs := []string{
		filepath.Join(userHome, "Documents"),
		filepath.Join(userHome, "Downloads"),
		filepath.Join(userHome, "Desktop"),
		filepath.Join(userHome, "Videos"),
		filepath.Join(userHome, "Pictures"),
	}

	cutoffTime := time.Now().AddDate(0, 0, -days)
	totalDirs := len(scanDirs)

	for i, dir := range scanDirs {
		s.setStatus("正在扫描 " + dir + "...")
		s.setProgress((i + 1) * 100 / totalDirs)

		if _, err := os.Stat(dir); os.IsNotExist(err) {
			continue
		}

		filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}

			if info.IsDir() {
				return nil
			}

			// Skip small files
			if info.Size() < minSize {
				return nil
			}

			// Check last access time
			atime := getAccessTime(info)
			if atime.Before(cutoffTime) {
				daysUnused := int(time.Since(atime).Hours() / 24)
				unusedFiles = append(unusedFiles, UnusedFile{
					Path:         path,
					Name:         info.Name(),
					Size:         info.Size(),
					LastAccessed: atime.Unix(),
					DaysUnused:   daysUnused,
					FileType:     getFileType(info.Name()),
				})
			}

			return nil
		})
	}

	s.setProgress(100)
	s.setStatus("扫描完成")

	// Sort by size (largest first)
	sortBySize(unusedFiles)

	return unusedFiles, nil
}

func getFileType(name string) string {
	ext := strings.ToLower(filepath.Ext(name))
	switch ext {
	case ".mp4", ".avi", ".mkv", ".mov", ".wmv":
		return "video"
	case ".mp3", ".wav", ".flac", ".aac":
		return "audio"
	case ".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp":
		return "image"
	case ".pdf", ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx":
		return "document"
	case ".zip", ".rar", ".7z", ".tar", ".gz":
		return "archive"
	case ".exe", ".msi", ".dmg", ".deb", ".rpm":
		return "installer"
	default:
		return "other"
	}
}

func sortBySize(files []UnusedFile) {
	for i := 0; i < len(files)-1; i++ {
		for j := i + 1; j < len(files); j++ {
			if files[j].Size > files[i].Size {
				files[i], files[j] = files[j], files[i]
			}
		}
	}
}
