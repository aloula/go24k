# Go24K

Ferramenta em Go para montar vídeos a partir de fotos JPEG, com suporte a 4K ou Full HD, Ken Burns, transições, música e mistura opcional de vídeos na mesma timeline.

## Recursos

- Converte JPEG para um canvas padronizado em 4K ou Full HD.
- Gera vídeo com Ken Burns, crossfade e fade de entrada e saída.
- Pode incluir vídeos na mesma timeline sem distorcer o enquadramento.
- Usa EXIF e metadados para ordenar cronologicamente, com fallback por nome.
- Pode manter o áudio dos vídeos e misturá-lo com MP3 de fundo.
- Mostra detalhes técnicos do vídeo gerado ao final.
- Detecta automaticamente aceleração por hardware e cai para CPU quando necessário.

## Requisitos

- FFmpeg com ffprobe no PATH.
- Go 1.25+ apenas para compilar do código-fonte.

## Instalação

```bash
git clone https://github.com/aloula/go24k.git
cd go24k
go build -o go24k
chmod +x go24k
```

## Uso rápido

Coloque pelo menos duas imagens JPG no diretório atual. Se quiser, adicione também arquivos MP3 e vídeos suportados.

```bash
./go24k
```

Saída padrão:

- converted/: imagens convertidas
- video.mp4: vídeo final

## Flags principais

- -d <segundos>: duração por imagem. Padrão: 5.
- -t <segundos>: duração da transição. Padrão: 1.
- -static: desabilita Ken Burns.
- -fps <30|60>: força o framerate de saída.
- -fullhd: gera em 1920x1080 em vez de 3840x2160.
- -convert-only: apenas converte imagens.
- -fit-audio: ajusta as imagens ao tempo da música quando aplicável.
- -include-videos: inclui mp4, mov, mkv, avi, webm e m4v na timeline.
- -keep-video-audio: preserva áudio dos vídeos de entrada.
- -order-by-filename: ignora metadata e ordena por nome.
- -kenburns-mode <cinematic|dynamic>: escolhe o perfil do movimento.
- -exif-overlay: adiciona legenda com dados da câmera.
- -overlay-font-size <pixels>: tamanho da fonte do overlay. Padrão: 36.
- --debug: mostra detecção de hardware e parâmetros do FFmpeg.

## Exemplos

```bash
# Padrão
./go24k

# Imagens estáticas com transição maior
./go24k -static -d 6 -t 2

# Full HD a 60 fps
./go24k -fullhd -fps 60

# Misturar fotos e vídeos
./go24k -include-videos

# Misturar áudio dos vídeos com o MP3 de fundo
./go24k -include-videos -keep-video-audio

# Ordenar pelo nome do arquivo
./go24k -order-by-filename

# Overlay EXIF
./go24k -exif-overlay -overlay-font-size 48

# Ajustar ao tempo da música
./go24k -fit-audio

# Diagnóstico de hardware
./go24k --debug
```

## EXIF overlay

Quando -exif-overlay está ativo, o programa tenta exibir fabricante, câmera, lente, distância focal, abertura, obturador, ISO e data de cada foto. Se algum campo não existir, ele simplesmente omite o que faltar.

Exemplo:

```text
Canon - EOS R5 - 50mm - f/2.8 - 1/125s - ISO 400 - 15/08/2024
```

## Build e desenvolvimento

Compilação local:

```bash
go build -o go24k
```

Builds para múltiplas plataformas:

```bash
./build.sh
```

Testes e checks mais usados:

```bash
./test.sh
./test.sh unit
./test.sh integration
./test.sh lint

make build
make build-all
make test-unit
make test-integration
make lint
```

## Problemas comuns

- FFmpeg não encontrado: confirme ffmpeg -version e ffprobe -version.
- Sem imagens suficientes: são necessárias pelo menos duas imagens ou mídias na timeline.
- Sem aceleração por hardware: use --debug e verifique drivers e suporte do FFmpeg.
- Sem áudio no resultado: confirme a presença de MP3 ou use -keep-video-audio.

## Licença

MIT