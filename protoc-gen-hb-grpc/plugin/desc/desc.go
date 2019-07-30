package desc

import (
	pb "github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/hb-go/grpc-contrib/protoc-gen-hb-grpc/generator"
)

func init() {
	generator.RegisterPlugin(new(exportDesc))
}

type exportDesc struct {
	gen *generator.Generator
}

// Name returns the name of this plugin, "desc".
func (g *exportDesc) Name() string {
	return "desc"
}

// The names for packages imported in the generated code.
// They may vary from the final path component of the import path
// if the name is used by other packages.
var (
	pkgImports map[generator.GoPackageName]bool
)

// Init initializes the plugin.
func (g *exportDesc) Init(gen *generator.Generator) {
	g.gen = gen
}

// P forwards to g.gen.P.
func (g *exportDesc) P(args ...interface{}) { g.gen.P(args...) }

// Generate generates code for the services in the given file.
func (g *exportDesc) Generate(file *generator.FileDescriptor) {
	if len(file.FileDescriptorProto.Service) == 0 {
		return
	}

	g.P()
	g.P("// Export service desc")

	for i, service := range file.FileDescriptorProto.Service {
		g.generateService(file, service, i)
	}
}

func (g *exportDesc) GenerateImports(file *generator.FileDescriptor) {

}

func (g *exportDesc) generateService(file *generator.FileDescriptor, service *pb.ServiceDescriptorProto, index int) {
	g.P()
	g.P("// " + service.GetName() + " desc")
	g.P("func ServiceDesc" + service.GetName() + "() *grpc.ServiceDesc {")
	g.P("return &_" + service.GetName() + "_serviceDesc")
	g.P("}")
}
