package generators

import (
	"reflect"
	"testing"

	"k8s.io/gengo/types"
)

func Test_extractOnlyForKeywords(t *testing.T) {
	tests := []struct {
		name     string
		comments []string
		want     []string
	}{
		{
			name: "single",
			comments: []string{
				"+onecloud:swagger-gen-only-for=server",
			},
			want: []string{"server"},
		},
		{
			name: "comma separated",
			comments: []string{
				"+onecloud:swagger-gen-only-for=server,disk,network",
			},
			want: []string{"disk", "network", "server"},
		},
		{
			name: "repeated tags",
			comments: []string{
				"+onecloud:swagger-gen-only-for=server",
				"+onecloud:swagger-gen-only-for=disk",
			},
			want: []string{"disk", "server"},
		},
		{
			name:     "empty",
			comments: []string{"+onecloud:swagger-gen-ignore"},
			want:     []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractOnlyForKeywords(tt.comments).List()
			if got == nil {
				got = []string{}
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("extractOnlyForKeywords() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func Test_skipMethodForResource(t *testing.T) {
	newMethod := func(comments []string) *types.Type {
		return &types.Type{CommentLines: comments}
	}
	tests := []struct {
		name     string
		comments []string
		keyword  string
		wantSkip bool
	}{
		{
			name:     "ignore always skips",
			comments: []string{"+onecloud:swagger-gen-ignore"},
			keyword:  "server",
			wantSkip: true,
		},
		{
			name:     "only-for match keeps",
			comments: []string{"+onecloud:swagger-gen-only-for=server,disk"},
			keyword:  "server",
			wantSkip: false,
		},
		{
			name:     "only-for miss skips",
			comments: []string{"+onecloud:swagger-gen-only-for=server,disk"},
			keyword:  "bucket",
			wantSkip: true,
		},
		{
			name: "only-for overrides ignore",
			comments: []string{
				"+onecloud:swagger-gen-ignore",
				"+onecloud:swagger-gen-only-for=server",
			},
			keyword:  "server",
			wantSkip: false,
		},
		{
			name:     "no tag keeps",
			comments: []string{"设置标签"},
			keyword:  "server",
			wantSkip: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := skipMethodForResource(newMethod(tt.comments), tt.keyword); got != tt.wantSkip {
				t.Errorf("skipMethodForResource() = %v, want %v", got, tt.wantSkip)
			}
		})
	}
}

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
