# mcpGW
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

2. Optional whisper.cpp support for speech to prompt
- Install Whisper.cpp from https://github.com/ggml-org/whisper.cpp
- Run whisper server https://github.com/ggml-org/whisper.cpp/tree/master/examples/server
- Install a copy of ffmpeg into /tools

3. Optional melotts
- https://github.com/myshell-ai/MeloTTS


4. Create a server config.. see config.sample.json 

```config.sample.json
{
  "UI": {
    "Listen": ":8080",
    "TLS": true,
    "Root": "example/ui"
  },
  "_comment": "You don't need SpeechToText or TextToSpeech, current code only suports whisper-server and melotts",
  "SpeechToText": {
    "Provider": "whisper",
    "Host":  "http://10.0.0.10:8802",
    "Model": ""
  },
  "TextToSpeech": {
    "Provider": "melotts",
    "Host":  "http://10.0.0.10:7777",
    "Model": ""
  },
  "Inference": {
       "Provider": "ollama",
       "Host":  "http://localhost:11434",
       "Token": "",
       "Model": "mistral-small3.1",
       "SystemPrompt": ""
   },
   "Servers": {
     "mcpServers": {
       "reminders": {
         "command": "./reminders",
         "args": [
         ]
       }
     }
   }
}
```


5. Build and run
```bash
make all
./mcpgateway --config ./config.sample.json
```

## Writing your own application

mcpGW provides a set of apis (see server/server.go) and will run any application pointed to by the config UI section:

We included a basic sample UI:
```aiignore
  "UI": {
    "Listen": ":8080",
    "TLS": true,
    "Root": "example/ui"
  },
```


## Writing your own mcp server 

mcpGW will run any mcpserver, but we also included a simple one in ```examples/mcpservers/reminders```

You can then add the server to the mcpServers section of the config

```
 "Servers": {
     "mcpServers": {
       "reminders": {
         "command": "./reminders",
         "args": [
         ]
       }
     }
   }
```

