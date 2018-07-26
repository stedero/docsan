include /usr/local/IBFD/etc/docker.env

TAG=$(shell date +%Y-%m-%d-%H%M)
VERSION=1.0
IMAGE=ibfd/docsan

GOLANG=golang:1.8
RUNOPTS=--rm --network ibfd_exposed -v ${HOME}/.ssh:/root/.ssh -v "${PWD}":/usr/local/go/src/ibfd.org/docsan -w /usr/local/go/src/ibfd.org/docsan --dns-search ${DNS_SEARCH} --dns ${DNS1_IP} --dns ${DNS2_IP}

clean:
	rm -f main

version:
	@echo "package main" > version.go
	@echo "const version = \"${VERSION}\";" >> version.go

ci-build:
	docker run ${RUNOPTS} ${GOLANG} ./alpine_build.sh
	cp docsan docker/docsan
	docker build docker/ -t ${REGISTRY}/${IMAGE}:latest
	docker tag ${REGISTRY}/${IMAGE}:latest ${REGISTRY}/${IMAGE}:${TAG}
	docker tag ${REGISTRY}/${IMAGE}:latest ${REGISTRY}/${IMAGE}:${VERSION}
	docker push ${REGISTRY}/${IMAGE}:latest
	docker push ${REGISTRY}/${IMAGE}:${TAG}
	docker push ${REGISTRY}/${IMAGE}:${VERSION}
	tar czvf package.tgz docker

goshell:
	docker run -ti ${RUNOPTS} ${GOLANG} /bin/bash
