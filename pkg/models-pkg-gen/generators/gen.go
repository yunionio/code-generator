package generators

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"k8s.io/gengo/args"
	"k8s.io/gengo/generator"
	"k8s.io/gengo/namer"
	"k8s.io/gengo/types"
	"k8s.io/klog"

	"yunion.io/x/pkg/util/sets"

	"yunion.io/x/code-generator/pkg/common"
)

func NameSystems() namer.NameSystems {
	return namer.NameSystems{
		"public":  namer.NewPublicNamer(0),
		"private": namer.NewPrivateNamer(0),
		"raw":     namer.NewRawNamer("", nil),
	}
}

func DefaultNameSystem() string {
	return "public"
}

func Packages(ctx *generator.Context, arguments *args.GeneratorArgs) generator.Packages {
	boilerplate, err := arguments.LoadGoBoilerplate()
	if err != nil {
		klog.Fatalf("Failed loading boilerplate: %v", err)
	}
	pkgs := generator.Packages{}
	inputs := sets.NewString(ctx.Inputs...)
	header := append([]byte(fmt.Sprintf("// +build !%s\n\n", arguments.GeneratedBuildTag)), boilerplate...)

	for i := range inputs {
		pkg := ctx.Universe[i]
		if pkg == nil {
			continue
		}
		klog.Infof("Considering pkg %q", pkg.Path)
		outPkgName := strings.Split(filepath.Base(arguments.OutputPackagePath), ".")[0]
		pkgs = append(pkgs,
			&generator.DefaultPackage{
				PackageName: outPkgName,
				PackagePath: arguments.OutputPackagePath,
				HeaderText:  header,
				GeneratorFunc: func(c *generator.Context) []generator.Generator {
					return []generator.Generator{
						// Always generate a "doc.go" file.
						generator.DefaultGen{OptionalName: "doc"},
						// Generate swagger code by model.
						NewModelPkgGen(arguments.OutputFileBaseName, pkg.Path, ctx.Order),
					}
				},
			})
	}
	return pkgs
}

type modelPkgGen struct {
	generator.DefaultGen
	ident         string
	sourcePackage string
	modelManagers map[string]*types.Type
}

func NewModelPkgGen(sanitizedName, sourcePackage string, pkgTypes []*types.Type) generator.Generator {
	ident := filepath.Base(strings.TrimRight(sourcePackage, "models"))
	gen := &modelPkgGen{
		DefaultGen: generator.DefaultGen{
			OptionalName: fmt.Sprintf("%s_%s", sanitizedName, ident),
		},
		ident:         ident,
		sourcePackage: sourcePackage,
		modelManagers: make(map[string]*types.Type),
	}
	gen.collectTypes(pkgTypes)
	return gen
}

func (g *modelPkgGen) collectTypes(pkgTypes []*types.Type) {
	common.CollectModelManager(g.sourcePackage, pkgTypes, sets.NewString(), g.modelManagers)
	for _, man := range g.modelManagers {
		g.modelManagers[man.String()] = man
	}
}

func (g *modelPkgGen) Filter(c *generator.Context, t *types.Type) bool {
	if key, ok := g.modelManagers[t.String()]; ok && strings.HasSuffix(key.Name.Name, "Manager") {
		return true
	}
	return false
}

func (g *modelPkgGen) Imports(c *generator.Context) []string {
	return []string{g.sourcePackage}
}

func (g *modelPkgGen) Init(c *generator.Context, w io.Writer) error {
	sw := common.NewSnippetWriter(w, c)
	sw.Do("func init() {\n", nil)
	return sw.Error()
}

func (g *modelPkgGen) GenerateType(c *generator.Context, t *types.Type, w io.Writer) error {
	sw := common.NewSnippetWriter(w, c)
	g.generateCode(t, sw)
	return sw.Error()
}

func getArgs(t *types.Type) interface{} {
	return common.GetArgs(t)
}

func (g *modelPkgGen) generateCode(t *types.Type, sw *generator.SnippetWriter) {
	args := getArgs(t)
	pkg := filepath.Base(t.Name.Package)
	manName := t.Name.Name
	if manName[0] == 'S' {
		manName = manName[1:len(manName)]
	}
	sw.Do(fmt.Sprintf("RegisterModelManager(%s.%s)\n", pkg, manName), args)
}

func (g *modelPkgGen) Finalize(c *generator.Context, w io.Writer) error {
	sw := common.NewSnippetWriter(w, c)
	sw.Do("}", nil)
	return sw.Error()
}
