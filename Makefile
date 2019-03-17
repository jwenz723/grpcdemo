version := $(shell git log --pretty=format:'%h' -n 1)

.PHONY: build-client
build-client:
	@echo ">> building client"
	docker build -t jwenz723/grpcdemo-client -f client/Dockerfile .

.PHONY: build-server
build-server:
	@echo ">> building server"
	docker build -t jwenz723/grpcdemo-server -f server/Dockerfile .

.PHONY: push-client
push-client:
	@echo ">> pushing client ${version}"
	docker tag jwenz723/grpcdemo-client jwenz723/grpcdemo-client:${version}
	docker tag jwenz723/grpcdemo-client jwenz723/grpcdemo-client:latest
	docker push jwenz723/grpcdemo-client:${version}
	docker push jwenz723/grpcdemo-client:latest

.PHONY: push-server
push-server:
	@echo ">> pushing server ${version}"
    docker tag jwenz723/grpcdemo-server jwenz723/grpcdemo-server:${version}
    docker tag jwenz723/grpcdemo-server jwenz723/grpcdemo-server:latest
    docker push jwenz723/grpcdemo-server:${version}
    docker push jwenz723/grpcdemo-server:latest