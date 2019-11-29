package report

import (
	"yunion.io/x/onecloud/pkg/cloudcommon/db"

	"yunion.io/x/code-generator/pkg/models"
)

type KindManager struct {
	modelManagers map[string]db.IModelManager
	resources     []*Resource
}

func NewKindManager() *KindManager {
	return &KindManager{
		modelManagers: models.GlobalManagers(),
	}
}

func (m *KindManager) initData() {
	for _, modelMan := range m.modelManagers {
	}
}
