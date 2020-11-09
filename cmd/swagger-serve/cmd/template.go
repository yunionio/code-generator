package cmd

import (
	"bytes"
	"html/template"
	"sort"
	"strings"
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
      function insertParam(key, value) {
        key = encodeURIComponent(key);
        value = encodeURIComponent(value);

        // kvp looks like ['key1=value1', 'key2=value2', ...]
        var kvp = document.location.search.substr(1).split('&');
        let i=0;

        for(; i<kvp.length; i++){
            if (kvp[i].startsWith(key + '=')) {
                let pair = kvp[i].split('=');
                pair[1] = value;
                kvp[i] = pair.join('=');
                break;
            }
        }

        if(i >= kvp.length){
            kvp[kvp.length] = [key,value].join('=');
        }

        // can return this or...
        let params = kvp.join('&');

        // reload page with new params
        document.location.search = params;
      }

      const apiIdx = 'apiIdx';
      var urlParams = new URLSearchParams(window.location.search);
      var apiIdxVal = urlParams.get(apiIdx);
      if (apiIdxVal) {
        apiIdxVal = parseInt(apiIdxVal);
      } else {
        apiIdxVal = 0;
      }

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
      Redoc.init(apis[apiIdxVal].url);

      function onClick() {
        var url = this.getAttribute('data-link');
        Redoc.init(url);
      }

      // dynamically building navigation items
      var $list = document.getElementById('links_container');
      apis.forEach(function(api, idx) {
        var $listitem = document.createElement('li');
        $listitem.setAttribute('data-link', api.url);
        var tmpIdx = idx;
        $listitem.innerText = api.name;
        //$listitem.addEventListener('click', onClick);
        $listitem.addEventListener('click', function() {
          var url = this.getAttribute('data-link');
          insertParam(apiIdx, tmpIdx);
        });
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

type SwaggerFiles []*SwaggerFile

var _ sort.Interface = SwaggerFiles(make([]*SwaggerFile, 0))

func (sf SwaggerFiles) Len() int {
	return len(sf)
}

func (sf SwaggerFiles) Swap(i, j int) {
	sf[i], sf[j] = sf[j], sf[i]
}

func (sf SwaggerFiles) Less(i, j int) bool {
	item1 := sf[i]
	item2 := sf[j]
	return strings.Compare(item1.Name, item2.Name) < 0
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
