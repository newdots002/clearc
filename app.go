package main

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"clearc/internal/analyzer"
	"clearc/internal/cleaner"
	"clearc/internal/config"
	"clearc/internal/platform"
	"clearc/internal/scanner"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// 激活码验证服务器地址（请修改为实际地址）
const ActivationServerURL = "https://clearc.top/check.php"

// App struct
type App struct {
	ctx                context.Context
	scanner            *scanner.Scanner
	cleaner            *cleaner.Cleaner
	config             *config.Config
	analyzer           *analyzer.Analyzer
	autoAnalyzeStop    chan struct{}
	autoAnalyzeMutex   sync.Mutex
	lastAnalyzeTime    time.Time
	isAutoAnalyzing    bool
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		scanner:  scanner.New(),
		cleaner:  cleaner.New(),
		config:   config.New(),
		analyzer: analyzer.New(),
	}
}

// startup is called when the app starts
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.config.Load()
	
	// 启动自动分析（如果已开启）
	if a.config.AutoAnalyze {
		a.startAutoAnalyze()
	}
}

// shutdown is called when the app is closing
func (a *App) shutdown(ctx context.Context) {
	a.stopAutoAnalyze()
	a.config.Save()
	a.analyzer.SaveCache()
}

// startAutoAnalyze 启动后台自动分析
func (a *App) startAutoAnalyze() {
	a.autoAnalyzeMutex.Lock()
	defer a.autoAnalyzeMutex.Unlock()
	
	if a.autoAnalyzeStop != nil {
		return // 已经在运行
	}
	
	a.autoAnalyzeStop = make(chan struct{})
	
	go func() {
		// 首次启动延迟5分钟再开始
		initialDelay := time.NewTimer(5 * time.Minute)
		select {
		case <-initialDelay.C:
		case <-a.autoAnalyzeStop:
			initialDelay.Stop()
			return
		}
		
		for {
			// 检查距离上次分析的时间
			interval := time.Duration(a.config.AutoAnalyzeInterval) * time.Minute
			if interval < 30*time.Minute {
				interval = 30 * time.Minute // 最小30分钟
			}
			
			timeSinceLastAnalyze := time.Since(a.lastAnalyzeTime)
			if timeSinceLastAnalyze >= interval {
				a.runBackgroundAnalyze()
			}
			
			// 每10分钟检查一次
			checkTimer := time.NewTimer(10 * time.Minute)
			select {
			case <-checkTimer.C:
				continue
			case <-a.autoAnalyzeStop:
				checkTimer.Stop()
				return
			}
		}
	}()
}

// stopAutoAnalyze 停止后台自动分析
func (a *App) stopAutoAnalyze() {
	a.autoAnalyzeMutex.Lock()
	defer a.autoAnalyzeMutex.Unlock()
	
	if a.autoAnalyzeStop != nil {
		close(a.autoAnalyzeStop)
		a.autoAnalyzeStop = nil
	}
}

// runBackgroundAnalyze 执行后台分析
func (a *App) runBackgroundAnalyze() {
	if a.isAutoAnalyzing {
		return
	}
	
	a.isAutoAnalyzing = true
	defer func() {
		a.isAutoAnalyzing = false
		a.lastAnalyzeTime = time.Now()
	}()
	
	// 静默执行快速扫描，只更新缓存
	result, err := a.analyzer.QuickScan()
	if err != nil {
		return
	}
	
	// 计算大小并更新缓存（静默，不发送事件到前端）
	a.analyzer.CalculateSizesAsync(result.Children, false, nil)
	
	// 通知前端缓存已更新（可选）
	if a.ctx != nil {
		runtime.EventsEmit(a.ctx, "backgroundAnalyzeComplete", map[string]interface{}{
			"time": time.Now().Unix(),
		})
	}
}

// GetLastAnalyzeTime 获取上次分析时间
func (a *App) GetLastAnalyzeTime() int64 {
	return a.lastAnalyzeTime.Unix()
}

// IsAutoAnalyzing 检查是否正在自动分析
func (a *App) IsAutoAnalyzing() bool {
	return a.isAutoAnalyzing
}

// DiskUsage represents disk usage information
type DiskUsage struct {
	Total       uint64  `json:"total"`
	Used        uint64  `json:"used"`
	Free        uint64  `json:"free"`
	UsedPercent float64 `json:"usedPercent"`
	Path        string  `json:"path"`
}

// GetDiskUsage returns disk usage information
func (a *App) GetDiskUsage() (*DiskUsage, error) {
	usage, err := platform.GetDiskUsage(platform.GetSystemDrive())
	if err != nil {
		return nil, err
	}
	return &DiskUsage{
		Total:       usage.Total,
		Used:        usage.Used,
		Free:        usage.Free,
		UsedPercent: usage.UsedPercent,
		Path:        platform.GetSystemDrive(),
	}, nil
}

// ScanResult represents the result of a scan
type ScanResult struct {
	Categories []CategoryResult `json:"categories"`
	TotalSize  int64            `json:"totalSize"`
	TotalFiles int              `json:"totalFiles"`
}

// CategoryResult represents a category of junk files
type CategoryResult struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Size        int64  `json:"size"`
	FileCount   int    `json:"fileCount"`
	Color       string `json:"color"`
	Selected    bool   `json:"selected"`
}

// ScanForJunk scans for junk files
func (a *App) ScanForJunk() (*ScanResult, error) {
	result, err := a.scanner.ScanAll()
	if err != nil {
		return nil, err
	}
	categories := make([]CategoryResult, len(result.Categories))
	for i, c := range result.Categories {
		categories[i] = CategoryResult{
			ID:          c.ID,
			Name:        c.Name,
			Description: c.Description,
			Size:        c.Size,
			FileCount:   c.FileCount,
			Color:       c.Color,
			Selected:    c.Selected,
		}
	}
	return &ScanResult{
		Categories: categories,
		TotalSize:  result.TotalSize,
		TotalFiles: result.TotalFiles,
	}, nil
}

// CleanResult represents the result of cleaning
type CleanResult struct {
	CleanedSize  int64    `json:"cleanedSize"`
	CleanedFiles int      `json:"cleanedFiles"`
	Errors       []string `json:"errors"`
}

// CleanCategories cleans selected categories
func (a *App) CleanCategories(categoryIDs []string) (*CleanResult, error) {
	result, err := a.cleaner.CleanCategories(categoryIDs)
	if err != nil {
		return nil, err
	}
	return &CleanResult{
		CleanedSize:  result.CleanedSize,
		CleanedFiles: result.CleanedFiles,
		Errors:       result.Errors,
	}, nil
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
func (a *App) GetUnusedFiles(days int, minSizeMB int64) ([]UnusedFile, error) {
	files, err := a.scanner.GetUnusedFiles(days, minSizeMB*1024*1024)
	if err != nil {
		return nil, err
	}
	result := make([]UnusedFile, len(files))
	for i, f := range files {
		result[i] = UnusedFile{
			Path:         f.Path,
			Name:         f.Name,
			Size:         f.Size,
			LastAccessed: f.LastAccessed,
			DaysUnused:   f.DaysUnused,
			FileType:     f.FileType,
		}
	}
	return result, nil
}

// DeleteFiles deletes specified files
func (a *App) DeleteFiles(paths []string) (*CleanResult, error) {
	result, err := a.cleaner.DeleteFiles(paths)
	if err != nil {
		return nil, err
	}
	return &CleanResult{
		CleanedSize:  result.CleanedSize,
		CleanedFiles: result.CleanedFiles,
		Errors:       result.Errors,
	}, nil
}

// GetConfig returns the current configuration
func (a *App) GetConfig() *config.Config {
	return a.config
}

// SaveConfig saves the configuration
func (a *App) SaveConfig(cfg *config.Config) error {
	oldAutoAnalyze := a.config.AutoAnalyze
	a.config = cfg
	
	// 处理自动分析设置变化
	if cfg.AutoAnalyze && !oldAutoAnalyze {
		a.startAutoAnalyze()
	} else if !cfg.AutoAnalyze && oldAutoAnalyze {
		a.stopAutoAnalyze()
	}
	
	return a.config.Save()
}

// GetScanProgress returns the current scan progress (0-100)
func (a *App) GetScanProgress() int {
	return a.scanner.GetProgress()
}

// GetScanStatus returns the current scan status message
func (a *App) GetScanStatus() string {
	return a.scanner.GetStatus()
}

// FileNode represents a file or directory in the tree
type FileNode struct {
	Name           string      `json:"name"`
	Path           string      `json:"path"`
	Size           int64       `json:"size"`
	IsDir          bool        `json:"isDir"`
	Children       []*FileNode `json:"children,omitempty"`
	IsProtected    bool        `json:"isProtected"`
	FileCount      int         `json:"fileCount"`
	DirType        string      `json:"dirType"`
	Recommendation string      `json:"recommendation"`
	Description    string      `json:"description"`
}

// AnalyzeDrive analyzes a drive and returns a tree structure (fast - sizes calculated async)
func (a *App) AnalyzeDrive(drivePath string, skipSystem bool) (*FileNode, error) {
	result, err := a.analyzer.AnalyzeDrive(drivePath)
	if err != nil {
		return nil, err
	}

	// Start async size calculation
	go func() {
		a.analyzer.CalculateSizesAsync(result.Children, skipSystem, func(path string, size int64, fileCount int) {
			// Emit event to frontend with size update
			runtime.EventsEmit(a.ctx, "sizeUpdate", map[string]interface{}{
				"path":      path,
				"size":      size,
				"fileCount": fileCount,
			})
		})
		// Emit completion event
		runtime.EventsEmit(a.ctx, "sizeCalculationComplete", nil)
	}()

	return convertFileNode(result), nil
}

// AnalyzeQuickScan performs a quick scan of important directories (Users, caches, etc.)
func (a *App) AnalyzeQuickScan() (*FileNode, error) {
	result, err := a.analyzer.QuickScan()
	if err != nil {
		return nil, err
	}

	// Start async size calculation (quick scan never has system dirs)
	go func() {
		a.analyzer.CalculateSizesAsync(result.Children, false, func(path string, size int64, fileCount int) {
			runtime.EventsEmit(a.ctx, "sizeUpdate", map[string]interface{}{
				"path":      path,
				"size":      size,
				"fileCount": fileCount,
			})
		})
		runtime.EventsEmit(a.ctx, "sizeCalculationComplete", nil)
	}()

	return convertFileNode(result), nil
}

// ExpandNode expands a directory node to show its children (fast - sizes calculated async)
func (a *App) ExpandNode(path string) (*FileNode, error) {
	result, err := a.analyzer.ExpandNode(path)
	if err != nil {
		return nil, err
	}

	// Start async size calculation for children
	if result != nil && len(result.Children) > 0 {
		go func() {
			a.analyzer.CalculateNodeSizesAsync(result.Children, func(p string, size int64, fileCount int) {
				runtime.EventsEmit(a.ctx, "sizeUpdate", map[string]interface{}{
					"path":      p,
					"size":      size,
					"fileCount": fileCount,
				})
			})
			// Emit completion for this expansion
			runtime.EventsEmit(a.ctx, "expandComplete", map[string]interface{}{
				"path": path,
			})
		}()
	}

	return convertFileNode(result), nil
}

// GetAnalyzeProgress returns the current analyze progress
func (a *App) GetAnalyzeProgress() int {
	return a.analyzer.GetProgress()
}

// GetAnalyzeStatus returns the current analyze status
func (a *App) GetAnalyzeStatus() string {
	return a.analyzer.GetStatus()
}

// DeletePaths deletes the specified paths
func (a *App) DeletePaths(paths []string) (*CleanResult, error) {
	size, count, errors := a.analyzer.DeletePaths(paths)
	return &CleanResult{
		CleanedSize:  size,
		CleanedFiles: count,
		Errors:       errors,
	}, nil
}

// GetWhitelistDirs returns the list of protected directories
func (a *App) GetWhitelistDirs() []string {
	return a.analyzer.GetWhitelistDirs()
}

// SetAnalyzerMinSize sets the minimum file/folder size to display (in MB)
func (a *App) SetAnalyzerMinSize(sizeMB int64) {
	a.analyzer.SetMinSize(sizeMB * 1024 * 1024)
}

// GetAnalyzerMinSize returns the current minimum size filter (in MB)
func (a *App) GetAnalyzerMinSize() int64 {
	return a.analyzer.GetMinSize() / (1024 * 1024)
}

// ClearAnalyzerCache clears the directory size cache
func (a *App) ClearAnalyzerCache() {
	a.analyzer.ClearCache()
}

// GetCacheStats returns cache statistics
func (a *App) GetCacheStats() map[string]int {
	total, valid := a.analyzer.GetCacheStats()
	return map[string]int{
		"total": total,
		"valid": valid,
	}
}

// convertFileNode converts analyzer.FileNode to main.FileNode
func convertFileNode(node *analyzer.FileNode) *FileNode {
	if node == nil {
		return nil
	}
	result := &FileNode{
		Name:           node.Name,
		Path:           node.Path,
		Size:           node.Size,
		IsDir:          node.IsDir,
		IsProtected:    node.IsProtected,
		FileCount:      node.FileCount,
		DirType:        string(node.DirType),
		Recommendation: string(node.Recommendation),
		Description:    node.Description,
	}
	if len(node.Children) > 0 {
		result.Children = make([]*FileNode, len(node.Children))
		for i, child := range node.Children {
			result.Children[i] = convertFileNode(child)
		}
	}
	return result
}

// VIPStatus represents the VIP status information
type VIPStatus struct {
	IsVIP          bool  `json:"isVip"`
	IsTrialExpired bool  `json:"isTrialExpired"`
	TrialDaysLeft  int   `json:"trialDaysLeft"`
	TrialDays      int   `json:"trialDays"`
	FirstUseTime   int64 `json:"firstUseTime"`
	VIPActivatedAt int64 `json:"vipActivatedAt"`
}

// GetVIPStatus returns the current VIP status
func (a *App) GetVIPStatus() *VIPStatus {
	now := time.Now().Unix()
	
	// 如果是首次使用，记录时间
	if a.config.FirstUseTime == 0 {
		a.config.FirstUseTime = now
		a.config.Save()
	}
	
	// 计算试用剩余天数
	daysSinceFirstUse := int((now - a.config.FirstUseTime) / 86400)
	trialDaysLeft := a.config.TrialDays - daysSinceFirstUse
	if trialDaysLeft < 0 {
		trialDaysLeft = 0
	}
	
	isTrialExpired := !a.config.IsVIP && daysSinceFirstUse >= a.config.TrialDays
	
	return &VIPStatus{
		IsVIP:          a.config.IsVIP,
		IsTrialExpired: isTrialExpired,
		TrialDaysLeft:  trialDaysLeft,
		TrialDays:      a.config.TrialDays,
		FirstUseTime:   a.config.FirstUseTime,
		VIPActivatedAt: a.config.VIPActivatedAt,
	}
}

// CheckTrialStatus checks if trial has expired (for scan operations)
func (a *App) CheckTrialStatus() bool {
	status := a.GetVIPStatus()
	return status.IsVIP || !status.IsTrialExpired
}

// ActivationResult represents the result of activation
type ActivationResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// ValidateActivationCode validates the format of an activation code
func (a *App) ValidateActivationCode(code string) bool {
	// 激活码格式: XXXX-XXXX-XXXX-XXXX (16位字母数字，用-分隔)
	if len(code) != 19 {
		return false
	}
	
	// 检查格式
	for i, c := range code {
		if i == 4 || i == 9 || i == 14 {
			if c != '-' {
				return false
			}
		} else {
			// 允许大写字母和数字（排除容易混淆的字符 O, I, L, 0, 1）
			if !((c >= 'A' && c <= 'Z' && c != 'O' && c != 'I' && c != 'L') ||
				(c >= '2' && c <= '9')) {
				return false
			}
		}
	}
	return true
}

// ServerVerifyResponse represents the response from activation server
type ServerVerifyResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	Valid     bool   `json:"valid"`
	Bound     bool   `json:"bound"`
	Activated bool   `json:"activated"`
	ErrorCode string `json:"error_code"`
}

// getDeviceID generates a unique device identifier
func (a *App) getDeviceID() string {
	// 使用机器名和配置首次使用时间生成设备ID
	hostname := platform.GetHostname()
	data := fmt.Sprintf("%s-%d", hostname, a.config.FirstUseTime)
	hash := md5.Sum([]byte(data))
	return hex.EncodeToString(hash[:])
}

// verifyActivationCodeOnline verifies the activation code with the server
func (a *App) verifyActivationCodeOnline(code string, action string) (*ServerVerifyResponse, error) {
	deviceID := a.getDeviceID()
	
	// 构建请求URL
	params := url.Values{}
	params.Set("action", action)
	params.Set("code", code)
	params.Set("device_id", deviceID)
	
	reqURL := ActivationServerURL + "?" + params.Encode()
	
	// 设置超时
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	
	resp, err := client.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("网络请求失败: %v", err)
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}
	
	var result ServerVerifyResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}
	
	return &result, nil
}

// ActivateVIP activates VIP status with an activation code
func (a *App) ActivateVIP(activationCode string) *ActivationResult {
	// 验证激活码格式
	if !a.ValidateActivationCode(activationCode) {
		return &ActivationResult{
			Success: false,
			Message: "激活码格式无效，请检查后重试",
		}
	}
	
	// 检查是否已经是VIP（且使用的是同一个激活码）
	if a.config.IsVIP && a.config.ActivationCode == activationCode {
		return &ActivationResult{
			Success: false,
			Message: "您已经是永久VIP用户",
		}
	}
	
	// 在线验证并激活激活码
	verifyResult, err := a.verifyActivationCodeOnline(activationCode, "activate")
	if err != nil {
		// 网络错误时，尝试使用本地验证（兼容离线场景）
		// 如果之前已经用这个激活码激活过，允许重新激活
		if a.config.ActivationCode == activationCode {
			a.config.IsVIP = true
			a.config.Save()
			return &ActivationResult{
				Success: true,
				Message: "恭喜！您已成功激活永久VIP（离线模式）",
			}
		}
		return &ActivationResult{
			Success: false,
			Message: fmt.Sprintf("验证失败：%v", err),
		}
	}
	
	if !verifyResult.Success {
		return &ActivationResult{
			Success: false,
			Message: verifyResult.Message,
		}
	}
	
	// 激活成功，保存到本地配置
	a.config.IsVIP = true
	a.config.VIPActivatedAt = time.Now().Unix()
	a.config.ActivationCode = activationCode
	
	if err := a.config.Save(); err != nil {
		return &ActivationResult{
			Success: false,
			Message: "保存配置失败，请重试",
		}
	}
	
	return &ActivationResult{
		Success: true,
		Message: "恭喜！您已成功激活永久VIP",
	}
}

// VerifyActivationCode verifies an activation code without activating
func (a *App) VerifyActivationCode(code string) *ActivationResult {
	if !a.ValidateActivationCode(code) {
		return &ActivationResult{
			Success: false,
			Message: "激活码格式无效",
		}
	}
	
	result, err := a.verifyActivationCodeOnline(code, "verify")
	if err != nil {
		return &ActivationResult{
			Success: false,
			Message: fmt.Sprintf("验证失败：%v", err),
		}
	}
	
	return &ActivationResult{
		Success: result.Success,
		Message: result.Message,
	}
}

// isActivationCodeUsed checks if an activation code has been used locally
func (a *App) isActivationCodeUsed(code string) bool {
	return a.config.ActivationCode == code
}

// DeactivateVIP deactivates VIP status (for testing)
func (a *App) DeactivateVIP() error {
	a.config.IsVIP = false
	a.config.VIPActivatedAt = 0
	return a.config.Save()
}

// ResetTrial resets the trial period (for testing)
func (a *App) ResetTrial() error {
	a.config.FirstUseTime = time.Now().Unix()
	a.config.IsVIP = false
	a.config.VIPActivatedAt = 0
	return a.config.Save()
}
