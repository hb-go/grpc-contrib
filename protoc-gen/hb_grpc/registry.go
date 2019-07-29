package hb_grpc

import (
	"path"
	"strconv"

	pb "github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/hb-go/grpc-contrib/protoc-gen/generator"
)

const (
	registryPkgPath = "github.com/hb-go/grpc-contrib/registry"
)

func init() {
	generator.RegisterPlugin(new(grpcRegistry))
}

type grpcRegistry struct {
	gen *generator.Generator
}

// Name returns the name of this plugin, "hb-grpc".
func (g *grpcRegistry) Name() string {
	return "registry"
}

// The names for packages imported in the generated code.
// They may vary from the final path component of the import path
// if the name is used by other packages.
var (
	registryPkg string
	pkgImports  map[generator.GoPackageName]bool
)

// Init initializes the plugin.
func (g *grpcRegistry) Init(gen *generator.Generator) {
	g.gen = gen
	registryPkg = generator.RegisterUniquePackageName("registry", nil)
}

// Given a type name defined in a .proto, return its object.
// Also record that we're using it, to guarantee the associated import.
func (g *grpcRegistry) objectNamed(name string) generator.Object {
	g.gen.RecordTypeUse(name)
	return g.gen.ObjectNamed(name)
}

// Given a type name defined in a .proto, return its name as we will print it.
func (g *grpcRegistry) typeName(str string) string {
	return g.gen.TypeName(g.objectNamed(str))
}

// P forwards to g.gen.P.
func (g *grpcRegistry) P(args ...interface{}) { g.gen.P(args...) }

// Generate generates code for the services in the given file.
func (g *grpcRegistry) Generate(file *generator.FileDescriptor) {
	if len(file.FileDescriptorProto.Service) == 0 {
		return
	}
	g.P("// gRPC Registry")
	g.P("// github.com/hb-go/grpc-contrib/registry")
	g.P()
	g.P("// Reference imports to suppress errors if they are not otherwise used.")
	g.P("var _ ", registryPkg, ".Registry")
	g.P()

	for i, service := range file.FileDescriptorProto.Service {
		g.generateService(file, service, i)
	}
}

func (g *grpcRegistry) GenerateImports(file *generator.FileDescriptor) {
	if len(file.FileDescriptorProto.Service) == 0 {
		return
	}
	g.P("import (")
	g.P(registryPkg, " ", strconv.Quote(path.Join(g.gen.ImportPrefix, registryPkgPath)))
	g.P(")")
	g.P()
}

func (g *grpcRegistry) generateService(file *generator.FileDescriptor, service *pb.ServiceDescriptorProto, index int) {
	g.P()
	g.P("// " + service.GetName() + " registry")
	g.P("func Target" + service.GetName() + "(opts ...registry.Option) string {")
	g.P("return registry.NewTarget(&_" + service.GetName() + "_serviceDesc, opts...)")
	g.P("}")
	g.P()

	g.P("func Register" + service.GetName() + "(opts ...registry.Option) error {")
	g.P("return registry.Register(&_" + service.GetName() + "_serviceDesc, opts...)")
	g.P("}")
	g.P()
	g.P("func Deregister" + service.GetName() + "(opts ...registry.Option) {")
	g.P("registry.Deregister(&_" + service.GetName() + "_serviceDesc, opts...)")
	g.P("}")
	g.P()

}
