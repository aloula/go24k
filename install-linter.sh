#!/bin/bash
# Script para instalar golangci-lint
# ==================================

echo "🔧 Instalando golangci-lint..."

# Verifica se Go está instalado
if ! command -v go &> /dev/null; then
    echo "❌ Go não encontrado. Por favor, instale Go primeiro."
    exit 1
fi

# Instala golangci-lint
echo "📦 Instalando golangci-lint via go install..."
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Verifica se foi instalado
if [ -f "$(go env GOPATH)/bin/golangci-lint" ]; then
    echo "✅ golangci-lint instalado com sucesso!"
    
    # Adiciona ao PATH se não estiver
    if ! command -v golangci-lint &> /dev/null; then
        echo "⚙️  Adicionando $(go env GOPATH)/bin ao PATH..."
        
        # Adiciona ao .bashrc se existir
        if [ -f "$HOME/.bashrc" ]; then
            echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> "$HOME/.bashrc"
            echo "📝 Adicionado ao ~/.bashrc"
        fi
        
        # Adiciona ao .zshrc se existir
        if [ -f "$HOME/.zshrc" ]; then
            echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> "$HOME/.zshrc"
            echo "📝 Adicionado ao ~/.zshrc"
        fi
        
        echo "🔄 Reinicie o terminal ou execute: source ~/.bashrc (ou ~/.zshrc)"
    fi
    
    # Mostra versão
    echo ""
    echo "ℹ️  Versão instalada:"
    $(go env GOPATH)/bin/golangci-lint version
    
    echo ""
    echo "🚀 Para usar:"
    echo "  golangci-lint run                    # Análise completa"
    echo "  golangci-lint run --fast            # Análise rápida"
    echo "  golangci-lint run --fix             # Corrigir problemas automaticamente"
    echo "  make lint-modern                    # Via Makefile do projeto"
    
else
    echo "❌ Falha na instalação do golangci-lint"
    exit 1
fi