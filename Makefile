model-api-gen:
	go build -o _output/bin/model-api-gen cmd/model-api-gen/main.go

swagger-gen:
	go build -o _output/bin/swagger-gen cmd/swagger-gen/main.go

swagger-serve:
	go build -o _output/bin/swagger-serve cmd/swagger-serve/main.go

install: model-api-gen swagger-gen swagger-serve
	rsync -avP _output/bin/* $$GOBIN

clean:
	rm -rf _output/bin/
