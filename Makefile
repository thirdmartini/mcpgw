# Build helpers

all:
	go build -o mcpGW github.com/thirdmartini/mcpgw/cmd/mcpgw
	go build -o reminders github.com/thirdmartini/mcpgw/example/mcpservers/reminders
	go build -o websearch github.com/thirdmartini/mcpgw/example/mcpservers/websearch


run: all
	./mcpGW --config ./config.json
