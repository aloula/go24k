#!/bin/bash

# Script de Testes para Go24K
# ===========================
# Executa todos os tipos de testes e gera relatórios

set -e  # Para na primeira falha

# Cores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Função para imprimir headers
print_header() {
    echo ""
    echo -e "${BLUE}================================================${NC}"
    echo -e "${BLUE} $1${NC}"
    echo -e "${BLUE}================================================${NC}"
    echo ""
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

# Limpar arquivos temporários de testes anteriores
cleanup() {
    echo "🧹 Limpando arquivos temporários de testes..."
    find . -name "*_test_*" -type d -exec rm -rf {} + 2>/dev/null || true
    rm -f coverage.out coverage.html 2>/dev/null || true
    rm -f go24k_test* 2>/dev/null || true
    echo "✨ Limpeza concluída!"
}

# Verificar se Go está instalado
check_go() {
    if ! command -v go &> /dev/null; then
        print_error "Go não está instalado ou não está no PATH"
        exit 1
    fi
    
    GO_VERSION=$(go version | awk '{print $3}')
    print_success "Go detectado: $GO_VERSION"
}

# Executar testes unitários
run_unit_tests() {
    print_header "TESTES UNITÁRIOS"
    
    echo "📊 Executando testes unitários com cobertura..."
    if go test -v -coverprofile=coverage.out ./utils/; then
        print_success "Todos os testes unitários passaram!"
        
        # Gerar relatório de cobertura
        echo ""
        echo "📈 Relatório de Cobertura:"
        go tool cover -func=coverage.out
        
        # Gerar HTML da cobertura (opcional)
        if command -v open &> /dev/null || command -v xdg-open &> /dev/null; then
            echo ""
            echo "🌐 Gerando relatório HTML de cobertura..."
            go tool cover -html=coverage.out -o coverage.html
            print_success "Relatório HTML salvo em: coverage.html"
        fi
    else
        print_error "Falha nos testes unitários!"
        return 1
    fi
}

# Executar testes da função main
run_main_tests() {
    print_header "TESTES DA FUNÇÃO MAIN"
    
    echo "🎯 Executando testes da função main..."
    if go test -v .; then
        print_success "Todos os testes da main passaram!"
    else
        print_error "Falha nos testes da main!"
        return 1
    fi
}

# Executar testes de integração
run_integration_tests() {
    print_header "TESTES DE INTEGRAÇÃO"
    
    # Verificar se FFmpeg está disponível
    if ! command -v ffmpeg &> /dev/null; then
        print_warning "FFmpeg não encontrado - pulando testes de integração"
        print_warning "Instale FFmpeg para executar testes completos"
        return 0
    fi
    
    echo "🔗 Executando testes de integração..."
    if go test -tags=integration -v .; then
        print_success "Todos os testes de integração passaram!"
    else
        print_error "Falha nos testes de integração!"
        return 1
    fi
}

# Executar benchmarks
run_benchmarks() {
    print_header "BENCHMARKS"
    
    echo "⚡ Executando benchmarks de performance..."
    if go test -bench=. -benchmem ./utils/; then
        print_success "Benchmarks concluídos!"
    else
        print_warning "Problemas nos benchmarks (não crítico)"
    fi
}

# Análise estática do código
run_static_analysis() {
    print_header "ANÁLISE ESTÁTICA"
    
    echo "🔍 Executando go vet..."
    if go vet ./...; then
        print_success "go vet: nenhum problema encontrado"
    else
        print_error "go vet encontrou problemas!"
        return 1
    fi
    
    echo ""
    echo "🔍 Executando go fmt..."
    UNFORMATTED=$(go fmt ./...)
    if [ -z "$UNFORMATTED" ]; then
        print_success "go fmt: código está formatado corretamente"
    else
        print_warning "go fmt formatou os seguintes arquivos:"
        echo "$UNFORMATTED"
    fi
    
    # Verificar se golint está instalado
    if command -v golint &> /dev/null; then
        echo ""
        echo "🔍 Executando golint..."
        LINT_OUTPUT=$(golint ./...)
        if [ -z "$LINT_OUTPUT" ]; then
            print_success "golint: nenhum problema encontrado"
        else
            print_warning "golint encontrou problemas:"
            echo "$LINT_OUTPUT"
        fi
    else
        print_warning "golint não está instalado - pulando análise de lint"
        echo "   Instale com: go install golang.org/x/lint/golint@latest"
    fi
}

# Testar compilação para diferentes plataformas
test_cross_compilation() {
    print_header "TESTE DE COMPILAÇÃO CRUZADA"
    
    echo "🏗️  Testando compilação para diferentes plataformas..."
    
    PLATFORMS=(
        "linux/amd64"
        "linux/arm64"
        "darwin/amd64"
        "darwin/arm64"
        "windows/amd64"
        "windows/arm64"
    )
    
    for platform in "${PLATFORMS[@]}"; do
        IFS='/' read -r GOOS GOARCH <<< "$platform"
        echo "   Compilando para $GOOS/$GOARCH..."
        
        if GOOS=$GOOS GOARCH=$GOARCH go build -o go24k_test_${GOOS}_${GOARCH} .; then
            print_success "✓ $GOOS/$GOARCH"
            rm -f go24k_test_${GOOS}_${GOARCH}*  # Limpar binário
        else
            print_error "✗ Falha na compilação para $GOOS/$GOARCH"
            return 1
        fi
    done
}

# Verificar dependências
check_dependencies() {
    print_header "VERIFICAÇÃO DE DEPENDÊNCIAS"
    
    echo "📦 Verificando módulo Go..."
    if go mod verify; then
        print_success "Módulo Go verificado com sucesso"
    else
        print_error "Problemas com módulo Go!"
        return 1
    fi
    
    echo ""
    echo "📦 Verificando dependências não utilizadas..."
    go mod tidy
    
    if git diff --quiet go.mod go.sum 2>/dev/null; then
        print_success "Dependências estão limpas"
    else
        print_warning "go mod tidy fez alterações - verifique go.mod e go.sum"
    fi
}

# Função principal
main() {
    echo "🚀 Iniciando suite completa de testes para Go24K"
    echo "================================================"
    
    # Limpar antes de começar
    cleanup
    
    # Verificações iniciais
    check_go
    check_dependencies
    
    # Executar análise estática primeiro
    run_static_analysis
    
    # Executar testes
    run_unit_tests
    run_main_tests
    run_integration_tests
    
    # Benchmarks (opcional)
    run_benchmarks
    
    # Teste de compilação cruzada
    test_cross_compilation
    
    print_header "RESUMO FINAL"
    print_success "Todos os testes foram executados!"
    print_success "O projeto Go24K está pronto para uso!"
    
    # Mostrar informações finais
    if [ -f coverage.out ]; then
        COVERAGE=$(go tool cover -func=coverage.out | tail -1 | awk '{print $3}')
        print_success "Cobertura de código: $COVERAGE"
    fi
    
    echo ""
    echo "📁 Arquivos gerados:"
    [ -f coverage.out ] && echo "   • coverage.out - Dados de cobertura"
    [ -f coverage.html ] && echo "   • coverage.html - Relatório HTML de cobertura"
    
    # Limpar arquivos temporários no final
    cleanup
    
    echo ""
    print_success "Suite de testes concluída com sucesso! 🎉"
}

# Verificar argumentos da linha de comando
case "${1:-}" in
    "unit")
        check_go && run_unit_tests
        ;;
    "integration")
        check_go && run_integration_tests
        ;;
    "bench")
        check_go && run_benchmarks
        ;;
    "lint")
        check_go && run_static_analysis
        ;;
    "build")
        check_go && test_cross_compilation
        ;;
    "clean")
        cleanup
        ;;
    "help"|"-h"|"--help")
        echo "Uso: $0 [COMANDO]"
        echo ""
        echo "Comandos disponíveis:"
        echo "  (nenhum)     Executa todos os testes"
        echo "  unit         Apenas testes unitários"
        echo "  integration  Apenas testes de integração"
        echo "  bench        Apenas benchmarks"
        echo "  lint         Apenas análise estática"
        echo "  build        Apenas teste de compilação"
        echo "  clean        Limpar arquivos temporários"
        echo "  help         Mostrar esta ajuda"
        echo ""
        echo "Exemplos:"
        echo "  $0           # Executar todos os testes"
        echo "  $0 unit      # Apenas testes unitários"
        echo "  $0 clean     # Limpar arquivos temporários"
        ;;
    "")
        main
        ;;
    *)
        print_error "Comando desconhecido: $1"
        echo "Use '$0 help' para ver comandos disponíveis"
        exit 1
        ;;
esac