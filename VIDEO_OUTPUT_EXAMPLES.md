# Exemplo de Saída - Detalhes do Vídeo

Exemplo da nova funcionalidade que mostra detalhes técnicos do vídeo gerado:

## Exemplo no Windows com Snapdragon X Plus

```
Hardware: Media Foundation detected - using Windows hardware acceleration
Target bitrate: 20 Mbps | Max: 25 Mbps | Buffer: 50M | Mode: u_vbr

Generating video with audio...: ✓

=== Video generated successfully! ===
File: video.mp4
Resolution: 3840x2160 (4K UHD)
Duration: 15 seconds (15.0s actual)
File Size: 45.2 MB
Video Bitrate: 22.3 Mbps
Audio Bitrate: 192 kbps
Framerate: 30 fps
Images: 5
Audio Source: background_music.mp3
```

## Exemplo no Linux com CPU

```
CPU: Using libx264 software encoding

Generating video (no audio)...: ✓

=== Video generated successfully! ===
File: video.mp4
Resolution: 3840x2160 (4K UHD)
Duration: 12 seconds (12.0s actual)
File Size: 38.7 MB
Video Bitrate: 25.8 Mbps
Audio Bitrate: No audio
Framerate: 30 fps
Images: 4
Audio Source: None (no MP3 file found)
```

## Exemplo com NVIDIA NVENC

```
Hardware: NVIDIA NVENC detected - using GPU acceleration

Generating video with audio...: ✓

=== Video generated successfully! ===
File: video.mp4
Resolution: 3840x2160 (4K UHD)
Duration: 20 seconds (20.0s actual)
File Size: 62.1 MB
Video Bitrate: 24.7 Mbps
Audio Bitrate: 192 kbps
Framerate: 30 fps
Images: 7
Audio Source: vacation_music.mp3
```

## Informações Úteis

- **Video Bitrate**: Mostra o bitrate real alcançado, útil para verificar se as otimizações do Snapdragon X Plus estão funcionando (deve ser >15 Mbps)
- **File Size**: Tamanho final do arquivo para planejamento de armazenamento
- **Duration**: Duração exata vs esperada para validar processamento
- **Audio Bitrate**: Confirma se o áudio foi processado corretamente
- **Images**: Número de imagens processadas no vídeo