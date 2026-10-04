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

// Table mirrors an embedded DOORS table.
type Table struct {
	ID      string   `json:"id"`
	Title   string   `json:"title,omitempty"`
	Columns []string `json:"columns,omitempty"`
	Rows    int      `json:"rows,omitempty"`
}

// HierarchyNode is one node in a module/requirement tree.
type HierarchyNode struct {
	ID       string          `json:"id"`
	Label    string          `json:"label,omitempty"`
	Children []HierarchyNode `json:"children,omitempty"`
}
