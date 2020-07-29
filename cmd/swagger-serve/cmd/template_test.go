package cmd

import (
	"reflect"
	"sort"
	"testing"
)

func TestSwaggerFilesSort(t *testing.T) {
	input := []*SwaggerFile{
		{
			Name: "monitor",
		},
		{
			Name: "compute",
		},
		{
			Name: "identity",
		},
	}
	sort.Sort(SwaggerFiles(input))
	sortNames := make([]string, 0)
	for _, n := range input {
		sortNames = append(sortNames, n.Name)
	}
	expected := []string{"compute", "identity", "monitor"}
	if !reflect.DeepEqual(sortNames, expected) {
		t.Errorf("sort names %v != %v", sortNames, expected)
	}
}
