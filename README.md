# code-generator for OneCloud

## Purpose

These code-generators can be used:

- [cmd/model-api-gen](./cmd/model-api-gen): generate and copy api models definition to package according by models.
- [cmd/swagger-gen](./cmd/swagger-gen): generate [go-swagger spec](https://goswagger.io/generate/spec.html) by parsing models.

## Install

```bash
$ git clone https://github.com/yunionio/code-generator $GOPATH/src/yunion.io/x/code-generator
$ cd $GOPATH/src/yunion.io/x/code-generator
$ make install
```

## Usage

### Simple test

```bash
# test model-api-gen
$ ./hack/model-generate.sh

# test swagger-gen
$ ./hack/swagger-generate.sh
```

### For onecloud project

Suppose you already clone https://github.com/yunionio/onecloud at **$GOPATH/src/yunion.io/x/onecloud**.

```bash
$ cd $GOPATH/src/yunion.io/x/onecloud

# generate models definition at apis package
$ make gen-model-api

# generate swagger spec
$ make gen-swagger
# view swagger web page
$ make swagger-serve
```
