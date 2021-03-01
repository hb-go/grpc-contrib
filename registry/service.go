package registry

type Service struct {
	Name     string            `json:"name"`
	Version  string            `json:"version"`
	Metadata map[string]string `json:"metadata"`
	Methods  []*Method         `json:"methods"`
	Nodes    []*Node           `json:"nodes"`
}

type Node struct {
	Id       string            `json:"id"`
	Address  string            `json:"address"`
	Metadata map[string]string `json:"metadata"`
}

type Method struct {
	Name     string     `json:"name"`
	Bindings []*Binding `json:"bindings"`
}

type Binding struct {
	Method          string    `json:"method"`
	PathTmpl        *PathTmpl `json:"path_tmpl"`
	AssumeColonVerb bool      `json:"assume_colon_verb"`
}

// PathTmpl is a compiled representation of path templates.
type PathTmpl struct {
	// Version is the version number of the format.
	Version int
	// OpCodes is a sequence of operations.
	OpCodes []int
	// Pool is a constant pool
	Pool []string
	// Verb is a VERB part in the template.
	Verb string
	// Fields is a list of field paths bound in this template.
	Fields []string
	// Original template (example: /v1/a_bit_of_everything)
	Template string
}
