package registry

type Service struct {
	Name    string
	Version string
	Methods []*Method
}

type Method struct {
	Name   string
	Routes []*Route
}

type Route struct {
	Method  string
	Pattern *Pattern
}

type Pattern struct {
	Version         int
	Ops             []int
	Pool            []string
	Verb            string
	AssumeColonVerb bool
}
