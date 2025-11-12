# Go24K

Utilitário pessoal para criar vídeos 4K a partir de imagens JPEG com efeitos Ken Burns e transições suaves.

## O que faz

- Converte imagens JPEG para 4K UHD (3840x2160) com upscaling inteligente
- Cria vídeos com efeito Ken Burns (zoom/pan suave) com 9 variações
- Adiciona música de fundo automática (se houver MP3)
- Transições crossfade entre imagens com duração configurável
- Usa timestamps EXIF para ordenação cronológica automática
- **NOVO**: Exibe detalhes técnicos do vídeo gerado (tamanho, bitrate, framerate)

## Aceleração de Hardware 🚀

Detecta automaticamente e usa a melhor opção disponível com **detecção real de funcionalidade** (não apenas presença):

### 🏆 **Hardware Encoders (Ordem de Prioridade)**
- **NVIDIA NVENC**: 5-10x mais rápido (GPUs GeForce GTX 10+ / RTX)
  - Bitrate: ~15 Mbps para 4K, qualidade excepcional
  - **NOVO**: Evita falsos positivos em WSL
- **Apple VideoToolbox**: 3-8x mais rápido (Apple Silicon M1/M2/M3/M4, macOS)
  - Otimizado para processadores Apple com Neural Engine
- **Windows Media Foundation**: 3-5x mais rápido (Windows 10/11)
  - **OTIMIZADO**: Snapdragon X Plus agora atinge 20+ Mbps (antes <10 Mbps)
  - Excelente para processadores Intel/AMD/ARM no Windows
- **Intel Quick Sync (QSV)**: 2-4x mais rápido (iGPU Intel 7ª geração+)
  - Disponível em processadores Intel Core com gráficos integrados
- **AMD AMF**: 2-4x mais rápido (GPUs/APUs AMD Radeon)
  - Suporte para placas discretas e APUs AMD
- **Linux VAAPI**: 2-4x mais rápido (Linux com drivers VAAPI)
  - Funciona com Intel iGPU e algumas GPUs AMD no Linux

### 💻 **Software Fallback**
- **CPU libx264**: Fallback universal, funciona em qualquer sistema
  - CRF 21 para alta qualidade com compressão eficiente

## 📷 Legenda EXIF Automática

### 🆕 **Nova Funcionalidade: Overlay de Informações da Câmera**

O Go24K agora pode extrair automaticamente informações técnicas das fotos e exibi-las como legenda no rodapé direito do vídeo.

#### **Informações Exibidas:**
- **Câmera**: Fabricante e modelo (ex: "Canon EOS R5")
- **Lente**: Modelo da lente (ex: "RF 24-70mm F2.8 L IS USM")
- **Configurações técnicas**:
  - **Distância focal**: ex: "50mm"
  - **Abertura**: ex: "f/2.8"
  - **Velocidade do obturador**: ex: "1/125s"
  - **ISO**: ex: "ISO 400"

#### **Como Usar:**
```bash
# Habilitar legenda EXIF (desabilitada por padrão)
./go24k -exif-overlay

# Combinar com outras opções
./go24k -exif-overlay -d 8 -t 2
```

#### **Exemplo de Legenda:**
```
Canon EOS R5
RF 24-70mm F2.8 L IS USM
50mm • f/2.8 • 1/125s • ISO 400
```

#### **Notas Técnicas:**
- ✅ **Dados extraídos dos arquivos originais**: As informações vêm dos arquivos JPEG originais antes da conversão
- ✅ **Fallback inteligente**: Se alguns dados EXIF não estiverem disponíveis, exibe apenas os disponíveis
- ✅ **Posicionamento otimizado**: Rodapé direito com fundo semi-transparente para legibilidade
- ✅ **Sem impacto na performance**: Extração rápida durante o processamento

## Requisitos

- **FFmpeg** 4.0+ instalado no sistema (com `ffprobe`)
- **Go 1.25+** (apenas para compilar do código-fonte)
- **Drivers atualizados** para melhor aceleração de hardware

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

#### Parâmetros Principais
- `-d <segundos>` - Duração por imagem (padrão: 5)
- `-t <segundos>` - Duração da transição crossfade (padrão: 1)
- `-static` - Desabilita efeito Ken Burns (imagens estáticas)
- `-convert-only` - Apenas converte imagens, sem gerar vídeo

#### Utilitários
- `--debug` - **NOVO**: Mostra detecção completa de hardware e configurações FFmpeg
- `--exif-overlay` - **NOVO**: Adiciona legenda com informações da câmera no rodapé direito
- `--help` - Exibe ajuda com todas as opções

**Exemplos:**
```bash
# Padrão (5s por imagem, 1s transição, Ken Burns ativo)
./go24k

# Rápido (2s por imagem, sem Ken Burns)  
./go24k -d 2 -static

# Longo com transições suaves (8s por imagem, 2s transição)
./go24k -d 8 -t 2

# Apenas converter imagens para 4K
./go24k -convert-only

# Com legenda de informações da câmera
./go24k -exif-overlay

# Verificar hardware disponível  
./go24k --debug
```

## Compilação

### Build Atual
```bash
# Plataforma atual
go build -o go24k
```

### Cross-Platform Build
```bash
# Builds específicos
GOOS=linux GOARCH=amd64 go build -o go24k-linux
GOOS=darwin GOARCH=arm64 go build -o go24k-macos  
GOOS=windows GOARCH=amd64 go build -o go24k.exe

# Todas as plataformas automaticamente
./build.sh
```

### 📦 **Builds Automáticos Disponíveis**
O projeto gera automaticamente executáveis para:
- **Linux**: AMD64 + ARM64
- **macOS**: Intel + Apple Silicon  
- **Windows**: AMD64 + ARM64

Arquivos gerados em `builds/`:
```
builds/
├── linux/
│   ├── amd64/go24k     # Intel/AMD Linux
│   └── arm64/go24k     # ARM Linux (Raspberry Pi, etc.)
├── macos/
│   ├── intel/go24k     # Intel Mac
│   └── arm/go24k       # Apple Silicon (M1/M2/M3/M4)
└── windows/
    ├── amd64/go24k.exe # Intel/AMD Windows
    └── arm64/go24k.exe # ARM Windows (Snapdragon)
```

## Saída

### Arquivos Gerados
- **`converted/`** - Imagens processadas em 4K UHD com upscaling inteligente
- **`video.mp4`** - Vídeo final 4K UHD (H.264, 30fps, alta qualidade)

### 📊 **NOVO: Informações Técnicas Detalhadas** 

Ao final da geração, o Go24K exibe automaticamente os detalhes técnicos do vídeo:

```
📹 Video Details:
File Size: 45.7 MB
Duration: 32.5 seconds  
Video Bitrate: 18.4 Mbps
Audio Bitrate: 128 kbps
Framerate: 30 fps
Resolution: 4K UHD (3840x2160)
Total time: 8.3 sec.
```

#### O que significam esses números:
- **File Size**: Tamanho total do arquivo de vídeo
- **Video Bitrate**: Taxa de bits real do vídeo (importante para qualidade)
  - NVENC: ~15-18 Mbps
  - Snapdragon X Plus: ~20+ Mbps (otimizado)
  - CPU: ~12-15 Mbps
- **Audio Bitrate**: Taxa do áudio (128-320 kbps) ou "No audio"
- **Duration**: Tempo exato calculado do vídeo final
- **Total time**: Tempo de processamento (conversão + geração)

## Problemas Comuns

### 🔧 **Instalação e Execução**
- **FFmpeg não encontrado**: Instalar FFmpeg e verificar se está no PATH
  ```bash
  # Verificar instalação
  ffmpeg -version
  ffprobe -version
  ```
- **Sem imagens**: Colocar arquivos `.jpg` no diretório atual (mínimo 2 imagens)
- **Permissão negada**: `chmod +x go24k` (Linux/macOS)
- **"No such file or directory"**: Verificar se o executável foi compilado corretamente

### ⚡ **Aceleração de Hardware**
- **Sem aceleração detectada**: 
  - Atualizar drivers de vídeo
  - Verificar se FFmpeg foi compilado com suporte aos codecs de hardware
  - Usar `./go24k --debug` para diagnosticar
- **WSL detectando NVENC incorretamente**: 
  - ✅ **CORRIGIDO**: Agora usa detecção real de funcionalidade
- **Bitrate baixo no Snapdragon**: 
  - ✅ **CORRIGIDO**: Otimizado para 20+ Mbps

### 📊 **Qualidade de Vídeo**
- **Vídeo com qualidade baixa**: Verificar se aceleração de hardware está funcionando
- **Arquivo muito grande**: Usar `-static` para desabilitar Ken Burns
- **Sem áudio**: Verificar se há arquivo MP3 no diretório

## Desenvolvimento

### Qualidade de Código 🏆

O projeto mantém **altos padrões de qualidade** com:
- ✅ **golangci-lint** moderno com **revive** + **stylecheck**
- ✅ **42.1% cobertura de testes** com casos edge-cases
- ✅ **Complexidade baixa** (todas funções < 25 ciclos)
- ✅ **0 problemas de linting** detectados
- ✅ **Refatoração modular** para melhor manutenibilidade

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

# Análise estática moderna (vet, fmt, golangci-lint)
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

## 🆕 Novidades Recentes

### v2.1.0 - Novembro 2025

#### 🎯 **Melhorias de Qualidade**
- ✅ **Resolução de problemas do linter**: 0 issues detectados
- ✅ **Refatoração da função `getVideoDetails()`**: Complexidade reduzida de 44 → <25
- ✅ **Constante `resolution4K`**: Substituição de 20+ strings hardcoded
- ✅ **golangci-lint moderno**: Upgrade para revive + stylecheck (substituto do golint deprecado)

#### 🚀 **Otimizações de Performance**  
- ✅ **Snapdragon X Plus**: Bitrate otimizado para 20+ Mbps (antes <10 Mbps)
- ✅ **Detecção de hardware melhorada**: Evita falsos positivos em WSL
- ✅ **Configurações agressivas**: Media Foundation com 25M max bitrate

#### 📊 **Funcionalidades Novas**
- ✅ **Legenda EXIF automática**: Overlay no rodapé direito com informações da câmera
  - Modelo da câmera e fabricante
  - Modelo da lente (se disponível)
  - Distância focal, abertura, velocidade do obturador, ISO
- ✅ **Detalhes técnicos do vídeo**: Exibição automática de tamanho, bitrate, framerate
- ✅ **Interface melhorada**: Spinner reposicionado, formato de progresso mais limpo
- ✅ **Ambiente de desenvolvimento**: VS Code configurado para testes de integração

#### 🧪 **Testes e Validação**
- ✅ **Suite de testes atualizada**: 27 testes unitários + 12 integração
- ✅ **42.1% cobertura**: Incluindo todas as novas funcionalidades
- ✅ **Go 1.25.4**: Upgrade da versão com compatibilidade total

## 📈 Performance e Benchmarks

### Tempos Típicos (5 imagens, 5s cada, Ken Burns ativo)

| Hardware | Resolução | Tempo Total | Speedup | Bitrate |
|----------|-----------|-------------|---------|---------|
| **NVIDIA RTX 4090** | 4K | ~8-12s | 10x | ~18 Mbps |
| **Apple M3 Pro** | 4K | ~15-20s | 6x | ~16 Mbps |
| **Snapdragon X Plus** | 4K | ~20-25s | 4x | ~22 Mbps |
| **Intel i7 + QSV** | 4K | ~25-35s | 3x | ~15 Mbps |
| **AMD Ryzen + CPU** | 4K | ~60-90s | 1x | ~14 Mbps |

### Otimizações Implementadas

- **NVENC**: CQ 21 para qualidade consistente
- **VideoToolbox**: Preset balanced para speed/quality  
- **Media Foundation**: 20M target + 25M max bitrate (Snapdragon)
- **QSV**: Preset medium com CQ 21
- **CPU**: Preset slow + CRF 21 para máxima qualidade

### Comandos Make Disponíveis

```bash
# Desenvolvimento rápido
make dev          # fmt + lint + test-unit
make check        # Verificação pré-commit

# Compilação
make build        # Compilar para plataforma atual
make build-all    # Compilar para todas as plataformas

# Testes específicos
make test-unit         # Apenas testes unitários (27 testes, 42.1% cobertura)
make test-integration  # Apenas testes de integração (12 testes)
make bench             # Benchmarks de performance
make coverage          # Relatório HTML de cobertura

# Análise de Código  
make lint         # golangci-lint básico
make lint-modern  # golangci-lint moderno (revive + stylecheck)

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