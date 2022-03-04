.PHONY: default build install uninstall snapshot publish clean

NAME = aws-extras
VERSION = 0.0.0
HOSTNAME = registry.terraform.io
NAMESPACE = drdaeman
OS_ARCH = darwin_amd64
BINARY = terraform-provider-${NAME}

default: install

build:
	go build -trimpath -o ${BINARY}

install: build
	mkdir -p ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}
	cp ${BINARY} ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}

uninstall:
	rm -r ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}

snapshot:
	goreleaser release --rm-dist --snapshot --skip-publish --skip-sign

publish:
	goreleaser release --rm-dist

clean:
	test -e ${BINARY} && rm ${BINARY} || true
	test -e dist && rm -rf dist || true
