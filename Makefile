ROOT     := $(CURDIR)
ENVDSPEC := $(ROOT)/submodules/e2b-runtime/packages/envd/spec

# The generators live in ./tools, a separate module, so that buf and
# oapi-codegen never reach the dependency graph of anyone importing this SDK.
# They run with tools/ as the working directory, which is why every path handed
# to them below is absolute.
GOTOOL := cd $(ROOT)/tools && go tool

.PHONY: generate generate-volume-v2 build test lint tidy

## generate: regenerate API clients from the OpenAPI and proto specs.
## Requires the submodules: git submodule update --init
generate: generate-volume-v2
	$(GOTOOL) oapi-codegen -config $(ROOT)/spec/api.models.cfg.yaml \
	  -o $(ROOT)/pkg/api/types.gen.go $(ROOT)/spec/openapi.yml
	$(GOTOOL) oapi-codegen -config $(ROOT)/spec/api.client.cfg.yaml \
	  -o $(ROOT)/pkg/api/client.gen.go $(ROOT)/spec/openapi.yml
	$(GOTOOL) oapi-codegen -config $(ROOT)/spec/envd.cfg.yaml \
	  -o $(ROOT)/pkg/envd/api/api.gen.go $(ENVDSPEC)/envd.yaml
	$(GOTOOL) buf generate --template $(ROOT)/spec/buf.gen.yaml $(ENVDSPEC) \
	  -o $(ROOT)/pkg/envd \
	  --path $(ENVDSPEC)/process \
	  --path $(ENVDSPEC)/filesystem
	gofmt -w $(ROOT)/pkg/api $(ROOT)/pkg/envd

## generate-volume-v2: regenerate the independent /v2/volumes client.
generate-volume-v2:
	$(GOTOOL) oapi-codegen -config $(ROOT)/spec/volume.v2.cfg.yaml \
	  -o $(ROOT)/pkg/volume/v2/api/api.gen.go $(ROOT)/spec/openapi-volume.yaml

build:
	go build ./...

test:
	go test ./... -race

lint:
	golangci-lint run

tidy:
	go mod tidy
	cd tools && go mod tidy
