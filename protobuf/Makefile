# default build target
all::

all:: build
.PHONY: all

build:

	protoc --proto_path=. \
				 --proto_path=include \
				 --go_out=plugins=grpc:. \
				 *.proto

clean:

	rm *.pb.go
