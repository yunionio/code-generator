package generators

import (
	"reflect"
	"testing"
)

func Test_extractSwaggerRoute(t *testing.T) {
	tests := []struct {
		name     string
		comments []string
		want     *SwaggerConfigRoute
	}{
		{
			name: "normal input",
			comments: []string{
				"+onecloud:swagger-gen-route-method=GET",
				"+onecloud:swagger-gen-route-path=/v2.0/tokens",
				"+onecloud:swagger-gen-route-tag=tag1",
				"+onecloud:swagger-gen-route-tag=tag2",
			},
			want: &SwaggerConfigRoute{
				Method: "GET",
				Path:   "/v2.0/tokens",
				Tags:   []string{"tag1", "tag2"},
			},
		},
		{
			name: "no method input",
			comments: []string{
				"+onecloud:swagger-gen-route-path=/v2.0/tokens",
				"+onecloud:swagger-gen-route-tag=tag1",
				"+onecloud:swagger-gen-route-tag=tag2",
			},
			want: nil,
		},
		{
			name: "no path input",
			comments: []string{
				"+onecloud:swagger-gen-route-method=GET",
				"+onecloud:swagger-gen-route-tag=tag1",
			},
			want: nil,
		},
		{
			name: "no tag input",
			comments: []string{
				"+onecloud:swagger-gen-route-method=GET",
				"+onecloud:swagger-gen-route-path=/v2.0/tokens",
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := extractSwaggerRoute(tt.comments); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("extractSwaggerTags() = %v, want %v", got, tt.want)
			}
		})
	}
}
