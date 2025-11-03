# Go24K

Utilitário pessoal para criar vídeos 4K a partir de imagens JPEG com efeitos Ken Burns e transições suaves.

## O que faz

- Converte imagens JPEG para 4K UHD (3840x2160)
- Cria vídeos com efeito Ken Burns (zoom/pan suave)
- Adiciona música de fundo automática (se houver MP3)
- Transições crossfade entre imagens
- Usa timestamps EXIF para ordenação cronológica

## Aceleração de Hardware

Detecta automaticamente e usa a melhor opção disponível:
- **NVIDIA NVENC**: 5-10x mais rápido (GPUs GeForce GTX 10+ / RTX)
- **Apple VideoToolbox**: 3-8x mais rápido (Apple Silicon M1/M2/M3, macOS)
- **Windows Media Foundation**: 3-5x mais rápido (Snapdragon X, Intel, AMD no Windows)
- **Intel Quick Sync**: 2-4x mais rápido (processadores com gráficos integrados)
- **AMD AMF**: 2-4x mais rápido (GPUs/APUs AMD)
- **Linux VAAPI**: 2-4x mais rápido (Linux com drivers VAAPI)
- **CPU**: Fallback universal (funciona em qualquer sistema)

## Requisitos

- FFmpeg instalado no sistema
- Go 1.16+ (apenas para compilar)

## Instalação

1. **Instalar FFmpeg**: https://ffmpeg.org/download.html
2. **Compilar**:
   ```bash
   git clone https://github.com/aloula/go24k.git
   cd go24k
   go build -o go24k
   chmod +x go24k  # Linux/macOS apenas
   ```

## Como Usar

1. **Colocar imagens JPEG** no diretório atual
2. **Opcional**: Adicionar arquivo MP3 para música de fundo
3. **Executar**:
   ```bash
   ./go24k
   ```

### Opções

```bash
./go24k [OPÇÕES]
```

- `-d <segundos>` - Duração por imagem (padrão: 5)
- `-t <segundos>` - Duração da transição (padrão: 1)
- `-static` - Desabilita efeito Ken Burns
- `-convert-only` - Apenas converte imagens, sem gerar vídeo
- `--debug` - Mostra informações de hardware

**Exemplos:**
```bash
# Padrão (5s por imagem, 1s transição)
./go24k

# Rápido (2s por imagem, sem Ken Burns)
./go24k -d 2 -static

# Apenas converter imagens
./go24k -convert-only
```

## Compilar para Outras Plataformas

```bash
# Atual plataforma
go build -o go24k

# Específica
GOOS=linux GOARCH=amd64 go build -o go24k-linux
GOOS=darwin GOARCH=arm64 go build -o go24k-macos
GOOS=windows GOARCH=amd64 go build -o go24k.exe

# Todas (usando build.sh)
./build.sh
```

## Saída

Gera:
- **`converted/`** - Imagens processadas em 4K
- **`video.mp4`** - Vídeo final 4K UHD (H.264, 30fps)

## Problemas Comuns

- **FFmpeg não encontrado**: Instalar e verificar se está no PATH
- **Sem imagens**: Colocar arquivos .jpg no diretório atual
- **Permissão negada**: `chmod +x go24k` (Linux/macOS)
- **Sem aceleração**: Drivers atualizados e FFmpeg com suporte a hardware

## Desenvolvimento

### Executar Testes

O projeto inclui uma suite completa de testes:

```bash
# Todos os testes
./test.sh

# Apenas testes unitários
./test.sh unit

# Apenas testes de integração
./test.sh integration

# Benchmarks de performance
./test.sh bench

# Análise estática (vet, fmt, lint)
./test.sh lint

# Teste de compilação cruzada
./test.sh build

# Limpar arquivos temporários
./test.sh clean
```

### Cobertura de Testes

- **Testes Unitários**: Funções de conversão de imagens e geração de vídeo
- **Testes de Integração**: Workflow completo com FFmpeg
- **Benchmarks**: Performance de conversão de imagens
- **Análise Estática**: Qualidade e formatação do código

O relatório de cobertura é gerado automaticamente em `coverage.html`.

### Comandos Make Disponíveis

```bash
# Desenvolvimento rápido
make dev          # fmt + lint + test-unit
make check        # Verificação pré-commit

# Compilação
make build        # Compilar para plataforma atual
make build-all    # Compilar para todas as plataformas

# Testes específicos
make test-unit         # Apenas testes unitários
make test-integration  # Apenas testes de integração
make bench             # Benchmarks de performance
make coverage          # Relatório HTML de cobertura

# Utilitários
make clean        # Limpar arquivos gerados
make install      # Instalar no sistema
make help         # Ver todos os comandos
```

### Estrutura dos Testes

```
├── utils/
│   ├── convertImages_test.go    # Testes de conversão de imagens
│   └── generateVideo_test.go    # Testes de geração de vídeo
├── main_test.go                 # Testes da função main
├── integration_test.go          # Testes de integração
├── test.sh                     # Script de execução de testes
├── Makefile                    # Comandos de desenvolvimento
└── .github/workflows/ci.yml    # CI/CD automático
```

## Licença

MIT License