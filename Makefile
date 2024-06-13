# Переменные
DOCKER_IMAGE_NAME := backend-general
REMOTE_HOST := user1@91.224.86.182

# Путь к файлу с Docker-образом (на локальной машине)
DOCKER_IMAGE_FILE := backend-general.tar

# Путь на удаленной машине для загрузки Docker-образа
REMOTE_DIR := ~/docker_images

# Команда для запуска проекта
run-local:
	@echo "\033[32m local run project... \033[0m"
	AGORA_ADDR="http://localhost:8081" go run main.go

docker-local:
	@echo "\033[32m local docker compose... \033[0m"
	docker-compose -f docker-compose.local.yaml up

# Команда для сборки Docker-образа
build:
	@echo "\033[32m docker build... \033[0m"
	docker build -t $(DOCKER_IMAGE_NAME) . --platform linux/amd64
	
# Команда для сохранения Docker-образа в файл
save:
	@echo "\033[32m docker save image... \033[0m"
	docker save -o out/$(DOCKER_IMAGE_FILE) $(DOCKER_IMAGE_NAME)

# Команда для загрузки Docker-образа на удаленную машину
upload-images:
	@echo "\033[32m download image to server... \033[0m"
	scp out/$(DOCKER_IMAGE_FILE) $(REMOTE_HOST):$(REMOTE_DIR)

# Команда для удаленной загрузки и запуска Docker-образа
deploy: build save upload-images
	@echo "\033[32m stop docker compose... \033[0m"
	ssh $(REMOTE_HOST) "cd /home/user1/docker_images/; sudo docker-compose down || true"

	@echo "\033[32m docker remove old image... \033[0m"
	ssh $(REMOTE_HOST) "sudo docker rmi -f $(DOCKER_IMAGE_NAME) || true"

	@echo "\033[32m remove unused container... \033[0m"
	ssh $(REMOTE_HOST) "sudo docker container prune -f || true"

	@echo "\033[32m remove unused images... \033[0m"
	ssh $(REMOTE_HOST) "sudo docker image prune -f || true"

	@echo "\033[32m copy docker compose file... \033[0m"
	scp docker-compose.yaml $(REMOTE_HOST):/home/user1/docker_images/docker-compose.yaml

	@echo "\033[32m docker load image... \033[0m"
	ssh $(REMOTE_HOST) "sudo docker load -i $(REMOTE_DIR)/$(DOCKER_IMAGE_FILE)"

	@echo "\033[32m start docker compose... \033[0m"
	ssh $(REMOTE_HOST) "cd /home/user1/docker_images/; sudo docker-compose up -d || true"