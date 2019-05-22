TAG?=0.13.0

build:
	go mod vendor
	docker build --build-arg http_proxy=$http_proxy --build-arg https_proxy=$https_proxy -t openfaas-incubator/faas-rancher:$(TAG) .

push:
	docker push openfaas-incubator/faas-rancher:$(TAG)
