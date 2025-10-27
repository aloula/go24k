# Exemplos de Uso do Go24K - GIFs Animados

## Como criar GIFs animados

### 🎯 **CONTROLE POR TEMPO TOTAL (RECOMENDADO)**

### 1. GIF rápido de 5 segundos
```bash
./go24k --gif-optimized --gif-total-time 5
```
- **NOVO**: Controla duração total do GIF
- 5 segundos totais, ~0.4s por imagem (12 imagens)
- Sistema recomenda FPS otimizado automaticamente

### 2. GIF ultra-rápido de 3 segundos
```bash
./go24k --gif-optimized --gif-total-time 3 --gif-fps 60
```
- 3 segundos totais, ~0.25s por imagem
- FPS alto para animação suave
- Ideal para demonstrações rápidas

### 3. GIF slideshow de 15 segundos
```bash
./go24k --gif-optimized --gif-total-time 15
```
- 15 segundos totais, ~1.25s por imagem
- Ritmo confortável para visualização
- Bom para apresentações

### 🔧 **CONTROLE POR IMAGEM (TRADICIONAL)**

### 4. GIF padrão otimizado
```bash
./go24k --gif-optimized
```
- Usa 1 segundo por imagem automaticamente
- 15 FPS (padrão otimizado)
- Para quando você quer controle granular

### 5. GIF micro para web
```bash
./go24k --gif-optimized --gif-total-time 4 --gif-scale 0.3
```
- 4 segundos totais
- Escala reduzida (30% do tamanho)
- Arquivo muito pequeno para web

## Dicas para otimização de tamanho

1. **Use --gif-optimized** sempre que possível
2. **Reduza o FPS**: 8-12 FPS geralmente é suficiente
3. **Diminua a escala**: 0.15-0.25 para web, 0.3-0.5 para qualidade
4. **Reduza a duração**: Menos tempo = arquivo menor
5. **Menos transição**: Transições mais curtas = arquivo menor

## Comparação de tamanhos e velocidades

### Velocidade vs Tamanho:
- **GIF padrão otimizado**: ~2.2MB, 12 segundos totais (1s/imagem)
- **GIF super rápido**: ~2.2MB, 6 segundos totais (0.5s/imagem)  
- **GIF slideshow**: ~2.2MB, 24 segundos totais (2s/imagem)
- **GIF micro web**: ~1MB, 12 segundos (escala reduzida)

### 🎯 Recomendações por uso:
- **Apresentações**: `--gif-optimized --gif-total-time 15` (ritmo confortável)
- **Redes sociais**: `--gif-optimized --gif-total-time 5` (dinâmico)
- **Web/mobile**: `--gif-optimized --gif-total-time 4 --gif-scale 0.3` (pequeno)
- **Demonstrações rápidas**: `--gif-optimized --gif-total-time 3 --gif-fps 60` (ultra-rápido)

### 💡 Como escolher o tempo total:
- **3-5 segundos**: Muito rápido, ideal para redes sociais
- **6-10 segundos**: Dinâmico mas confortável
- **10-15 segundos**: Ritmo de apresentação
- **15+ segundos**: Slideshow contemplativo

### ⚡ Sistema inteligente:
- **Tempo muito curto**: Sistema recomenda FPS otimizado automaticamente
- **Cálculo automático**: Divide o tempo total pelo número de imagens
- **Feedback visual**: Mostra duração calculada por imagem

## Como funciona a otimização

### Processo otimizado para GIFs:
1. **Conversão inteligente**: Imagens são convertidas para máximo 1080p (em vez de 4K UHD)
2. **Diretório separado**: Imagens otimizadas são salvas em `gif_converted/`
3. **Menor processamento**: Evita redimensionamento desnecessário durante geração do GIF
4. **Numeração ordenada**: Arquivos são numerados para manter ordem correta

### Comparação com o processo de vídeo:
- **Vídeo**: Converte para 4K UHD (3840x2160) → `uhd_converted/`
- **GIF**: Converte para máx 1080p → `gif_converted/`

## Resultados de exemplo:
- **Imagens originais**: ~3-8MB cada (variável)
- **Imagens convertidas para GIF**: ~200KB cada (1024x683)
- **GIF final otimizado**: ~2.3MB (12 imagens, 2s cada)

## Pré-requisitos

- Imagens JPEG no diretório atual
- FFmpeg instalado no sistema
- Go 1.16+ (se compilando do código fonte)