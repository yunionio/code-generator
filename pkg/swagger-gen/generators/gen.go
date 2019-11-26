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

	"yunion.io/x/log"
	"yunion.io/x/onecloud/pkg/cloudcommon/db"
	"yunion.io/x/pkg/util/sets"

	"yunion.io/x/code-generator/pkg/common"
	"yunion.io/x/code-generator/pkg/models"
)

// 1. find model rest api functions and parameters
// 2. according step 1 result generate swagger spec

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

	outPkgName := strings.Split(filepath.Base(arguments.OutputPackagePath), ".")[0]
	pkgPath := arguments.OutputPackagePath
	svcName := outPkgName
	pkgs = append(pkgs, NewDocPackage(outPkgName, pkgPath, header, svcName))
	for i := range inputs {
		pkg := ctx.Universe[i]
		if pkg == nil {
			continue
		}
		klog.Infof("Considering pkg %q", pkg.Path)
		pkgs = append(pkgs,
			&generator.DefaultPackage{
				PackageName: outPkgName,
				PackagePath: pkgPath,
				HeaderText:  header,
				GeneratorFunc: func(c *generator.Context) []generator.Generator {
					return []generator.Generator{
						// Generate swagger code by model.
						NewSwaggerGen(arguments.OutputFileBaseName, pkg.Path, ctx.Order),
					}
				},
			})
	}
	return pkgs
}

type swaggerGen struct {
	generator.DefaultGen
	sourcePackage string
	modelTypes    sets.String
	modelManagers map[string]*types.Type
}

func NewSwaggerGen(sanitizedName, sourcePackage string, pkgTypes []*types.Type) generator.Generator {
	ident := filepath.Base(strings.TrimRight(sourcePackage, "models"))
	gen := &swaggerGen{
		DefaultGen: generator.DefaultGen{
			OptionalName: fmt.Sprintf("%s_%s", sanitizedName, ident),
		},
		sourcePackage: sourcePackage,
		modelTypes:    sets.NewString(),
		modelManagers: make(map[string]*types.Type),
	}
	gen.collectTypes(pkgTypes)
	klog.V(5).Infof("modelTypes: %v, modelManagers: %v", gen.modelTypes.List(), gen.modelManagers)
	return gen
}

func (g *swaggerGen) collectTypes(pkgTypes []*types.Type) {
	common.CollectModelManager(g.sourcePackage, pkgTypes, g.modelTypes, g.modelManagers)
}

func (g *swaggerGen) getModelManager(t *types.Type) *types.Type {
	return g.modelManagers[t.String()]
}

func (g *swaggerGen) getModelManagerInstance(t *types.Type) db.IModelManager {
	mt := g.getModelManager(t)
	return models.GetModelManagerByType(mt)
}

func isModelManagerRegistered(mt *types.Type) bool {
	if mt == nil {
		return false
	}
	man := models.GetModelManagerByType(mt)
	if man == nil {
		return false
	}
	return true
}

func (g *swaggerGen) Filter(c *generator.Context, t *types.Type) bool {
	if g.modelTypes.Has(t.String()) && isModelManagerRegistered(g.getModelManager(t)) {
		return true
	}
	return false
}

func (g *swaggerGen) GenerateType(c *generator.Context, t *types.Type, w io.Writer) error {
	klog.V(2).Infof("Generating api model for type %s", t)
	sw := generator.NewSnippetWriter(w, c, "$", "$")
	g.generateCode(g.getModelManager(t), t, sw)
	return sw.Error()
}

func (g *swaggerGen) generateCode(manType *types.Type, modelType *types.Type, sw *generator.SnippetWriter) {
	manIns := g.getModelManagerInstance(modelType)
	parser := newTypeParser(manIns, manType, modelType)

	generateCreate(parser.createM(), sw)
	getM := parser.getM()
	generateGet(getM, sw)
	lm := parser.listM()
	generateList(lm, getM, sw)
	generateUpdate(parser.updateM(), sw)
	generateDelete(parser.deleteM(), sw)

	applyGenerateFunc(generateGetSpec, parser.getSpecM, sw)
	applyGenerateFunc(generatePerformAction, parser.performActionM, sw)
}

func applyGenerateFunc(genFunc func(*Method, *generator.SnippetWriter), getMethods func() []*Method, sw *generator.SnippetWriter) {
	for _, m := range getMethods() {
		genFunc(m, sw)
	}
}

const (
	// model or model manager func keyword
	Create                      = "ValidateCreateData"
	List                        = "ListItemFilter"
	Get                         = "GetExtraDetails"
	GetCustomizedGetDetailsBody = "CustomizedGetDetailsBody"
	GetSpec                     = "GetDetails"
	GetProperty                 = "GetProperty"
	Perform                     = "Perform"
	Update                      = "ValidateUpdateData"
	Delete                      = "CustomizeDelete"
)

type Method struct {
	resSingular string
	resPlural   string
	receiver    *types.Type
	name        string
	method      *types.Type
}

func NewMethod(receiver *types.Type, name string, method *types.Type, singular, plural string) *Method {
	return &Method{
		receiver:    receiver,
		name:        name,
		method:      method,
		resSingular: singular,
		resPlural:   plural,
	}
}

func (m *Method) Receiver() *types.Type {
	return m.receiver
}

func (m *Method) Name() string {
	return m.name
}

func (m *Method) Signature() *types.Signature {
	return m.method.Signature
}

func (m *Method) Params(idx int) *types.Type {
	return m.Signature().Parameters[idx]
}

func (m *Method) Resutls(idx int) *types.Type {
	return m.Signature().Results[idx]
}

func (m *Method) Method() *types.Type {
	return m.method
}

func (m *Method) String() string {
	return fmt.Sprintf("%s.%s", m.Receiver().String(), m.Name())
}

func getTypeMethods(
	funcPrefixKeyword string,
	keyword, keywordPlural string,
	t *types.Type,
	predicateF func(*Method) bool,
) []*Method {
	if t.Methods == nil {
		return nil
	}
	methods := make([]*Method, 0)
	for name, m := range t.Methods {
		if strings.HasPrefix(name, funcPrefixKeyword) {
			useIt := true
			mWrap := NewMethod(t, name, m, keyword, keywordPlural)
			if predicateF != nil {
				useIt = predicateF(mWrap)
			}
			if !useIt {
				continue
			}
			methods = append(methods, mWrap)
		}
	}
	return methods
}

type typeParser struct {
	managerInstance db.IModelManager
	manager         *types.Type
	model           *types.Type
	singular        string
	plural          string
}

func newTypeParser(manIns db.IModelManager, man *types.Type, model *types.Type) *typeParser {
	keyword, keywordPlural := getManagerKeywords(manIns)
	return &typeParser{
		managerInstance: manIns,
		manager:         man,
		model:           model,
		singular:        keyword,
		plural:          keywordPlural,
	}
}

func getManagerKeywords(man db.IModelManager) (string, string) {
	return man.Keyword(), man.KeywordPlural()
}

func validInputOutput(input, output *types.Type) error {
	for key, t := range map[string]*types.Type{
		"input":  input,
		"output": output,
	} {
		if isStructPointer(t) {
			return fmt.Errorf("invalid %s %s kind %s", key, t.String(), t.Kind)
		}
	}
	return nil
}

func isStructPointer(t *types.Type) bool {
	if t.Kind != types.Pointer {
		return false
	}
	elem := t.Elem
	if elem.Kind != types.Struct {
		return false
	}
	if strings.Contains(elem.Name.Package, "yunion.io/x/jsonutils") {
		return false
	}
	return true
}

func (p *typeParser) createM() *Method {
	return p.getMethod(Create, p.manager,
		func(m *Method) bool {
			sig := m.Signature()
			paramsLen := len(sig.Parameters)
			retLen := len(sig.Results)
			// ValidateCreateData(context.Context, mcclient.TokenCredential, mcclient.IIdentityProvider, query jsonutils.JSONObject, data *jsonutils.JSONDict) (Object, error)
			if paramsLen != 5 || retLen != 2 {
				return false
			}
			return true
		},
	)
}

func (p *typeParser) listM() *Method {
	return p.getMethod(List, p.manager,
		func(m *Method) bool {
			sig := m.Signature()
			paramsLen := len(sig.Parameters)
			retLen := len(sig.Results)
			// ListItemFilter(context.Context, *sqlchemy.SQuery, mcclient.TokenCredential, query jsonutils.JSONObject) (*sqlchemy.SQuery, error)
			if paramsLen != 4 || retLen != 2 {
				return false
			}
			return true
		})
}

func (p *typeParser) getM() *Method {
	return p.getMethod(Get, p.model,
		func(m *Method) bool {
			sig := m.Signature()
			paramsLen := len(sig.Parameters)
			retLen := len(sig.Results)
			// GetExtraDetails(context.Context, userCred mcclient.TokenCredential, query Object) (Object, error)
			if paramsLen != 3 || retLen != 2 {
				return false
			}
			return true
		})
}

func (p *typeParser) updateM() *Method {
	return p.getMethod(Update, p.model, func(m *Method) bool {
		sig := m.Signature()
		paramsLen := len(sig.Parameters)
		retLen := len(sig.Results)
		// ValidateUpdateData(context.Context, mcclient.TokenCredential, query Object, data Object) (Object, error)
		if paramsLen != 4 || retLen != 2 {
			return false
		}
		return true
	})
}

func (p *typeParser) deleteM() *Method {
	return p.getMethod(Delete, p.model, func(m *Method) bool {
		sig := m.Signature()
		paramsLen := len(sig.Parameters)
		retLen := len(sig.Results)
		// CustomizeDelete(context.Context, mcclient.TokenCredential, query Object, body Object) error
		if paramsLen != 4 || retLen != 1 {
			return false
		}
		return true
	})
}

func (p *typeParser) performActionM() []*Method {
	return p.getMethods(Perform, p.model,
		func(m *Method) bool {
			sig := m.Signature()
			paramsLen := len(sig.Parameters)
			retLen := len(sig.Results)
			if paramsLen != 4 || retLen != 2 {
				return false
			}
			body := sig.Parameters[3]
			output := sig.Results[0]
			// input body and output must struct pointer
			if err := validInputOutput(body, output); err != nil {
				log.Warningf("validInputOutput for method %s: %v", m.String(), err)
				//return false
			}
			return true
		},
	)
}

func (p *typeParser) getSpecM() []*Method {
	return p.getMethods(GetSpec, p.model,
		func(m *Method) bool {
			paramsLen := len(m.Signature().Parameters)
			retLen := len(m.Signature().Results)
			if paramsLen != 3 || retLen != 2 {
				return false
			}
			sig := m.Signature()
			//input := sig.Parameters[2]
			output := sig.Results[0]
			if !isStructPointer(output) {
				log.Warningf("method %s: output is not struct pointer", m.String())
				//return false
			}
			return true
		},
	)
}

func (p *typeParser) getMethods(funcPreKeyword string, model *types.Type, preF func(*Method) bool) []*Method {
	return getTypeMethods(funcPreKeyword, p.singular, p.plural, model, preF)
}

func (p *typeParser) getMethod(funcPreKeyword string, model *types.Type, preF func(*Method) bool) *Method {
	ms := p.getMethods(funcPreKeyword, model, preF)
	if len(ms) == 0 {
		return nil
	}
	return ms[0]
}

func getArgs(t *types.Type) interface{} {
	return common.GetArgs(t)
}

type commenter struct {
	route     *route
	parameter *parameter
	response  *response
}

func (c commenter) Do(sw *generator.SnippetWriter) {
	for _, f := range []func(*generator.SnippetWriter){
		c.route.Do,
		c.parameter.Do,
		c.response.Do,
	} {
		f(sw)
	}
}

type snippetWriter struct {
	sw *generator.SnippetWriter
}

func newSW(sw *generator.SnippetWriter) *snippetWriter {
	return &snippetWriter{sw}
}

func (w snippetWriter) lines(lines []string) {
	for _, l := range lines {
		w.sw.Do(fmt.Sprintf("// %s\n", l), nil)
	}
}

func (w snippetWriter) emptyLine() {
	w.sw.Do("//\n", nil)
}

func (w snippetWriter) line(l string) {
	w.lines([]string{l})
}

func generateCreate(method *Method, sw *generator.SnippetWriter) {
	if method == nil {
		return
	}
	param := newParameterFactory(method).Create()
	resp := newResponseFactory(method).FirstSingularResult()
	route := newRouteFactory(method).Create(param, resp)
	c := &commenter{
		route:     route,
		parameter: param,
		response:  resp,
	}
	c.Do(sw)
}

func generateList(listMethod, getMethod *Method, sw *generator.SnippetWriter) {
	if listMethod == nil || getMethod == nil {
		return
	}
	param := newParameterFactory(listMethod).List()
	resp := newResponseFactory(listMethod).ListResult(getMethod)
	route := newRouteFactory(listMethod).List(param, resp)
	c := &commenter{
		route:     route,
		parameter: param,
		response:  resp,
	}
	c.Do(sw)
}

func generateGet(method *Method, sw *generator.SnippetWriter) {
	if method == nil {
		return
	}
	param := newParameterFactory(method).Get()
	resp := newResponseFactory(method).FirstSingularResult()
	route := newRouteFactory(method).Get(param, resp)
	c := &commenter{
		route:     route,
		parameter: param,
		response:  resp,
	}
	c.Do(sw)
}

func generateUpdate(method *Method, sw *generator.SnippetWriter) {
	if method == nil {
		return
	}
	param := newParameterFactory(method).Update()
	resp := newResponseFactory(method).FirstSingularResult()
	route := newRouteFactory(method).Update(param, resp)
	c := &commenter{
		route:     route,
		parameter: param,
		response:  resp,
	}
	c.Do(sw)
}

func generateDelete(method *Method, sw *generator.SnippetWriter) {
	if method == nil {
		return
	}
	param := newParameterFactory(method).Delete()
	resp := newResponseFactory(method).FirstSingularResult()
	route := newRouteFactory(method).Delete(param, resp)
	c := &commenter{
		route:     route,
		parameter: param,
		response:  resp,
	}
	c.Do(sw)
}

func generateGetSpec(method *Method, sw *generator.SnippetWriter) {
	if method == nil {
		return
	}
	param := newParameterFactory(method).GetSpec()
	resp := newResponseFactory(method).FirstSingularResult()
	route := newRouteFactory(method).GetSpec(param, resp)
	c := &commenter{
		route:     route,
		parameter: param,
		response:  resp,
	}
	c.Do(sw)
}

func generatePerformAction(method *Method, sw *generator.SnippetWriter) {
	if method == nil {
		return
	}
	param := newParameterFactory(method).PerformAction()
	resp := newResponseFactory(method).FirstSingularResult()
	route := newRouteFactory(method).PerformAction(param, resp)
	c := &commenter{
		route:     route,
		parameter: param,
		response:  resp,
	}
	c.Do(sw)
}
