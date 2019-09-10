# default build target
all::

all:: build
.PHONY: all

PROTOBUF_PATH:=../protobuf

build:

	protoc --proto_path=. \
				 --proto_path=$(PROTOBUF_PATH) \
				 --proto_path=$(PROTOBUF_PATH)/include \
				 --go_out=. \
				 *.proto

clean:

	rm *.pb.go
