package cleaner

import (
	"os"
	"path/filepath"
	"sync"
)

// Cleaner handles file cleaning operations
type Cleaner struct {
	mu sync.Mutex
}

// New creates a new Cleaner instance
func New() *Cleaner {
	return &Cleaner{}
}

// CleanResult represents the result of a cleaning operation
type CleanResult struct {
	CleanedSize  int64    `json:"cleanedSize"`
	CleanedFiles int      `json:"cleanedFiles"`
	Errors       []string `json:"errors"`
}

// CleanCategories cleans files in specified categories
func (c *Cleaner) CleanCategories(categoryIDs []string) (*CleanResult, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	result := &CleanResult{
		Errors: make([]string, 0),
	}

	// This would be called with actual paths from the scanner
	// For now, return empty result
	return result, nil
}

// DeleteFiles deletes specified files
func (c *Cleaner) DeleteFiles(paths []string) (*CleanResult, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	result := &CleanResult{
		Errors: make([]string, 0),
	}

	for _, path := range paths {
		info, err := os.Stat(path)
		if err != nil {
			result.Errors = append(result.Errors, "无法访问: "+path)
			continue
		}

		size := info.Size()
		if info.IsDir() {
			size = getDirSize(path)
		}

		err = os.RemoveAll(path)
		if err != nil {
			result.Errors = append(result.Errors, "删除失败: "+path+": "+err.Error())
			continue
		}

		result.CleanedSize += size
		result.CleanedFiles++
	}

	return result, nil
}

// CleanDirectory removes all files in a directory
func (c *Cleaner) CleanDirectory(dir string) (*CleanResult, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	result := &CleanResult{
		Errors: make([]string, 0),
	}

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return result, nil
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		result.Errors = append(result.Errors, "无法读取目录: "+dir)
		return result, nil
	}

	for _, entry := range entries {
		path := filepath.Join(dir, entry.Name())
		info, err := entry.Info()
		if err != nil {
			continue
		}

		size := info.Size()
		if info.IsDir() {
			size = getDirSize(path)
		}

		err = os.RemoveAll(path)
		if err != nil {
			result.Errors = append(result.Errors, "删除失败: "+path)
			continue
		}

		result.CleanedSize += size
		result.CleanedFiles++
	}

	return result, nil
}

func getDirSize(path string) int64 {
	var size int64
	filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size
}

// MoveToTrash moves a file to the system trash
func (c *Cleaner) MoveToTrash(path string) error {
	// For now, just delete the file
	// TODO: Implement proper trash functionality per platform
	return os.RemoveAll(path)
}
