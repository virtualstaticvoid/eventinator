# default build target
all::

all:: build
.PHONY: all

build:

	docker build -t vsv/eventinator:latest .

build-certs:

	$(MAKE) -C certs build

PROTOBUF_PATH=protobuf

build-protofiles:

	protoc --proto_path=$(PROTOBUF_PATH) \
				 --proto_path=$(PROTOBUF_PATH)/include \
				 --go_out=$(PROTOBUF_PATH) \
				 eventinator.proto

	protoc --proto_path=$(PROTOBUF_PATH) \
				 --proto_path=$(PROTOBUF_PATH)/include \
				 --go_out=plugins=grpc:$(PROTOBUF_PATH) \
				 api.proto

	protoc --proto_path=$(PROTOBUF_PATH) \
				 --proto_path=$(PROTOBUF_PATH)/include \
				 --go_out=plugins=grpc:$(PROTOBUF_PATH) \
				 internal.proto

	protoc --proto_path=test \
				 --proto_path=$(PROTOBUF_PATH) \
				 --proto_path=$(PROTOBUF_PATH)/include \
				 --go_out=plugins=grpc:test \
				 examples.proto
