package module

import "sync"

// Cache keeps fetched modules, requirements, and baselines so repeated
// access does not hit the DXL server again. Modules are cached by path/ID,
// requirements by the module-scoped "moduleID/requirementID" composite.
type Cache struct {
	mu           sync.RWMutex
	modules      map[string]*Module
	requirements map[string]*Requirement
	baselines    map[string][]Baseline
}

func NewCache() *Cache {
	return &Cache{
		modules:      map[string]*Module{},
		requirements: map[string]*Requirement{},
		baselines:    map[string][]Baseline{},
	}
}

func (c *Cache) PutModule(m *Module) {
	c.mu.Lock()
	defer c.mu.Unlock()
	key := m.Path
	if key == "" {
		key = m.ID
	}
	c.modules[key] = m
}
func (c *Cache) GetModule(id string) *Module {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.modules[id]
}

func requirementKey(moduleID, reqID string) string { return moduleID + "/" + reqID }

func (c *Cache) PutRequirement(r *Requirement) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.requirements[requirementKey(r.ModuleID, r.ID)] = r
}
func (c *Cache) GetRequirement(moduleID, reqID string) *Requirement {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.requirements[requirementKey(moduleID, reqID)]
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

// InvalidateModule drops a module and its derived entries.
func (c *Cache) InvalidateModule(id string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.modules, id)
	delete(c.baselines, id)
	for rid, r := range c.requirements {
		if r.ModuleID == id {
			delete(c.requirements, rid)
		}
	}
}

// Clear drops everything.
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.modules = map[string]*Module{}
	c.requirements = map[string]*Requirement{}
	c.baselines = map[string][]Baseline{}
}
