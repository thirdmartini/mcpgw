package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	flagSearch := flag.String("search", "", "query to search for")
	flagNews := flag.String("news", "", "query to news for")
	flag.Parse()

	// BRAVE gets it's key from the environment (so mcphost has to set that in the env section
	//
	key := os.Getenv("BRAVE_API_KEY")
	if key == "" {
		os.Exit(-1)
	}

	// TODO: add support for different search providers using the SearchProvider interface
	// for now Brave search API is sufficient
	search := NewBraveSearch(key)

	if *flagSearch != "" {
		results, err := search.Search(*flagSearch, 1)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		fmt.Println(results)
		return
	}

	if *flagNews != "" {
		results, err := search.News(*flagNews, 1)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		fmt.Println(results)
		return
	}

	s := server.NewMCPServer(
		"Internet SearchProvider",
		"1.0.0",
		server.WithResourceCapabilities(true, true),
		server.WithLogging(),
	)
	// Add tool

	s.AddTool(mcp.NewTool("web_search",
		mcp.WithDescription("search the internet for a topic or query string"),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("The query string on which to search the web with"),
		),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return searchWeb(ctx, search, request)
	})

	s.AddTool(mcp.NewTool("news_search",
		mcp.WithDescription("get the latest news on a topic"),
		mcp.WithString("topic",
			mcp.Required(),
			mcp.Description("The topic to get latest news on"),
		),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return searchNews(ctx, search, request)
	})

	// Start the stdio server
	if err := server.ServeStdio(s); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}

}

func searchWeb(ctx context.Context, search SearchProvider, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query, ok := request.Params.Arguments["query"].(string)
	if !ok {
		return nil, errors.New("the title must be a string")
	}

	result, err := search.Search(query, 5)
	if err != nil {
		return nil, err
	}

	return mcp.NewToolResultText(result), nil
}

func searchNews(ctx context.Context, search SearchProvider, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query, ok := request.Params.Arguments["topic"].(string)
	if !ok {
		return nil, errors.New("topic must be a string")
	}

	result, err := search.News(query, 5)
	if err != nil {
		return nil, err
	}

	return mcp.NewToolResultText(result), nil
}
