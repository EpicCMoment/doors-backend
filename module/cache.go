package module

import "sync"

// Cache keeps fetched modules, requirements, and baselines so repeated
// access does not hit the DXL server again. Cache entries are keyed per
// baseline view: a module/requirement cached at baseline B1 is distinct
// from the same object viewed at B2 (or current). The subItems map caches
// the direct sub-items of each parent item, keyed by that parent path, keyed by that parent path.
type Cache struct {
	mu           sync.RWMutex
	modules      map[string]*Item
	requirements map[string]*Requirement
	baselines    map[string][]Baseline
	// Children of an Item, keyed by that parent's path (from list_items).
	subItems    map[string][]Item
}

func NewCache() *Cache {
	return &Cache{
		modules:      map[string]*Item{},
		requirements: map[string]*Requirement{},
		baselines:    map[string][]Baseline{},
		subItems:    map[string][]Item{},
	}
}

func moduleKey(path, baseline string) string {
	if baseline == "" {
		return path + "@current"
	}
	return path + "@" + baseline
}

func requirementKey(moduleID, reqID, baseline string) string {
	if baseline == "" {
		return moduleID + "/" + reqID + "@current"
	}
	return moduleID + "/" + reqID + "@" + baseline
}

func (c *Cache) PutModule(m *Item) {
	c.mu.Lock()
	defer c.mu.Unlock()
	key := m.Path
	if key == "" {
		key = m.ID
	}
	c.modules[moduleKey(key, m.Baseline)] = m
}
func (c *Cache) GetModule(path, baseline string) *Item {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.modules[moduleKey(path, baseline)]
}

func (c *Cache) PutRequirement(r *Requirement) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.requirements[requirementKey(r.ModulePath, r.ID, r.Baseline)] = r
}
func (c *Cache) GetRequirement(moduleID, reqID, baseline string) *Requirement {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.requirements[requirementKey(moduleID, reqID, baseline)]
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

func (c *Cache) PutSubItems(path string, items []Item) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.subItems[path] = items
}
func (c *Cache) GetSubItems(path string) []Item {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.subItems[path]
}

// Clear drops everything.
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.modules = map[string]*Item{}
	c.requirements = map[string]*Requirement{}
	c.baselines = map[string][]Baseline{}
	c.subItems = map[string][]Item{}
}
