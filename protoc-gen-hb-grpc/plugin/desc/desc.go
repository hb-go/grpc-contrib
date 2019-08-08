package desc

import (
	"path"
	"strconv"

	pb "github.com/golang/protobuf/protoc-gen-go/descriptor"

	"github.com/hb-go/grpc-contrib/protoc-gen-hb-grpc/generator"
)

const (
	grpcPkgPath = "google.golang.org/grpc"
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
	grpcPkg string
)

// Init initializes the plugin.
func (g *exportDesc) Init(gen *generator.Generator) {
	g.gen = gen
	grpcPkg = generator.RegisterUniquePackageName("grpc", nil)
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

	if len(file.FileDescriptorProto.Service) > 0 {
		for i, service := range file.FileDescriptorProto.Service {
			g.generateService(file, service, i)
		}
	} else {
		g.P()
		g.P("// Reference imports to suppress errors if they are not otherwise used.")
		g.P("var _ ", grpcPkg, ".Registry")
		g.P()
	}
}

func (g *exportDesc) GenerateImports(file *generator.FileDescriptor) {
	if len(file.FileDescriptorProto.Service) == 0 {
		return
	}
	g.P("import (")
	g.P(grpcPkg, " ", strconv.Quote(path.Join(g.gen.ImportPrefix, grpcPkgPath)))
	g.P(")")
	g.P()
}

func (g *exportDesc) generateService(file *generator.FileDescriptor, service *pb.ServiceDescriptorProto, index int) {
	g.P()
	g.P("// " + service.GetName() + " desc")
	g.P("func ServiceDesc" + service.GetName() + "() grpc.ServiceDesc {")
	g.P("return _" + service.GetName() + "_serviceDesc")
	g.P("}")
	g.P()
}
