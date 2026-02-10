package main

import (
	"context"

	"clearc/internal/analyzer"
	"clearc/internal/cleaner"
	"clearc/internal/config"
	"clearc/internal/platform"
	"clearc/internal/scanner"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx      context.Context
	scanner  *scanner.Scanner
	cleaner  *cleaner.Cleaner
	config   *config.Config
	analyzer *analyzer.Analyzer
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
}

// shutdown is called when the app is closing
func (a *App) shutdown(ctx context.Context) {
	a.config.Save()
	a.analyzer.SaveCache()
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
	a.config = cfg
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
