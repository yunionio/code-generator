package models

import (
	"reflect"
	"testing"

	"yunion.io/x/onecloud/pkg/cloudcommon/db"
	"yunion.io/x/onecloud/pkg/compute/models"
)

func init() {
	RegisterModelManager(models.GuestManager)
}

func TestGetModelManager(t *testing.T) {
	type args struct {
		typeName string
	}
	tests := []struct {
		name string
		args args
		want db.IModelManager
	}{
		{
			name: "getModelManager",
			args: args{"SGuestManager"},
			want: models.GuestManager,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetModelManager(tt.args.typeName); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetModelManager() = %v, want %v", got, tt.want)
			}
		})
	}
}
