# Makefile

# Nome da imagem Docker
IMAGE_NAME := italobbarros/rinha-backend-2024-q1

# Nome do contêiner Docker
CONTAINER_NAME := italobbarros/rinha-backend-2024-q1

# Diretório onde está o Dockerfile
DOCKER_DIR := ./docker

# Comando para construir a imagem Docker
DOCKER_BUILD_CMD := docker-compose build --build-arg APP_VERSION=$(VERSION)

# Comando para iniciar o Docker Compose
DOCKER_COMPOSE_UP_CMD := docker-compose -f docker-compose-prod.yml up -d

# Comando para executar o script de teste
TEST_SCRIPT_CMD := sh test.sh

# Comando para parar e remover os contêineres Docker
DOCKER_COMPOSE_DOWN_CMD := docker-compose -f docker-compose-prod.yml down

# Alvo padrão (executado ao chamar apenas 'make')
all: build up test

up-test: up test
# Alvo para construir a imagem Docker
build:
	@echo "Construindo a imagem Docker com a versão $(VERSION)..."
	$(DOCKER_BUILD_CMD)

# Alvo para iniciar o Docker Compose
up:
	@echo "Iniciando o Docker Compose..."
	$(DOCKER_COMPOSE_UP_CMD)

# Alvo para executar o script de teste
test:
	@echo "Executando o script de teste..."
	$(TEST_SCRIPT_CMD)

# Alvo para parar e remover os contêineres Docker
down:
	@echo "Parando e removendo os contêineres Docker..."
	$(DOCKER_COMPOSE_DOWN_CMD)

# Alvo para construir, iniciar e testar
build-test: build up test down

# Alvo para limpar a construção e os contêineres
clean:
	@echo "Limpando a construção e os contêineres Docker..."
	docker-compose down --volumes --remove-orphans
