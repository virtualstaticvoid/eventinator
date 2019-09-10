# default build target
all::

all:: build
.PHONY: all

.PHONY: build
build: build-certs build-protobuf

	docker build --tag vsv/eventinator:latest \
		--build-arg GO_VERSION=1.12.9 \
		--build-arg DEBIAN_VERSION=stretch \
		.

.PHONY: build-certs
build-certs:

	$(MAKE) -C certs build

.PHONY: build-protobuf
build-protobuf:

	$(MAKE) -C protobuf build
	$(MAKE) -C test build

.PHONY: exportimage
exportimage:

	docker image save vsv/eventinator:latest --output vsv-eventinator-latest.tar.gz
