package registry

import (
	"github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway/httprule"
)

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
	Method          string             `json:"method"`
	PathTmpl        *httprule.Template `json:"path_tmpl"`
	AssumeColonVerb bool               `json:"assume_colon_verb"`
}
