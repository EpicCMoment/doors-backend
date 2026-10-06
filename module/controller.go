package module

import (
	"encoding/json"
	"fmt"

	"github.com/EpicCMoment/doors-backend/dxl"
)

// ExecClient is the transport ModuleController needs; *dxl.DxlController satisfies it.
type ExecClient interface {
	ExecTemplate(name string, params map[string]dxl.DxlLiteral) ([]byte, error)
}

// ModuleController manages Module/Requirement/Baseline/Table data fetched through
// the DXL server, with an in-memory cache.
type ModuleController struct {
	dxl   ExecClient
	cache *Cache
}

func NewModuleController(dxlClient ExecClient) *ModuleController {
	return &ModuleController{dxl: dxlClient, cache: NewCache()}
}

// Cache exposes the underlying cache for inspection/invalidation.
func (c *ModuleController) Cache() *Cache { return c.cache }

// decodeEnvelope parses the JSON reply and validates the status field.
func decodeEnvelope(raw []byte, dst any) error {
	var env dxl.Envelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return fmt.Errorf("unmarshal envelope: %w", err)
	}
	if env.Status != "ok" {
		if env.Message == "" {
			env.Message = "unknown error"
		}
		return fmt.Errorf("dxl error: %s", env.Message)
	}
	if dst == nil || len(env.Data) == 0 {
		return nil
	}
	if err := json.Unmarshal(env.Data, dst); err != nil {
		return fmt.Errorf("unmarshal data: %w", err)
	}
	return nil
}

// GetModule returns a module's current view (baseline == "") or a baseline view.
func (c *ModuleController) GetModule(path, baseline string) (*Item, error) {
	if m := c.cache.GetModule(path, baseline); m != nil {
		return m, nil
	}
	raw, err := c.dxl.ExecTemplate("get_module", map[string]dxl.DxlLiteral{
		"path":       dxl.String(path),
		"baseline": dxl.String(baseline),
	})
	if err != nil {
		return nil, err
	}
	var it Item
	if err := decodeEnvelope(raw, &it); err != nil {
		return nil, err
	}
	if it.Baseline == "" {
		it.Baseline = baseline
	}
	if it.Type == "" {
		it.Type = ItemModule
	}
	c.cache.PutModule(&it)
	return &it, nil
}

// ListItems returns the database items (projects, folders, modules) directly
// under the given path, from cache if present.
func (c *ModuleController) ListItems(path string) ([]Item, error) {
	if items := c.cache.GetChildren(path); items != nil {
		return items, nil
	}
	raw, err := c.dxl.ExecTemplate("list_items", map[string]dxl.DxlLiteral{"path": dxl.String(path)})
	if err != nil {
		return nil, err
	}
	var items []Item
	if err := decodeEnvelope(raw, &items); err != nil {
		return nil, err
	}
	c.cache.PutChildren(path, items)
	return items, nil
}

// ListModules returns the modules directly under the given folder/project path.
// It is a convenience wrapper around ListItems.
func (c *ModuleController) ListModules(path string) ([]Item, error) {
	items, err := c.ListItems(path)
	if err != nil {
		return nil, err
	}
	var mods []Item
	for _, it := range items {
		if it.Type != ItemModule {
			continue
		}
		c.cache.PutModule(&it)
		mods = append(mods, it)
	}
	return mods, nil
}

// GetRequirements fetches all requirements of a module and caches them.
func (c *ModuleController) GetRequirements(moduleID, baseline string) ([]Requirement, error) {
	raw, err := c.dxl.ExecTemplate("get_requirements", map[string]dxl.DxlLiteral{
		"modulePath":   dxl.String(moduleID),
		"baseline": dxl.String(baseline),
	})
	if err != nil {
		return nil, err
	}
	var reqs []Requirement
	if err := decodeEnvelope(raw, &reqs); err != nil {
		return nil, err
	}
	for i := range reqs {
		if reqs[i].Baseline == "" {
			reqs[i].Baseline = baseline
		}
		c.cache.PutRequirement(&reqs[i])
	}
	return reqs, nil
}

// GetBaselines lists baselines of a module (cached).
func (c *ModuleController) GetBaselines(moduleID string) ([]Baseline, error) {
	if bs := c.cache.GetBaselines(moduleID); bs != nil {
		return bs, nil
	}
	raw, err := c.dxl.ExecTemplate("get_baselines", map[string]dxl.DxlLiteral{"modulePath": dxl.String(moduleID)})
	if err != nil {
		return nil, err
	}
	var bs []Baseline
	if err := decodeEnvelope(raw, &bs); err != nil {
		return nil, err
	}
	c.cache.PutBaselines(moduleID, bs)
	return bs, nil
}

// TraverseHierarchy returns the flat requirement outline of a module, in
// document order, each entry linked to its parent via ParentID/Level.
func (c *ModuleController) TraverseHierarchy(moduleID, baseline string) ([]HierarchyEntry, error) {
	raw, err := c.dxl.ExecTemplate("traverse_hierarchy", map[string]dxl.DxlLiteral{
		"modulePath":   dxl.String(moduleID),
		"baseline": dxl.String(baseline),
	})
	if err != nil {
		return nil, err
	}
	var entries []HierarchyEntry
	if err := decodeEnvelope(raw, &entries); err != nil {
		return nil, err
	}
	for i := range entries {
		if entries[i].Baseline == "" {
			entries[i].Baseline = baseline
		}
	}
	return entries, nil
}

// InvalidateModule forces a refresh of cached data for the given module path.
func (c *ModuleController) InvalidateModule(path string) { c.cache.InvalidateModule(path) }
