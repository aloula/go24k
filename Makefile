# Makefile para Go24K
# ===================
# Comandos de desenvolvimento e build

.PHONY: all build test clean install help dev lint bench integration coverage

# Configurações
BINARY_NAME=go24k
GO_FILES=$(shell find . -name '*.go' -type f -not -path './vendor/*')
BUILD_DIR=builds

# Comando padrão
all: clean lint test build

# Compilar para plataforma atual
build:
	@echo "🏗️  Compilando Go24K..."
	go build -ldflags="-s -w" -o $(BINARY_NAME) .
	@echo "✅ Binário criado: $(BINARY_NAME)"

# Compilar para todas as plataformas
build-all:
	@echo "🏗️  Compilando para todas as plataformas..."
	./build.sh

# Executar todos os testes
test:
	@echo "🧪 Executando todos os testes..."
	./test.sh

# Testes unitários apenas
test-unit:
	@echo "🧪 Executando testes unitários..."
	./test.sh unit

# Testes de integração apenas
test-integration:
	@echo "🧪 Executando testes de integração..."
	./test.sh integration

# Benchmarks
bench:
	@echo "⚡ Executando benchmarks..."
	./test.sh bench

# Cobertura de testes
coverage:
	@echo "📊 Gerando relatório de cobertura..."
	go test -coverprofile=coverage.out ./utils/
	go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Relatório salvo em coverage.html"

# Análise estática
lint:
	@echo "🔍 Executando golangci-lint..."
	golangci-lint run --timeout 2m

# Análise com golangci-lint moderno (revive + stylecheck)
lint-modern:
	@echo "🔍 Executando golangci-lint moderno..."
	golangci-lint run --timeout 2m
	@echo "✅ Análise moderna concluída"

# Instalar dependências de desenvolvimento
dev-deps:
	@echo "📦 Instalando dependências de desenvolvimento..."
	go install golang.org/x/lint/golint@latest
	go install golang.org/x/tools/cmd/goimports@latest
	@echo "✅ Dependências instaladas"

# Formatar código
fmt:
	@echo "🎨 Formatando código..."
	go fmt ./...
	goimports -w $(GO_FILES) 2>/dev/null || true
	@echo "✅ Código formatado"

# Verificar dependências
mod:
	@echo "📦 Verificando módulo Go..."
	go mod verify
	go mod tidy
	@echo "✅ Módulo verificado"

# Limpar arquivos gerados
clean:
	@echo "🧹 Limpando arquivos..."
	rm -f $(BINARY_NAME) $(BINARY_NAME).exe
	rm -f go24k-* go24k_test_*
	rm -f coverage.out coverage.html
	rm -rf $(BUILD_DIR)
	find . -name "*_test_*" -type d -exec rm -rf {} + 2>/dev/null || true
	@echo "✅ Limpeza concluída"

# Instalar no sistema
install: build
	@echo "📦 Instalando Go24K..."
	sudo cp $(BINARY_NAME) /usr/local/bin/
	@echo "✅ Go24K instalado em /usr/local/bin/"

# Desinstalar do sistema
uninstall:
	@echo "🗑️  Desinstalando Go24K..."
	sudo rm -f /usr/local/bin/$(BINARY_NAME)
	@echo "✅ Go24K removido do sistema"

# Executar em modo de desenvolvimento
dev: fmt lint-modern test-unit
	@echo "🚀 Modo desenvolvimento ativo"

# Release - preparar para release
release: clean lint-modern test build-all
	@echo "🎉 Release preparado!"
	@echo "📦 Binários disponíveis em: $(BUILD_DIR)/"

# Verificação rápida antes de commit
check: fmt lint-modern test-unit
	@echo "✅ Verificações pré-commit concluídas"

# Docker build (se houver Dockerfile no futuro)
docker-build:
	@echo "🐳 Docker build não implementado ainda"

# Mostrar ajuda
help:
	@echo "Go24K - Comandos Disponíveis"
	@echo "============================"
	@echo ""
	@echo "Principais:"
	@echo "  make build         Compilar para plataforma atual"
	@echo "  make build-all     Compilar para todas as plataformas"
	@echo "  make test          Executar todos os testes"
	@echo "  make clean         Limpar arquivos gerados"
	@echo ""
	@echo "Desenvolvimento:"
	@echo "  make dev           Modo desenvolvimento (fmt + lint-modern + test-unit)"
	@echo "  make check         Verificação pré-commit (fmt + lint-modern + test-unit)"
	@echo "  make fmt           Formatar código"
	@echo "  make lint          Análise estática (tradicional)"
	@echo "  make lint-modern   Análise estática (golangci-lint)"
	@echo "  make coverage      Gerar relatório de cobertura"
	@echo ""
	@echo "Testes:"
	@echo "  make test-unit     Apenas testes unitários"
	@echo "  make test-integration  Apenas testes de integração"
	@echo "  make bench         Benchmarks de performance"
	@echo ""
	@echo "Sistema:"
	@echo "  make install       Instalar no sistema"
	@echo "  make uninstall     Remover do sistema"
	@echo ""
	@echo "Outros:"
	@echo "  make dev-deps      Instalar dependências de desenvolvimento"
	@echo "  make mod           Verificar módulo Go"
	@echo "  make release       Preparar release completo"
	@echo "  make help          Mostrar esta ajuda"
	@echo ""
	@echo "Exemplos:"
	@echo "  make              # build padrão"
	@echo "  make dev          # desenvolvimento"
	@echo "  make release      # preparar release"