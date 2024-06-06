# Переменные
DOCKER_IMAGE_NAME := backend-general
REMOTE_HOST := user1@91.224.86.182

# Путь к файлу с Docker-образом (на локальной машине)
DOCKER_IMAGE_FILE := backend-general.tar

# Путь на удаленной машине для загрузки Docker-образа
REMOTE_DIR := ~/docker_images

# Команда для сборки Docker-образа
build:
	@echo "\033[32m docker build... \033[0m"
	docker build -t $(DOCKER_IMAGE_NAME) . --platform linux/amd64
	
# Команда для сохранения Docker-образа в файл
save:
	@echo "\033[32m docker save image... \033[0m"
	docker save -o $(DOCKER_IMAGE_FILE) $(DOCKER_IMAGE_NAME)

# Команда для загрузки Docker-образа на удаленную машину
upload:
	@echo "\033[32m download image to server... \033[0m"
	scp $(DOCKER_IMAGE_FILE) $(REMOTE_HOST):$(REMOTE_DIR)

# Команда для удаленной загрузки и запуска Docker-образа
deploy: build save
	@echo "\033[32m docker remove old image... \033[0m"
	ssh $(REMOTE_HOST) "sudo docker rm -f $(DOCKER_IMAGE_NAME).tar || true"
	@echo "\033[32m docker load image... \033[0m"
	ssh $(REMOTE_HOST) "sudo docker load -i $(REMOTE_DIR)/$(DOCKER_IMAGE_FILE)"
	@echo "\033[32m docker run image... \033[0m"
	ssh $(REMOTE_HOST) "sudo docker run -d -p 8080:8080 $(DOCKER_IMAGE_NAME)"
