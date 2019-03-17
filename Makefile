.PHONY: build-client
build-client:
	@echo ">> building client"
	docker build -t jwenz723/grpcdemo_client -f client/Dockerfile .

.PHONY: build-server
build-server:
	@echo ">> building server"
	docker build -t jwenz723/grpcdemo_server -f server/Dockerfile .

.PHONY: push-client
push-client:
	@echo ">> pushing client"
	docker push jwenz723/grpcdemo_client

.PHONY: push-server
push-server:
	@echo ">> pushing server"
	docker push jwenz723/grpcdemo_server