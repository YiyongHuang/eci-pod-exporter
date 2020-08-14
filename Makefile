REGISTRIES?=huangyiyong
APP=eci-pod-exporter
V=$(shell cat VERSION)


build:
	GOOS=linux GOARCH=amd64 go build -o deploy/bin/eci-pod-exporter main.go


image: build
	@docker build -f deploy/Dockerfile deploy -t $(REGISTRIES)/$(APP):$(V)
	@docker push $(REGISTRIES)/$(APP):$(V)
	@echo "$(REGISTRIES)/$(APP):$(V)"
