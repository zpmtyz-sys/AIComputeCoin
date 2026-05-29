package reporter

import (
	"encoding/json"
	"os"
	"sync"
)

// Cache provides local persistence for crash recovery of pending reports.
type Cache struct {
	mu      sync.Mutex
	path    string
	pending map[string]*OnChainReport
}

// NewCache creates a new Cache with the specified file path.
func NewCache(path string) *Cache {
	c := &Cache{
		path:    path,
		pending: make(map[string]*OnChainReport),
	}
	c.load()
	return c
}

// Store persists a report to the cache.
func (c *Cache) Store(report *OnChainReport) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.pending[report.NodeID] = report
	return c.persist()
}

// Remove removes a report from the cache by node ID.
func (c *Cache) Remove(nodeID string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.pending, nodeID)
	_ = c.persist()
}

// GetPending returns all pending reports from the cache.
func (c *Cache) GetPending() []*OnChainReport {
	c.mu.Lock()
	defer c.mu.Unlock()

	reports := make([]*OnChainReport, 0, len(c.pending))
	for _, r := range c.pending {
		reports = append(reports, r)
	}
	return reports
}

// Count returns the number of pending reports.
func (c *Cache) Count() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.pending)
}

// persist writes the cache to disk.
func (c *Cache) persist() error {
	if c.path == "" {
		return nil
	}

	data, err := json.Marshal(c.pending)
	if err != nil {
		return err
	}

	return os.WriteFile(c.path, data, 0600)
}

// load reads the cache from disk.
func (c *Cache) load() {
	if c.path == "" {
		return
	}

	data, err := os.ReadFile(c.path)
	if err != nil {
		return
	}

	_ = json.Unmarshal(data, &c.pending)
}
