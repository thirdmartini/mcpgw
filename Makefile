all:
	go build github.com/thirdmartini/mcpgw/cmd/mcpgw
	go build github.com/thirdmartini/mcpgw/example/mcpservers/reminders

# run-basic runs the barebones gw using the example mcphost.example.json config that enables the example reminders mcp tool server
run-basic: all
	OLLAMA_HOST="http://localhost:11434" ./mcpgw  --config ./mcphost.example.json -mollama:mistral-small3.1 --server-root=example/ui
.PHONY: run

# Enables advanced features like whisper voice to text, and melo text to speech
run: all
	OLLAMA_HOST="http://localhost:11434" go run github.com/thirdmartini/mcpgw/cmd/mcpgw  --config ./mcphost.json -mollama:mistral-small3.1 --server-root=example/ui --whisper=http://10.0.0.208:8802 --melo=http://10.0.0.208:7777
.PHONY: run
