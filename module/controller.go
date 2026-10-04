package module

import (
	"encoding/json"
	"fmt"

	"gitlab.com/ariffil/doors-backend/dxl"
)

// ExecClient is the transport ModuleController needs; *dxl.ModuleController satisfies it.
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

// GetModule returns a module by DOORS full path, from cache if present.
func (c *ModuleController) GetModule(path string) (*Module, error) {
	if m := c.cache.GetModule(path); m != nil {
		return m, nil
	}
	raw, err := c.dxl.ExecTemplate("get_module", map[string]dxl.DxlLiteral{"path": dxl.String(path)})
	if err != nil {
		return nil, err
	}
	var m Module
	if err := decodeEnvelope(raw, &m); err != nil {
		return nil, err
	}
	c.cache.PutModule(&m)
	return &m, nil
}

// ListModules returns the modules directly under the given folder path.
func (c *ModuleController) ListModules(folderPath string) ([]Module, error) {
	raw, err := c.dxl.ExecTemplate("list_modules", map[string]dxl.DxlLiteral{"path": dxl.String(folderPath)})
	if err != nil {
		return nil, err
	}
	var mods []Module
	if err := decodeEnvelope(raw, &mods); err != nil {
		return nil, err
	}
	for i := range mods {
		c.cache.PutModule(&mods[i])
	}
	return mods, nil
}

// GetRequirements fetches all requirements of a module and caches them.
func (c *ModuleController) GetRequirements(moduleID string) ([]Requirement, error) {
	raw, err := c.dxl.ExecTemplate("get_requirements", map[string]dxl.DxlLiteral{"moduleId": dxl.String(moduleID)})
	if err != nil {
		return nil, err
	}
	var reqs []Requirement
	if err := decodeEnvelope(raw, &reqs); err != nil {
		return nil, err
	}
	for i := range reqs {
		c.cache.PutRequirement(&reqs[i])
	}
	return reqs, nil
}

// GetBaselines lists baselines of a module (cached).
func (c *ModuleController) GetBaselines(moduleID string) ([]Baseline, error) {
	if bs := c.cache.GetBaselines(moduleID); bs != nil {
		return bs, nil
	}
	raw, err := c.dxl.ExecTemplate("get_baselines", map[string]dxl.DxlLiteral{"moduleId": dxl.String(moduleID)})
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
func (c *ModuleController) TraverseHierarchy(moduleID string) ([]HierarchyEntry, error) {
	raw, err := c.dxl.ExecTemplate("traverse_hierarchy", map[string]dxl.DxlLiteral{"moduleId": dxl.String(moduleID)})
	if err != nil {
		return nil, err
	}
	var entries []HierarchyEntry
	if err := decodeEnvelope(raw, &entries); err != nil {
		return nil, err
	}
	return entries, nil
}

// InvalidateModule forces a refresh of cached data for the given module path.
func (c *ModuleController) InvalidateModule(path string) { c.cache.InvalidateModule(path) }
