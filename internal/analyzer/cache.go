package analyzer

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// DirCacheEntry represents a cached directory size entry
type DirCacheEntry struct {
	Path         string    `json:"path"`
	Size         int64     `json:"size"`
	FileCount    int       `json:"fileCount"`
	ModTime      time.Time `json:"modTime"`      // Directory's modification time
	ChildCount   int       `json:"childCount"`   // Number of direct children
	ChildModHash int64     `json:"childModHash"` // Hash of children's mod times
	CachedAt     time.Time `json:"cachedAt"`     // When this entry was cached
}

// DirCache manages cached directory sizes
type DirCache struct {
	mu       sync.RWMutex
	entries  map[string]*DirCacheEntry
	filePath string
	dirty    bool // Track if cache needs saving
}

// NewDirCache creates a new directory cache
func NewDirCache() *DirCache {
	cache := &DirCache{
		entries: make(map[string]*DirCacheEntry),
	}

	// Set cache file path
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		cacheDir = os.TempDir()
	}
	cache.filePath = filepath.Join(cacheDir, "clearc", "dir_cache.json")

	// Load existing cache
	cache.Load()

	return cache
}

// Get retrieves a cached entry if it's still valid
func (c *DirCache) Get(path string) (*DirCacheEntry, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.entries[path]
	if !exists {
		return nil, false
	}

	// Check if directory still exists
	info, err := os.Stat(path)
	if err != nil {
		return nil, false
	}

	// Check if cached entry is too old (max 1 year)
	if time.Since(entry.CachedAt) > 365*24*time.Hour {
		return nil, false
	}

	// Quick check: directory's own modification time
	if !info.ModTime().Equal(entry.ModTime) {
		return nil, false
	}

	// Deep check: verify children haven't changed
	childCount, childModHash := c.calculateChildHash(path)
	if childCount != entry.ChildCount || childModHash != entry.ChildModHash {
		return nil, false
	}

	return entry, true
}

// calculateChildHash calculates a hash based on direct children's modification times
func (c *DirCache) calculateChildHash(path string) (int, int64) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return 0, 0
	}

	var hash int64
	count := len(entries)

	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}
		// Combine modification times into a hash
		hash += info.ModTime().UnixNano()
		// Also include size for files to detect content changes
		if !entry.IsDir() {
			hash += info.Size()
		}
	}

	return count, hash
}

// Set stores a directory size in the cache
func (c *DirCache) Set(path string, size int64, fileCount int) {
	info, err := os.Stat(path)
	if err != nil {
		return
	}

	// Calculate children hash for validation
	childCount, childModHash := c.calculateChildHash(path)

	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries[path] = &DirCacheEntry{
		Path:         path,
		Size:         size,
		FileCount:    fileCount,
		ModTime:      info.ModTime(),
		ChildCount:   childCount,
		ChildModHash: childModHash,
		CachedAt:     time.Now(),
	}
	c.dirty = true
}

// Invalidate removes a path and all its children from cache
func (c *DirCache) Invalidate(path string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Remove the path itself
	delete(c.entries, path)

	// Remove all children
	pathPrefix := path + string(os.PathSeparator)
	for key := range c.entries {
		if len(key) > len(pathPrefix) && key[:len(pathPrefix)] == pathPrefix {
			delete(c.entries, key)
		}
	}
	c.dirty = true
}

// Load loads the cache from disk
func (c *DirCache) Load() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	data, err := os.ReadFile(c.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No cache file yet
		}
		return err
	}

	var entries map[string]*DirCacheEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return err
	}

	c.entries = entries
	return nil
}

// Save saves the cache to disk
func (c *DirCache) Save() error {
	c.mu.RLock()
	if !c.dirty {
		c.mu.RUnlock()
		return nil
	}
	entries := c.entries
	c.mu.RUnlock()

	// Ensure directory exists
	dir := filepath.Dir(c.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.Marshal(entries)
	if err != nil {
		return err
	}

	c.mu.Lock()
	c.dirty = false
	c.mu.Unlock()

	return os.WriteFile(c.filePath, data, 0644)
}

// Clear removes all entries from the cache
func (c *DirCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries = make(map[string]*DirCacheEntry)
	c.dirty = true
}

// Stats returns cache statistics
func (c *DirCache) Stats() (total int, valid int) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	total = len(c.entries)
	for path, entry := range c.entries {
		info, err := os.Stat(path)
		if err != nil {
			continue
		}
		if !info.ModTime().Equal(entry.ModTime) {
			continue
		}
		childCount, childModHash := c.calculateChildHash(path)
		if childCount == entry.ChildCount && childModHash == entry.ChildModHash {
			valid++
		}
	}
	return
}
