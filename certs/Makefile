# default build target
all::

all:: build
.PHONY: all

build:

	@cp --no-clobber password.example password
	@if [ ! -f ca.crt ]; then ./generate-ca; fi
	@if [ ! -f server-cert.crt ]; then ./generate-server-cert; fi
	@if [ ! -f client-cert.crt ]; then ./generate-client-cert; fi

rebuild:

	@rm *.crt
	@rm *.key
	@rm *.csr
	@rm *.pem
	@rm *.txt

	make build
