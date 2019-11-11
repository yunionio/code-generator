model-api-gen:
	go build -o _output/bin/model-api-gen cmd/model-api-gen/main.go

swagger-gen:
	go build -o _output/bin/swagger-gen cmd/swagger-gen/main.go

install: model-api-gen swagger-gen
	cp -a _output/bin/* $$GOBIN

clean:
	rm -rf _output/bin/

#models-pkg-gen:
	#go build -o _output/bin/models-pkg-gen cmd/models-pkg-gen/main.go

#generate-pkg: models-pkg-gen
	#_output/bin/models-pkg-gen \
		#-i yunion.io/x/onecloud/pkg/compute/models \
		#-i yunion.io/x/onecloud/pkg/image/models \
		#-p yunion.io/x/code-generator/pkg/models

