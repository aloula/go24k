# Exemplo de Saída - Detalhes do Vídeo 📊

Exemplos da funcionalidade que mostra **detalhes técnicos completos** do vídeo gerado.

> **NOVO**: Informações extraídas automaticamente via `ffprobe` para validação precisa da qualidade e configurações utilizadas.

## 🚀 Exemplo no Windows com Snapdragon X Plus (OTIMIZADO)

```
Converting 5 images to 4K UHD...
[1/5] | vacation_001.jpg...
[2/5] | vacation_002.jpg...
[3/5] | vacation_003.jpg...
[4/5] | vacation_004.jpg...
[5/5] | vacation_005.jpg...

Hardware: Media Foundation detected - using Windows hardware acceleration  
🎯 Generating video with audio... ✓

📹 Video Details:
File Size: 45.2 MB
Duration: 15.0 seconds
Video Bitrate: 22.3 Mbps  ← OTIMIZADO! (antes <10 Mbps)
Audio Bitrate: 192 kbps
Framerate: 30 fps  
Resolution: 4K UHD (3840x2160)
Total time: 12.4 sec.
```

**🎯 Melhoria**: Snapdragon X Plus agora atinge consistentemente **20+ Mbps** vs <10 Mbps anterior.

## 💻 Exemplo no Linux com CPU (Software)

```
Converting 4 images to 4K UHD...
[1/4] | photo_001.jpg...
[2/4] | photo_002.jpg...  
[3/4] | photo_003.jpg...
[4/4] | photo_004.jpg...

CPU: Using libx264 software encoding
🎯 Generating video (no audio)... ✓

📹 Video Details:
File Size: 38.7 MB
Duration: 12.0 seconds
Video Bitrate: 14.2 Mbps  ← Alta qualidade CPU (CRF 21)
Audio Bitrate: No audio
Framerate: 30 fps
Resolution: 4K UHD (3840x2160)
Total time: 68.3 sec.
```

**💡 Observação**: CPU ainda oferece excelente qualidade, apenas mais lento que hardware acceleration.

## 🎮 Exemplo com NVIDIA NVENC (GPU)

```
Converting 7 images to 4K UHD...
[1/7] | summer_001.jpg...
[2/7] | summer_002.jpg...
[3/7] | summer_003.jpg...
[4/7] | summer_004.jpg...
[5/7] | summer_005.jpg...
[6/7] | summer_006.jpg...
[7/7] | summer_007.jpg...

Hardware: NVIDIA NVENC detected - using GPU acceleration
🎯 Generating video with audio... ✓

📹 Video Details:
File Size: 62.1 MB
Duration: 20.0 seconds  
Video Bitrate: 18.4 Mbps  ← Excelente qualidade NVENC
Audio Bitrate: 192 kbps
Framerate: 30 fps
Resolution: 4K UHD (3840x2160)
Total time: 8.7 sec.
```

**⚡ Performance**: NVENC oferece o melhor balance speed/quality para 4K.

## 🍎 Exemplo com Apple VideoToolbox (macOS)

```
Converting 6 images to 4K UHD...
[1/6] | nature_001.jpg...
[2/6] | nature_002.jpg...
[3/6] | nature_003.jpg...
[4/6] | nature_004.jpg...
[5/6] | nature_005.jpg...
[6/6] | nature_006.jpg...

Hardware: Apple VideoToolbox detected - using macOS hardware acceleration
🎯 Generating video with audio... ✓

📹 Video Details:
File Size: 51.3 MB
Duration: 18.0 seconds
Video Bitrate: 16.8 Mbps  ← Otimizado para Apple Silicon
Audio Bitrate: 256 kbps
Framerate: 30 fps
Resolution: 4K UHD (3840x2160)
Total time: 11.2 sec.
```

## 📊 Interpretação dos Dados

### 🎯 **Video Bitrate** (Mais Importante)
- **20+ Mbps**: Snapdragon X Plus otimizado ✅
- **15-18 Mbps**: NVENC, VideoToolbox (qualidade excelente) ✅  
- **12-15 Mbps**: CPU, QSV, AMF (boa qualidade) ✅
- **<10 Mbps**: ⚠️ Possível problema (verificar `--debug`)

### 📁 **File Size**
- **~3-4 MB por segundo de vídeo** é normal para 4K
- Arquivo maior = mais qualidade/detalhes
- Ken Burns desabilitado (`-static`) = arquivo menor

### ⏱️ **Total Time** (Performance)
- **NVENC**: 8-15s para 5-7 imagens ⚡
- **VideoToolbox**: 10-20s para 5-7 imagens 🚀
- **Snapdragon**: 15-25s para 5-7 imagens ✨
- **CPU**: 60-90s para 5-7 imagens 🐌

### 🎵 **Audio Bitrate**
- **128-320 kbps**: Áudio de boa qualidade
- **"No audio"**: Nenhum MP3 encontrado no diretório

## 📷 **NOVO: Exemplo com Legenda EXIF** 

```
Converting 4 images to 4K UHD...
[1/4] | vacation_001.jpg...
[2/4] | vacation_002.jpg...
[3/4] | vacation_003.jpg...
[4/4] | vacation_004.jpg...

Hardware: NVIDIA NVENC detected - using GPU acceleration
🎯 Generating video with EXIF overlay... ✓

📹 Video Details:
File Size: 42.8 MB
Duration: 12.0 seconds
Video Bitrate: 17.2 Mbps
Audio Bitrate: 192 kbps
Framerate: 30 fps
Resolution: 4K UHD (3840x2160)
Total time: 9.1 sec.
```

### **🎬 Legenda Exibida no Vídeo:**

Rodapé centralizado com fundo semi-transparente:
```
Canon - EOS R5 - 85mm - f/1.8 - ISO 800 - 15.08.2024
```

**Formato atualizado:**
- Separadores com dashes (-) para compatibilidade Windows
- Data da foto extraída dos dados EXIF (DD.MM.YYYY)
- Layout compacto em linha única
- Posicionamento centralizado no rodapé

#### **Exemplos de Diferentes Câmeras:**

**📱 Smartphone:**
```
Apple - iPhone 15 Pro - 28mm - f/1.8 - ISO 125 - 12.08.2024
```

**🎥 Mirrorless:**
```
Sony - A7R V - 50mm - f/2.8 - ISO 400 - 15.07.2024
```

**📸 DSLR:**
```
Nikon - D850 - 85mm - f/2.0 - ISO 200 - 20.06.2024
```

**📷 Compacta:**
```
Fujifilm - X100VI - 23mm - f/2.0 - ISO 160 - 10.09.2024
```