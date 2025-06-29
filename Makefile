LOCAL_BIN := $(CURDIR)/bin

PROTOC = PATH="$$PATH:$(LOCAL_BIN)" protoc

PROTO_PATH := $(CURDIR)/api

PKG_PROTO_PATH := $(CURDIR)/pkg

VENDOR_PROTO_PATH := $(CURDIR)/vendor.protobuf

MIGRATION_DIR :=./migrations

.bin-deps: export GOBIN := $(LOCAL_BIN)
.bin-deps:
	$(info Installing binary dependencies...)

	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest


vendor:	.vendor-reset .vendor-googleapis .vendor-google-protobuf .vendor-protovalidate .vendor-protoc-gen-openapiv2 .vendor-tidy

.vendor-reset:
	rm -rf $(VENDOR_PROTO_PATH)
	mkdir -p $(VENDOR_PROTO_PATH)

.vendor-tidy:
	find $(VENDOR_PROTO_PATH) -type f ! -name "*.proto" -delete
	find $(VENDOR_PROTO_PATH) -empty -type d -delete

.vendor-google-protobuf:
	git clone -b main --single-branch -n --depth=1 --filter=tree:0 \
		https://github.com/protocolbuffers/protobuf $(VENDOR_PROTO_PATH)/protobuf &&\
	cd $(VENDOR_PROTO_PATH)/protobuf &&\
	git sparse-checkout set --no-cone src/google/protobuf &&\
	git checkout
	mkdir -p $(VENDOR_PROTO_PATH)/google
	mv $(VENDOR_PROTO_PATH)/protobuf/src/google/protobuf $(VENDOR_PROTO_PATH)/google
	rm -rf $(VENDOR_PROTO_PATH)/protobuf

.vendor-protovalidate:
	git clone -b main --single-branch --depth=1 --filter=tree:0 \
		https://github.com/bufbuild/protovalidate $(VENDOR_PROTO_PATH)/protovalidate && \
	cd $(VENDOR_PROTO_PATH)/protovalidate
	git checkout
	mv $(VENDOR_PROTO_PATH)/protovalidate/proto/protovalidate/buf $(VENDOR_PROTO_PATH)
	rm -rf $(VENDOR_PROTO_PATH)/protovalidate

.vendor-googleapis:
	git clone -b master --single-branch -n --depth=1 --filter=tree:0 \
		https://github.com/googleapis/googleapis $(VENDOR_PROTO_PATH)/googleapis &&\
	cd $(VENDOR_PROTO_PATH)/googleapis &&\
	git checkout
	mv $(VENDOR_PROTO_PATH)/googleapis/google $(VENDOR_PROTO_PATH)
	rm -rf $(VENDOR_PROTO_PATH)/googleapis

.vendor-protoc-gen-openapiv2:
	git clone -b main --single-branch -n --depth=1 --filter=tree:0 \
 		https://github.com/grpc-ecosystem/grpc-gateway $(VENDOR_PROTO_PATH)/grpc-gateway && \
 	cd $(VENDOR_PROTO_PATH)/grpc-gateway && \
	git sparse-checkout set --no-cone protoc-gen-openapiv2/options && \
	git checkout
	mkdir -p $(VENDOR_PROTO_PATH)/protoc-gen-openapiv2
	mv $(VENDOR_PROTO_PATH)/grpc-gateway/protoc-gen-openapiv2/options $(VENDOR_PROTO_PATH)/protoc-gen-openapiv2
	rm -rf $(VENDOR_PROTO_PATH)/grpc-gateway

.protoc-generate:
	mkdir -p $(PKG_PROTO_PATH)
	$(PROTOC) -I $(VENDOR_PROTO_PATH) --proto_path=$(CURDIR) \
	--go_out=$(PKG_PROTO_PATH) --go_opt paths=source_relative \
	--go-grpc_out=$(PKG_PROTO_PATH) --go-grpc_opt paths=source_relative \
	--grpc-gateway_out=$(PKG_PROTO_PATH) --grpc-gateway_opt paths=source_relative --grpc-gateway_opt generate_unbound_methods=true \
	$(PROTO_PATH)/orders_management_system/messages.proto $(PROTO_PATH)/orders_management_system/service.proto
	
	$(PROTOC) -I $(VENDOR_PROTO_PATH) --proto_path=$(CURDIR) \
	--openapiv2_out=. --openapiv2_opt logtostderr=true \
	$(PROTO_PATH)/orders_management_system/service.proto


.tidy:
	GOBIN=$(LOCAL_BIN) go mod tidy

generate: .bin-deps .protoc-generate .tidy

build:
	go build -o $(LOCAL_BIN) ./cmd/orders_management_system


.install-migrate: export GOBIN := $(LOCAL_BIN)
.install-migrate: 
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

create-migartion: .install-migrate
create-migartion:
	PATH=$(LOCAL_BIN) migrate create -ext sql -dir "${MIGRATION_DIR}" -seq migration


DB_DSN:=postgresql://user:password@0.0.0.0:6532/orders_management_system?sslmode=disable

migrate:
	PATH=$(LOCAL_BIN) migrate -path ${MIGRATION_DIR} -database "${DB_DSN}" -verbose up

.PHONY: \
	.bin-deps \
	.protoc-generate \
	.tidy \
	.vendor-protovalidate \
	.vendor-protoc-gen-openapiv2 \
	.vendor-googleapis \
	.vendor-google-protobuf \
	vendor \
	generate \
	build \
	migrate \
	create-migartion
