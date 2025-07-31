# Build helpers

all:
	go build -o mcpGW github.com/thirdmartini/mcpgw/cmd/mcpgw
	go build -o reminders github.com/thirdmartini/mcpgw/example/mcpservers/reminders


run: all
	./mcpGW --config ./config.json