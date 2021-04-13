BINARY_NAME = faas-idler
IMAGE = smvfal/faas-idler
TAG = latest

publish: docker-build docker-push

docker-build:
	DOCKER_BUILDKIT=1 docker build -t ${IMAGE}:${TAG} .

docker-push:
	docker push ${IMAGE}:${TAG}

install:
	kubectl apply -f kubernetes/rbac.yml
	kubectl apply -f kubernetes/deployment.yml

vendor:
	go mod vendor -v

build:
	go build -o bin/${BINARY_NAME}

fmt:
	gofmt -l -d $(shell find . -type f -name '*.go' -not -path "./vendor/*")

clean:
	rm -rf vendor/ bin/ go.sum
	go clean -r -modcache
