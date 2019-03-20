# you can specify this at runtime: make version=myvalue <command>
version := $(shell git log --pretty=format:'%h' -n 1)

.PHONY: build-client
build-client:
	@echo ">> building client"
	docker build -t grpcdemo-client -f client/Dockerfile .

.PHONY: build-server
build-server:
	@echo ">> building server"
	docker build -t grpcdemo-server -f server/Dockerfile .

.PHONY: push-client
push-client:
	@echo ">> tagging client ${version}"
	docker tag grpcdemo-client jwenz723/grpcdemo-client:${version}
	docker tag grpcdemo-client jwenz723/grpcdemo-client:latest

	@echo ">> pushing client ${version}"
	docker push jwenz723/grpcdemo-client:${version}
	docker push jwenz723/grpcdemo-client:latest

.PHONY: push-server
push-server:
	@echo ">> tagging server ${version}"
	docker tag grpcdemo-server jwenz723/grpcdemo-server:${version}
	docker tag grpcdemo-server jwenz723/grpcdemo-server:latest

	@echo ">> pushing server ${version}"
	docker push jwenz723/grpcdemo-server:${version}
	docker push jwenz723/grpcdemo-server:latest