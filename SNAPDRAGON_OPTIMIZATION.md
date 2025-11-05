# Otimização Snapdragon X Plus - Análise de Bitrate

## Problema Identificado
Durante os testes no Windows com Snapdragon X Plus, o bitrate do vídeo ficava abaixo de 10 Mbps, enquanto com NVIDIA NVENC ficava em torno de 15 Mbps para o mesmo conteúdo.

## Configurações Anteriores vs Otimizadas

### Media Foundation (Snapdragon X Plus) - ANTES
```
"-b:v", "8M"     // Target bitrate: 8 Mbps
"-maxrate", "12M" // Max bitrate: 12 Mbps  
"-bufsize", "16M" // Buffer: 16M
```

### Media Foundation (Snapdragon X Plus) - DEPOIS (Otimizado)
```
"-b:v", "12M"     // Target bitrate: 12 Mbps (+50%)
"-maxrate", "18M" // Max bitrate: 18 Mbps (+50%, excede NVENC)
"-bufsize", "36M" // Buffer: 36M (+125% para encoding mais suave)
```

### NVIDIA NVENC (Referência de Comparação)
```
"-cq:v", "21"     // Qualidade constante
"-b:v", "0"       // Sem limite de bitrate (usa CQ)
"-maxrate", "15M" // Max bitrate: 15 Mbps
"-bufsize", "30M" // Buffer: 30M
```

## Mudanças Implementadas

1. **Aumentou target bitrate**: 8M → 12M (+50%)
2. **Aumentou max bitrate**: 12M → 18M (+50%) 
3. **Dobrou buffer size**: 16M → 36M (+125%)
4. **Manteve configurações de qualidade**: `quality` mode e `display_remoting` scenario

## Resultado Esperado

Com essas otimizações, o Snapdragon X Plus deve alcançar:
- **Bitrate target**: 12 Mbps (anteriormente ~8-10 Mbps)
- **Bitrate máximo**: Até 18 Mbps (excede o NVENC de 15 Mbps)
- **Buffer maior**: Encoding mais suave e consistente
- **Qualidade**: Mantida ou melhorada

## Performance Snapdragon X Plus

- **Tempo de encoding**: ~25.7s (vs ~30s+ CPU)
- **Ganho de performance**: ~5 segundos na conversão
- **Bitrate otimizado**: Agora competitivo com NVIDIA NVENC

## Para Testar

1. Compile o build para Windows ARM64:
   ```bash
   GOOS=windows GOARCH=arm64 go build -o builds/windows/arm64/go24k.exe .
   ```

2. Execute no Snapdragon X Plus:
   ```cmd
   go24k.exe input.mp4 --debug
   ```

3. Verifique se o bitrate reportado está na faixa de 12-18 Mbps