package module

// Module mirrors the JSON object returned by get_module / list_modules scripts.
type Module struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Path          string `json:"path"`
	CreatedOn     string `json:"createdOn,omitempty"`
	ModifiedOn    string `json:"modifiedOn,omitempty"`
	BaselineCount int    `json:"baselineCount,omitempty"`
}

// Requirement mirrors one requirement row in a module.
type Requirement struct {
	ID         string            `json:"id"`
	ModuleID   string            `json:"moduleId"`
	Heading    string            `json:"heading"`
	Text       string            `json:"text"`
	CreatedBy  string            `json:"createdBy,omitempty"`
	ModifiedBy string            `json:"modifiedBy,omitempty"`
	Attributes map[string]string `json:"attributes,omitempty"`
}

// Baseline mirrors a module baseline.
type Baseline struct {
	ModuleID  string `json:"moduleId"`
	Name      string `json:"name"`
	CreatedOn string `json:"createdOn,omitempty"`
	Comment   string `json:"comment,omitempty"`
}

// Table mirrors an embedded DOORS table. Ownership is module-scoped:
// the same requirement ID can exist in different modules, so both keys are needed.
type Table struct {
	ID            string   `json:"id"`
	ModuleID      string   `json:"moduleId"`
	RequirementID string   `json:"requirementId"`
	Title         string   `json:"title,omitempty"`
	Columns       []string `json:"columns,omitempty"`
	Rows          int      `json:"rows,omitempty"`
}

// HierarchyEntry is one flat node in a module's requirement outline, linked
// to its parent. Flat lists stream well and are easy for AI agents to consume;
// the tree can be reconstructed with "level"/"parentId" when needed.
type HierarchyEntry struct {
	ID       string `json:"id"`
	ParentID string `json:"parentId,omitempty"`
	Level    int    `json:"level"`
	Heading  string `json:"heading,omitempty"`
	Text     string `json:"text,omitempty"`
}
