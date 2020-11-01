pkger:
ifeq (, $(shell which pkger))
	go install github.com/markbates/pkger/cmd/pkger
endif
	pkger

protoc:
ifeq (, $(shell which protoc))
	$(error "No protoc in $(PATH), consider installing it from https://github.com/protocolbuffers/protobuf#protocol-compiler-installation")
endif
ifeq (, $(shell which protoc-gen-go))
	go install google.golang.org/protobuf/cmd/protoc-gen-go
endif
ifeq (, $(shell which protoc-gen-go-grpc))
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc
endif
ifeq (, $(shell [ -d 'api/grpc' ] && echo 1))
	wget -q https://github.com/grpc/grpc-proto/archive/master.zip -O $${TMPDIR}/grpc-proto.zip
	unzip -q -o $${TMPDIR}/grpc-proto.zip -d $${TMPDIR}
	mv $${TMPDIR}/grpc-proto-master/grpc api/grpc
endif
ifeq (, $(shell [ -d 'api/google' ] && echo 1))
	wget -q https://github.com/googleapis/api-common-protos/archive/1.50.0.zip -O $${TMPDIR}/api-common-protos.zip
	unzip -q -o $${TMPDIR}/api-common-protos.zip -d $${TMPDIR}
	mv $${TMPDIR}/api-common-protos-*/google api/
endif
	protoc --go_out=api --go_opt paths=source_relative --go-grpc_out=api --go-grpc_opt paths=source_relative --proto_path=api api/api.proto api/grpc/health/v1/health.proto

grpc-gateway:
ifeq (, $(shell which protoc-gen-grpc-gateway))
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway
endif
ifeq (, $(shell which protoc-gen-openapiv2))
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2
endif
	protoc --grpc-gateway_out api --grpc-gateway_opt logtostderr=true --grpc-gateway_opt paths=source_relative --proto_path=api api/api.proto
	protoc --grpc-gateway_out api --grpc-gateway_opt logtostderr=true --grpc-gateway_opt paths=source_relative --grpc-gateway_opt grpc_api_configuration=api/health.yaml --proto_path=api api/grpc/health/v1/health.proto
	protoc --openapiv2_out api --openapiv2_opt logtostderr=true --proto_path=api api/api.proto
	cat api/api.swagger.json | yq r -P - > api/openapi.yaml

generate: protoc grpc-gateway pkger

buf:
ifeq (, $(shell which buf))
	$(error "No buf in $(PATH), consider installing it from https://docs.buf.build/installation")
endif
	cd api && buf check lint --file api.proto

golangci-lint:
ifeq (, $(shell which golangci-lint))
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin v1.31.0
endif
	golangci-lint run

spectral:
ifeq (, $(shell which spectral))
	curl -L https://raw.githack.com/stoplightio/spectral/master/scripts/install.sh | sh
endif
	cd api && spectral lint -F warn openapi.yaml

lint: buf golangci-lint spectral

test:
	go test -v ./internal/...

verify: lint test

build: generate verify
	go build -o bin/app

mod:
	go mod tidy
	go mod verify
