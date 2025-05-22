# MCPgw
An MCP gateway/proxy to simplify apps that use tools, derived from https://github.com/mark3labs/mcphost

This is a work in progress and only tested with ollama


## Demo Setup 🔧

1. Ollama Setup:
- Install Ollama from https://ollama.ai
- Pull your desired model:
```bash
ollama pull mistral
```
- Ensure Ollama is running:
```bash
ollama serve
```

2. Optional Whisper.cpp support for speech to prompt
- Install Whisper.cpp from https://github.com/ggml-org/whisper.cpp
- Run whisper server https://github.com/ggml-org/whisper.cpp/tree/master/examples/server
- Install a copy of ffmpeg into /tools

3. Build and run the demo app
```bash
go build github.com/thirdmartini/mcpgw/cmd/mcpgw
./mcpgw  --config ./mcphost.json -mollama:mistral-small3.1 --server-root=example/ui --whisper=http://localhost:8802
```

server-root points to your application html and js.
--whisper optionally points to your whisper server 

## Writing your own application

TODO