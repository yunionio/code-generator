package generators

import (
	"k8s.io/gengo/generator"
)

const swaggerMeta = `
// Documentation of OneCloud API
//
//     Schemes: https, http
//     BasePath: /
//     Version: 1.0
//     Host: "10.168.222.136:8889"
//     Contact: Zexi Li<lizexi@yunion.cn>
//     License: Apache 2.0 http://www.apache.org/licenses/LICENSE-2.0.html
//
//     Consumes:
//     - application/json
//
//     Produces:
//     - application/json
//
//     SecurityDefinitions:
//     keystone:
//       name: X-Auth-Token
//       type: apiKey
//       in: header
//
// swagger:meta
`

type swaggerDocGen struct {
	generator.DefaultGen
}

func NewSwaggerDocGen() generator.Generator {
	return &swaggerDocGen{
		DefaultGen: generator.DefaultGen{
			OptionalName: "doc",
			OptionalBody: []byte(swaggerMeta),
		},
	}
}
