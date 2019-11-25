module yunion.io/x/code-generator

go 1.13

require (
	github.com/go-openapi/loads v0.19.4
	github.com/pkg/errors v0.8.1
	github.com/serialx/hashring v0.0.0-20190515033939-7706f26af194 // indirect
	github.com/skratchdot/open-golang v0.0.0-20190402232053-79abb63cd66e
	github.com/spf13/cobra v0.0.3
	github.com/spf13/pflag v1.0.5
	github.com/tredoe/osutil v0.0.0-20191018075336-e272fdda81c8 // indirect
	golang.org/x/tools v0.0.0-20191112005509-a3f652f18032 // indirect
	k8s.io/gengo v0.0.0-20191120174120-e74f70b9b27e
	k8s.io/klog v1.0.0
	yunion.io/x/jsonutils v0.0.0-20191005115334-bb1c187fc0e7
	yunion.io/x/log v0.0.0-20190629062853-9f6483a7103d
	yunion.io/x/onecloud v0.0.0-00010101000000-000000000000
	yunion.io/x/pkg v0.0.0-20191121110824-e03b47b93fe0
	yunion.io/x/sqlchemy v0.0.0-20191122085525-2d3bfdb3f51c // indirect
)

replace yunion.io/x/onecloud => ../onecloud
