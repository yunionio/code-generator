package models

import (
	"fmt"
	"reflect"

	"k8s.io/gengo/types"

	"yunion.io/x/onecloud/pkg/appsrv"
	"yunion.io/x/onecloud/pkg/cloudcommon/db"
	computesvc "yunion.io/x/onecloud/pkg/compute/service"
	imagesvc "yunion.io/x/onecloud/pkg/image/service"
	identitysvc "yunion.io/x/onecloud/pkg/keystone/service"
)

func init() {
	cleanF := func(man map[string]db.IModelManager) {
		for key, _ := range man {
			delete(man, key)
		}
	}
	registerF := func(man map[string]db.IModelManager) {
		for _, man := range db.GlobalModelManagerTables() {
			RegisterModelManager(man)
		}
	}
	for _, f := range []func(*appsrv.Application){
		computesvc.InitHandlers,
		imagesvc.InitHandlers,
		identitysvc.InitHandlers,
	} {
		app := appsrv.NewApplication("", 1, false)
		f(app)
		registerF(db.GlobalModelManagerTables())
		// hack: clean all model manager to avoid duplicate registered
		cleanF(db.GlobalModelManagerTables())
	}
}

var globalManagers map[string]db.IModelManager

func GlobalManagers() map[string]db.IModelManager {
	return globalManagers
}

func GetModelManagerKey(man db.IModelManager) string {
	manType := reflect.TypeOf(man)
	if manType.Kind() == reflect.Ptr {
		manType = manType.Elem()
	}
	return fmt.Sprintf("%s.%s", manType.PkgPath(), manType.Name())
}

func RegisterModelManager(man db.IModelManager) {
	if globalManagers == nil {
		globalManagers = make(map[string]db.IModelManager)
	}
	globalManagers[GetModelManagerKey(man)] = man
}

func GetModelManager(typeName string) db.IModelManager {
	return globalManagers[typeName]
}

func GetModelManagerByType(t *types.Type) db.IModelManager {
	if t == nil {
		return nil
	}
	// t.String() is pkgPath.typeName, e.g:yunion.io/x/onecloud/pkg/keystone/models.SAssignmentManager
	return GetModelManager(t.String())
}
