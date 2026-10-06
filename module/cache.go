package module

import "sync"

// Cache keeps fetched modules, requirements, and baselines so repeated
// access does not hit the DXL server again. Cache entries are keyed per
// baseline view: a module/requirement cached at baseline B1 is distinct
// from the same object viewed at B2 (or current).
type Cache struct {
	mu           sync.RWMutex
	modules      map[string]*Item
	requirements map[string]*Requirement
	baselines    map[string][]Baseline
	items        map[string][]Item
}

func NewCache() *Cache {
	return &Cache{
		modules:      map[string]*Item{},
		requirements: map[string]*Requirement{},
		baselines:    map[string][]Baseline{},
		items:        map[string][]Item{},
	}
}

func moduleKey(path, baselineID string) string {
	if baselineID == "" {
		return path + "@current"
	}
	return path + "@" + baselineID
}

func requirementKey(moduleID, reqID, baselineID string) string {
	if baselineID == "" {
		return moduleID + "/" + reqID + "@current"
	}
	return moduleID + "/" + reqID + "@" + baselineID
}

func (c *Cache) PutModule(m *Item) {
	c.mu.Lock()
	defer c.mu.Unlock()
	key := m.Path
	if key == "" {
		key = m.ID
	}
	c.modules[moduleKey(key, m.BaselineID)] = m
}
func (c *Cache) GetModule(path, baselineID string) *Item {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.modules[moduleKey(path, baselineID)]
}

func (c *Cache) PutRequirement(r *Requirement) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.requirements[requirementKey(r.ModulePath, r.ID, r.BaselineID)] = r
}
func (c *Cache) GetRequirement(moduleID, reqID, baselineID string) *Requirement {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.requirements[requirementKey(moduleID, reqID, baselineID)]
}

func (c *Cache) PutBaselines(moduleID string, bs []Baseline) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.baselines[moduleID] = bs
}
func (c *Cache) GetBaselines(moduleID string) []Baseline {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.baselines[moduleID]
}

// InvalidateModule drops cached entries for one module (all baselines, and its
// baselines list + requirements). Use InvalidateModuleBaseline for finer control.
func (c *Cache) InvalidateModule(moduleRef string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for k := range c.modules {
		if len(k) >= len(moduleRef) && k[:len(moduleRef)] == moduleRef {
			delete(c.modules, k)
		}
	}
	delete(c.baselines, moduleRef)
	for rid, r := range c.requirements {
		if r.ModulePath == moduleRef {
			delete(c.requirements, rid)
		}
	}
}

func (c *Cache) PutItems(path string, items []Item) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[path] = items
}
func (c *Cache) GetItems(path string) []Item {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.items[path]
}

// Clear drops everything.
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.modules = map[string]*Item{}
	c.requirements = map[string]*Requirement{}
	c.baselines = map[string][]Baseline{}
	c.items = map[string][]Item{}
}
