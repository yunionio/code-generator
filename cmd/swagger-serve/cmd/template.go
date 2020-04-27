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

	RedocUIIndexTemplate = `<!DOCTYPE html>
<html>
  <head>
    <title>Swagger UI</title>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <style>
      body {
        margin: 0;
        padding-top: 40px;
      }

      nav {
        position: fixed;
        top: 0;
        width: 100%;
        z-index: 100;
      }
      #links_container {
          margin: 0;
          padding: 0;
          background-color: #0033a0;
      }

      #links_container li {
          display: inline-block;
          padding: 10px;
          color: white;
          cursor: pointer;
      }
    </style>
  </head>
  <body>

    <!-- Top navigation placeholder -->
    <nav>
      <ul id="links_container">
      </ul>
    </nav>

    <redoc scroll-y-offset="body > nav"></redoc>

    <script src="{{ .RedocURL }}"> </script>
    <script>
      // list of APIS
      var apis = [
        {{range .URLs}}
        {
          name: '{{.Name}}',
          url: './{{.Path}}'
        },
		{{end}}
      ];

      // initially render first API
      Redoc.init(apis[0].url);

      function onClick() {
        var url = this.getAttribute('data-link');
        Redoc.init(url);
      }

      // dynamically building navigation items
      var $list = document.getElementById('links_container');
      apis.forEach(function(api) {
        var $listitem = document.createElement('li');
        $listitem.setAttribute('data-link', api.url);
        $listitem.innerText = api.name;
        $listitem.addEventListener('click', onClick);
        $list.appendChild($listitem);
      });
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

type RedocUIIndexConfig struct {
	RedocURL string
	URLs     []*SwaggerFile
}

func (cfg RedocUIIndexConfig) Generate() ([]byte, error) {
	out := new(bytes.Buffer)
	t := template.Must(template.New("compiled_template").Parse(RedocUIIndexTemplate))
	if err := t.Execute(out, cfg); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}
