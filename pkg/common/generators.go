package common

import (
	"io"
	"strings"

	"k8s.io/gengo/generator"
	"k8s.io/gengo/types"

	"yunion.io/x/pkg/util/sets"
)

func EndWithResourceBase(t *types.Type) bool {
	resBases := []string{
		"ResourceBase",
		"JointsBase",
		"SharableBaseResource",
		"IdentityBaseResource",
	}
	for _, rb := range resBases {
		if strings.HasSuffix(t.Name.Name, rb) {
			return true
		}
	}
	return false
}

func IsResourceModel(t *types.Type, isCommonDBPkg bool) bool {
	endWithResBase := EndWithResourceBase(t)
	if isCommonDBPkg && endWithResBase {
		return true
	} else if endWithResBase {
		// service models pkg not generate cloudcommon/db models
		return false
	}

	for _, m := range t.Members {
		if EndWithResourceBase(m.Type) {
			return true
		}
	}
	return false
}

func InSourcePackage(t *types.Type, srcPkg string) bool {
	return IsSamePackage(t, srcPkg)
}

func IsSamePackage(t *types.Type, pkgPath string) bool {
	return t.Name.Package == pkgPath
}

func CollectModelManager(srcPkg string, pkgTypes []*types.Type, modelTypes sets.String, modelManagers map[string]*types.Type) {
	restTypes := make([]*types.Type, 0)
	for _, t := range pkgTypes {
		if t.Kind != types.Struct {
			continue
		}
		if !InSourcePackage(t, srcPkg) {
			continue
		}
		if IsResourceModel(t, false) {
			modelTypes.Insert(t.String())
		} else {
			restTypes = append(restTypes, t)
		}
	}
	for _, t := range restTypes {
		if strings.HasSuffix(t.Name.Name, "Manager") {
			modelName := strings.TrimSuffix(t.String(), "Manager")
			if modelTypes.Has(modelName) {
				modelManagers[modelName] = t
			}
		}
	}
}

func GetArgs(t *types.Type) interface{} {
	return generator.Args{
		"type": t,
	}
}

func NewSnippetWriter(w io.Writer, c *generator.Context) *generator.SnippetWriter {
	return generator.NewSnippetWriter(w, c, "$", "$")
}
