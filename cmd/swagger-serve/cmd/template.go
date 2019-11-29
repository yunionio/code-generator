package cmd

import (
	"bytes"
	"html/template"
)

const (
	UIIndexHTML = `
<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8">
    <title>Swagger UI</title>
    <link rel="stylesheet" type="text/css" href="{{.UICss}}" >
    <!--<link rel="icon" type="image/png" href="./favicon-32x32.png" sizes="32x32" />-->
    <!--<link rel="icon" type="image/png" href="./favicon-16x16.png" sizes="16x16" />-->
    <style>
      html
      {
        box-sizing: border-box;
        overflow: -moz-scrollbars-vertical;
        overflow-y: scroll;
      }

      *,
      *:before,
      *:after
      {
        box-sizing: inherit;
      }

      body
      {
        margin:0;
        background: #fafafa;
      }
    </style>
  </head>

  <body>
    <div id="swagger-ui"></div>

    <script src="{{.BundleJS}}"> </script>
    <script src="{{.StandalonePresetJS}}"> </script>
    <script>
    window.onload = function() {
      // Begin Swagger UI call region
      const ui = SwaggerUIBundle({
        urls: [
    {{range .URLs}}
          {url: "./{{.Path}}", name: "{{.Name}}"},
    {{end}}
        ],
        dom_id: '#swagger-ui',
        deepLinking: true,
        tagsSorter: "alpha",
        presets: [
          SwaggerUIBundle.presets.apis,
          SwaggerUIStandalonePreset
        ],
        plugins: [
          SwaggerUIBundle.plugins.DownloadUrl
        ],
        layout: "StandaloneLayout"
      })
      // End Swagger UI call region

      window.ui = ui
    }
  </script>
  </body>
</html>
`
)

type SwaggerFile struct {
	Path string
	Name string
}

type UIIndexHTMLConfig struct {
	UICss              string
	BundleJS           string
	StandalonePresetJS string
	URLs               []*SwaggerFile
}

func (cfg UIIndexHTMLConfig) Generate() ([]byte, error) {
	out := new(bytes.Buffer)
	t := template.Must(template.New("compiled_template").Parse(UIIndexHTML))
	if err := t.Execute(out, cfg); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}
