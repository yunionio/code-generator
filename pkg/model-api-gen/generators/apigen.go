package generators

import (
	"fmt"
	"io"
	"path/filepath"
	"reflect"
	"strings"

	"golang.org/x/tools/imports"
	"k8s.io/gengo/args"
	"k8s.io/gengo/generator"
	"k8s.io/gengo/namer"
	"k8s.io/gengo/types"
	"k8s.io/klog"

	"yunion.io/x/pkg/errors"
	"yunion.io/x/pkg/util/reflectutils"
	"yunion.io/x/pkg/util/sets"
	"yunion.io/x/pkg/utils"

	"yunion.io/x/code-generator/pkg/common"
	"yunion.io/x/code-generator/pkg/swagger-gen/generators"
)

const (
	tagName = "onecloud:model-api-gen"
	//tagPkgName           = "onecloud:model-api-gen-pkg"
	SModelBase           = "SModelBase"
	CloudCommonDBPackage = "yunion.io/x/onecloud/pkg/cloudcommon/db"
	CloudProviderPackage = "yunion.io/x/cloudmux/pkg/cloudprovider"
	MonitorModelsPackage = "yunion.io/x/onecloud/pkg/monitor/models"
)

func extractTag(comments []string) []string {
	return types.ExtractCommentTags("+", comments)[tagName]
}

/*func checkTag(comments []string, require ...string) bool {
	vals := types.ExtractCommentTags("+", comments)[tagName]
	if len(require) == 0 {
		return len(vals) == 1 && vals[0] == ""
	}
	return reflect.DeepEqual(vals, require)
}*/

/*func extractPkgTag(comments []string) []string {
	return types.ExtractCommentTags("+", comments)[tagPkgName]
}*/

var (
	APIsPackage              = "yunion.io/x/onecloud/pkg/apis"
	APIsCloudProviderPackage = filepath.Join(APIsPackage, "cloudprovider")
	APIsMonitorPackage       = filepath.Join(APIsPackage, "monitor")
)

func GetInputOutputPackageMap(apisPkg string) map[string]string {
	ret := map[string]string{
		CloudCommonDBPackage: APIsPackage,
		CloudProviderPackage: APIsCloudProviderPackage, // filepath.Join(apisPkg, "cloudprovider"),
		MonitorModelsPackage: APIsMonitorPackage,
	}
	return ret
}

// NameSystems returns the name system used by the generators in this package.
func NameSystems() namer.NameSystems {
	return namer.NameSystems{
		"public":  namer.NewPublicNamer(0),
		"private": namer.NewPrivateNamer(0),
		"raw":     namer.NewRawNamer("", nil),
	}
}

// DefaultNameSystem returns the default name system for ordering the types to be
// processed by the generators in this package.
func DefaultNameSystem() string {
	return "public"
}

// Packages makes the api-gen package definition.
func Packages(ctx *generator.Context, arguments *args.GeneratorArgs) generator.Packages {
	boilerplate, err := arguments.LoadGoBoilerplate()
	if err != nil {
		klog.Fatalf("Failed loading boilerplate: %v", err)
	}

	inputs := sets.NewString(ctx.Inputs...)
	packages := generator.Packages{}
	//header := append([]byte(fmt.Sprintf("// +build !%s\n\n", arguments.GeneratedBuildTag)), boilerplate...)

	for i := range inputs {
		pkg := ctx.Universe[i]
		if pkg == nil {
			// If the input had no Go files, for example
			continue
		}
		klog.Infof("Considering pkg %q", pkg.Path)
		//pkgPath := pkg.Path
		outPkgName := strings.Split(filepath.Base(arguments.OutputPackagePath), ".")[0]
		packages = append(packages,
			&generator.DefaultPackage{
				PackageName: outPkgName,
				PackagePath: arguments.OutputPackagePath,
				HeaderText:  boilerplate,
				GeneratorFunc: func(c *generator.Context) []generator.Generator {
					return []generator.Generator{
						// Always generate a "doc.go" file.
						// generator.DefaultGen{OptionalName: "doc"},
						// Generate api types by model.
						NewApiGen(arguments.OutputFileBaseName, pkg.Path, "", ctx.Order, arguments.OutputPackagePath),
					}
				},
			})
	}
	return packages
}

type apiGen struct {
	generator.DefaultGen
	// sourcePackage is source package of input types
	sourcePackage string
	// modelTypes record all model types in source package
	modelTypes sets.String
	// modelDependTypes record all model required types
	modelDependTypes sets.String
	// isCommonDBPackage
	isCommonDBPackage bool

	imports            namer.ImportTracker
	needImportPackages sets.String
	apisPkg            string
	outputPackage      string
}

func isCommonDBPackage(pkg string) bool {
	// pkg is yunion.io/x/onecloud/pkg/cloudcommon/db
	return strings.HasSuffix(pkg, CloudCommonDBPackage)
}

func defaultAPIsPkg(srcPkg string) string {
	yunionPrefix := "yunion.io/x/"
	parts := strings.Split(srcPkg, yunionPrefix)
	projectName := strings.Split(parts[1], "/")[0]
	return filepath.Join(yunionPrefix, projectName, "pkg", "apis")
}

func reviseImportPath() {
	imports.LocalPrefix = "yunion.io/x/:yunion.io/x/onecloud:yunion.io/x/meter:yunion.io/x/nocloud"
}

func NewApiGen(sanitizedName, sourcePackage, apisPkg string, pkgTypes []*types.Type, outputPkg string) generator.Generator {
	reviseImportPath()
	if apisPkg == "" {
		apisPkg = defaultAPIsPkg(sourcePackage)
	}
	gen := &apiGen{
		DefaultGen: generator.DefaultGen{
			OptionalName: sanitizedName,
		},
		sourcePackage:      sourcePackage,
		modelTypes:         sets.NewString(),
		modelDependTypes:   sets.NewString(),
		isCommonDBPackage:  isCommonDBPackage(sourcePackage),
		imports:            generator.NewImportTracker(),
		needImportPackages: sets.NewString(),
		apisPkg:            apisPkg,
		outputPackage:      outputPkg,
	}
	gen.collectTypes(pkgTypes)
	klog.V(1).Infof("sets: %v\ndepsets: %v", gen.modelTypes.List(), gen.modelDependTypes.List())
	return gen
}

func (g *apiGen) Namers(c *generator.Context) namer.NameSystems {
	// Have the raw namer for this file track what it imports.
	return namer.NameSystems{
		"public": namer.NewPublicNamer(0),
		"raw":    namer.NewRawNamer("", g.imports),
	}
}

func (g *apiGen) GetInputOutputPackageMap() map[string]string {
	return GetInputOutputPackageMap(g.apisPkg)
}

func (g *apiGen) collectTypes(pkgTypes []*types.Type) {
	for _, t := range pkgTypes {
		if t.Kind != types.Struct && t.Kind != types.Alias {
			continue
		}
		if !g.inSourcePackage(t) {
			continue
		}
		if common.IsPrivateStruct(t.Name.Name) {
			continue
		}
		if includeType(t) || g.isResourceModel(t) {
			g.modelTypes.Insert(t.String())
			g.addDependTypes(t, g.modelTypes, g.modelDependTypes)
		}
	}
}

// getPrimitiveType return the primitive type of Map, Slice, Pointer or Chan
func getPrimitiveType(t *types.Type) *types.Type {
	var compondKinds = sets.NewString(
		string(types.Map),
		string(types.Slice),
		string(types.Pointer),
		string(types.Chan))

	if !compondKinds.Has(string(t.Kind)) {
		return t
	}

	et := t.Elem
	return getPrimitiveType(et)
}

func isModelBase(t *types.Type) bool {
	return t.Name.Name == SModelBase
}

func (g *apiGen) addDependTypes(t *types.Type, out, dependOut sets.String) {
	if t.Kind == types.Alias || t.Kind == types.Struct {
		if !out.Has(t.String()) && g.inSourcePackage(t) && !isModelBase(t) {
			dependOut.Insert(t.String())
		}
		umt := underlyingType(t)
		if umt.Kind == types.Builtin {
			return
		}
		t = getPrimitiveType(umt)
		if t.Kind == types.Builtin {
			return
		}
		if !out.Has(t.String()) && g.inSourcePackage(t) && !isModelBase(t) {
			dependOut.Insert(t.String())
		}
	}
	for _, m := range t.Members {
		switch m.Type.Kind {
		case types.Struct, types.Alias:
			g.addDependTypes(m.Type, out, dependOut)
		case types.Pointer, types.Slice:
			g.addDependTypes(m.Type.Elem, out, dependOut)
		}
	}
}

func includeType(t *types.Type) bool {
	vals := extractTag(t.CommentLines)
	if len(vals) != 0 {
		return true
	}
	return false
}

func (g *apiGen) Filter(c *generator.Context, t *types.Type) bool {
	if generators.IncludeIgnoreTag(t) {
		return false
	}
	if g.modelTypes.Has(t.String()) {
		return true
	}
	if g.modelDependTypes.Has(t.String()) {
		return true
	}
	return false
}

func (g *apiGen) isResourceModel(t *types.Type) bool {
	val := common.IsResourceModel(t)
	return val
}

func (g *apiGen) args(t *types.Type) interface{} {
	a := generator.Args{
		"type": t,
	}
	return a
}

func (g *apiGen) Imports(c *generator.Context) []string {
	lines := []string{}
	for _, line := range g.imports.ImportLines() {
		if strings.Index(line, CloudProviderPackage) > -1 || strings.Index(line, CloudCommonDBPackage) > -1 {
			continue
		}
		parts := strings.Split(line, " ")
		if len(parts) == 2 {
			pkgName := strings.Trim(parts[1], `"`)
			if g.needImportPackages.Has(pkgName) || pkgName == g.outputPackage {
				continue
			}
		}
		lines = append(lines, line)
	}
	lines = append(lines, g.needImportPackages.List()...)
	return lines
}

func (g *apiGen) GenerateType(c *generator.Context, t *types.Type, w io.Writer) error {
	klog.V(1).Infof("Generating api model for type %s", t.String())

	err := g.generateTypeForOp(c, t, w)
	if err != nil {
		return errors.Wrap(err, "generateTypeForOp")
	}
	return nil
}

func (g *apiGen) generateTypeForOp(c *generator.Context, t *types.Type, w io.Writer) error {
	sw := generator.NewSnippetWriter(w, c, "$", "$")

	sw.Do(fmt.Sprintf("// %s is an autogenerated struct via %s.\n", t.Name.Name, t.Name.String()), nil)

	// 1. generate resource base output by model to pkg/apis/<pkg>/generated.model.go
	switch t.Kind {
	case types.Struct:
		g.generateStructType(t, sw)
	case types.Alias:
		g.generatorAliasType(t, sw)
	default:
		klog.Fatalf("Unsupported type %s %s", t, t.Kind)
	}
	return sw.Error()
}

func (g *apiGen) generateStructType(t *types.Type, sw *generator.SnippetWriter) {
	//klog.Errorf("for type %q", t.String())
	sw.Do("type $.type|public$ struct {\n", g.args(t))
	g.generateFor(t, sw)
	sw.Do("}\n", nil)
}

func (g *apiGen) generatorAliasType(t *types.Type, sw *generator.SnippetWriter) {
	sw.Do("type $.type|public$ ", g.args(t))
	ut := t.Underlying
	switch ut.Kind {
	case types.Slice:
		elem := ut.Elem
		content := "$.type|public$"
		if elem.Kind == types.Pointer {
			content = g.getPointerSourcePackageName(elem)
		} else if elem.Kind == types.Builtin {
			content = "$.type$"
		}
		sw.Do(fmt.Sprintf("[]%s", content), g.args(elem))
	case types.Builtin:
		sw.Do("$.type$ ", g.args(t.Underlying))
	default:
		sw.Do("$.type|public$ ", g.args(t.Underlying))
	}
	sw.Do("\n", nil)
}

func underlyingType(t *types.Type) *types.Type {
	for t.Kind == types.Alias {
		t = t.Underlying
	}
	return t
}

func (g *apiGen) needCopy(t *types.Type) bool {
	tStr := t.String()
	return !g.modelTypes.Has(tStr) && g.modelDependTypes.Has(tStr) && t.Kind != types.Builtin
}

func (g *apiGen) generateFor(t *types.Type, sw *generator.SnippetWriter) {
	for _, mem := range t.Members {
		g.generateForMember(t, mem, sw)
	}
}

func (g *apiGen) generateForMember(t *types.Type, mem types.Member, sw *generator.SnippetWriter) {
	if common.IsPrivateStruct(mem.Name) {
		return
	}
	mt := mem.Type
	if isModelBase(mt) {
		return
	}

	info := reflectutils.ParseFieldJsonInfo(mem.Name, reflect.StructTag(mem.Tags))
	if info.Ignore {
		return
	}
	if val, ok := info.Tags["ignore"]; ok && val == "true" {
		return
	}

	var f func(types.Member, *generator.SnippetWriter)
	switch mt.Kind {
	case types.Builtin:
		f = g.doBuiltin
	case types.Struct:
		f = func(member types.Member, sw *generator.SnippetWriter) {
			g.doStruct(t, member, sw)
		}
	case types.Interface:
		f = g.doInterface
	case types.Alias:
		f = func(member types.Member, sw *generator.SnippetWriter) {
			g.doAlias(t, member, sw)
		}
	case types.Pointer:
		f = g.doPointer
	case types.Slice:
		f = g.doSlice
	case types.Map:
		f = g.doMap
	default:
		klog.Fatalf("Hit an unsupported type %v.%s, kind is %s", t, mt.Name.Name, mt.Kind)
		//klog.Warningf("Hit an unsupported type %v.%s, kind is %s", t, mt.Name.Name, mt.Kind)
	}
	f(mem, sw)
}

type Member struct {
	name         string
	jsonTags     []string
	mType        string
	namer        string
	embedded     bool
	commentLines []string
}

func NewMember(name string, commentLines []string) *Member {
	clines := []string{}
	for _, cl := range commentLines {
		if len(cl) == 0 {
			continue
		}
		clines = append(clines, fmt.Sprintf("// %s", cl))
	}
	return &Member{
		name:         name,
		jsonTags:     make([]string, 0),
		namer:        "raw",
		commentLines: clines,
	}
}

// Type override types.Type raw type
func (m *Member) Type(mType string) *Member {
	m.mType = mType
	return m
}

func (m *Member) Name(name string) *Member {
	m.name = name
	return m
}

func (m *Member) Namer(namer string) *Member {
	m.namer = namer
	return m
}

func (m *Member) Embedded() *Member {
	m.embedded = true
	return m
}

func (m *Member) AddTag(tags ...string) *Member {
	for _, t := range tags {
		if !utils.IsInStringArray(t, m.jsonTags) {
			m.jsonTags = append(m.jsonTags, t)
		}
	}
	return m
}

func (m *Member) NoTag() *Member {
	m.jsonTags = nil
	return m
}

func memberJsonName(m types.Member) string {
	info := reflectutils.ParseFieldJsonInfo(m.Name, reflect.StructTag(m.Tags))
	return info.MarshalName()
}

func NewModelMember(member types.Member) *Member {
	m := NewMember(member.Name, member.CommentLines)
	return m.AddTag(memberJsonName(member))
}

func (m *Member) Do(sw *generator.SnippetWriter, args interface{}) {
	var (
		typePart string
		ret      string
	)
	namePart := m.name
	if m.mType != "" {
		typePart = m.mType
	} else {
		typePart = fmt.Sprintf("$.type|%s$", m.namer)
	}
	if m.embedded {
		ret = typePart
	} else {
		ret = fmt.Sprintf("%s %s", namePart, typePart)
	}
	if len(m.commentLines) != 0 {
		ret = fmt.Sprintf("%s\n%s", strings.Join(m.commentLines, "\n"), ret)
	}
	if len(m.jsonTags) != 0 {
		ret = fmt.Sprintf("%s `json:\"%s\"`", ret, strings.Join(m.jsonTags, ","))
	}
	sw.Do(fmt.Sprintf("%s\n", ret), args)
}

func (g *apiGen) doSlice(m types.Member, sw *generator.SnippetWriter) {
	memType := m.Type.Elem
	point := false
	if memType.Kind == types.Pointer {
		memType = memType.Elem
		point = true
	}
	memPkg := memType.Name.Package
	modelMem := NewModelMember(m)
	if memPkg == g.outputPackage || memPkg == g.sourcePackage {
		if point {
			modelMem.Type(fmt.Sprintf("[]*%s", memType.Name.Name))
		} else {
			modelMem.Type(fmt.Sprintf("[]%s", memType.Name.Name))
		}
	}
	modelMem.Do(sw, g.args(m.Type))
}

func (g *apiGen) doMap(m types.Member, sw *generator.SnippetWriter) {
	NewModelMember(m).Do(sw, g.args(m.Type))
}

func (g *apiGen) doBuiltin(m types.Member, sw *generator.SnippetWriter) {
	NewModelMember(m).Do(sw, g.args(m.Type))
}

var (
	TypeMap = map[string]struct {
		Type     string
		JSONTags []string
	}{
		"TriState": {
			"*bool",
			[]string{"omitempty"},
		},
	}
)

func (g *apiGen) doAlias(parentType *types.Type, member types.Member, sw *generator.SnippetWriter) {
	mt := member.Type
	if ct, ok := TypeMap[mt.Name.Name]; ok {
		m := NewModelMember(member).AddTag(ct.JSONTags...).Type(ct.Type)
		m.Do(sw, nil)
		return
	}
	ut := underlyingType(mt)
	// NewModelMember(member).Do(sw, g.args(ut))
	member.Type = ut
	g.generateForMember(parentType, member, sw)
}

func (g *apiGen) doStruct(parentType *types.Type, member types.Member, sw *generator.SnippetWriter) {
	mt := member.Type
	klog.V(1).Infof("doStruct for memeber %s of %s", mt.Name.String(), parentType.String())
	//inPkg := g.inSourcePackage(member.Type)
	m := NewModelMember(member)
	if member.Embedded {
		m.Embedded()
		m.NoTag()
	} else {
		// if not allow to get or list, then ignore the field
		info := reflectutils.ParseFieldJsonInfo(member.Name, reflect.StructTag(member.Tags))
		_, getTag := info.Tags["get"]
		_, listTag := info.Tags["list"]
		if !getTag && !listTag && g.modelTypes.Has(parentType.String()) {
			klog.V(1).Infof("doStruct ignore for memeber %s of %s cause of no get and list tag", mt.String(), parentType.String())
			return
		}
	}
	if g.inSourcePackage(mt) {
		m.Namer("public")
	} else if g.inOutputPackage(mt) {
		m.Type(mt.Name.Name)
	} else {
		if outPkg, ok := g.GetInputOutputPackageMap()[mt.Name.Package]; ok {
			g.needImportPackages.Insert(outPkg)
			m.Type(fmt.Sprintf("%s.%s", filepath.Base(outPkg), mt.Name.Name))
		} else if strings.HasPrefix(mt.Name.Package, g.sourcePackage+"/") {
			outPkg := filepath.Join(g.outputPackage, mt.Name.Package[len(g.sourcePackage)+1:])
			g.needImportPackages.Insert(outPkg)
			m.Type(fmt.Sprintf("%s.%s", filepath.Base(outPkg), mt.Name.Name))
		}
	}
	m.Do(sw, g.args(mt))
}

func (g *apiGen) doInterface(m types.Member, sw *generator.SnippetWriter) {
	// model can't embedded interface
	if m.Embedded {
		klog.Fatalf("%s used as embedded interface", m.String())
	}
	mem := NewModelMember(m)
	mem.Do(sw, g.args(m.Type))
}

func (g *apiGen) inSourcePackage(t *types.Type) bool {
	return common.InSourcePackage(t, g.sourcePackage)
}

func (g *apiGen) inOutputPackage(t *types.Type) bool {
	return common.IsSamePackage(t, g.outputPackage)
}

func (g *apiGen) inJSONUtilsPackage(t *types.Type) bool {
	ut := underlyingType(t)
	if t.Kind == types.Pointer {
		ut = t.Elem
	}
	return strings.Contains(ut.Name.Package, "yunion.io/x/jsonutils")
}

func (g *apiGen) getPointerSourcePackageName(t *types.Type) string {
	elem := t.Elem
	if !g.inSourcePackage(elem) {
		klog.Fatalf("pointer's elem %q not in package %q", elem.Name.String(), g.sourcePackage)
	}
	return fmt.Sprintf("*%s", elem.Name.Name)
}

func (g *apiGen) doPointer(m types.Member, sw *generator.SnippetWriter) {
	t := m.Type
	mem := NewModelMember(m)
	elem := m.Type.Elem
	if g.inSourcePackage(elem) {
		mem.Type(g.getPointerSourcePackageName(t))
	} else if g.inOutputPackage(elem) {
		mem.Type(fmt.Sprintf("*%s", elem.Name.Name))
	}
	if m.Embedded {
		mem.Embedded()
		mem.NoTag()
	}
	args := g.args(m.Type)
	mem.Do(sw, args)
}

type ResourceModel struct {
	t *types.Type
}

func NewModelByType(t *types.Type) *ResourceModel {
	return &ResourceModel{
		t: t,
	}
}
