#!/bin/bash

# Test script for the LLM Tools MCP Server

echo "Building MCP server..."
go build -o ../../bin/mcp-server main.go

if [ $? -ne 0 ]; then
    echo "Failed to build MCP server"
    exit 1
fi

echo "✅ MCP server built successfully"
echo ""

echo "Testing MCP server initialization..."
echo '{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {"protocolVersion": "2024-11-05", "capabilities": {}, "clientInfo": {"name": "test", "version": "1.0.0"}}}' | ../../bin/mcp-server > /tmp/init_response.json

if [ $? -eq 0 ]; then
    echo "✅ Server initialization successful"
    echo "Response: $(cat /tmp/init_response.json | jq .)"
else
    echo "❌ Server initialization failed"
    exit 1
fi

echo ""
echo "Testing tools list..."
echo '{"jsonrpc": "2.0", "id": 2, "method": "tools/list"}' | ../../bin/mcp-server > /tmp/tools_response.json

if [ $? -eq 0 ]; then
    echo "✅ Tools list successful"
    echo "Available tools: $(cat /tmp/tools_response.json | jq '.result.tools[].name')"
else
    echo "❌ Tools list failed"
    exit 1
fi

echo ""
echo "Testing batch_read_file tool with type-safe implementation..."
echo '{"jsonrpc": "2.0", "id": 3, "method": "tools/call", "params": {"name": "batch_read_file", "arguments": {"explanation": "Testing type-safe BatchReadFile function", "files": [{"target_file": "../../go.mod", "should_read_entire_file": true}], "continue_on_error": true, "include_outline": true}}}' | ../../bin/mcp-server > /tmp/tool_response.json

if [ $? -eq 0 ]; then
    echo "✅ Type-safe batch_read_file tool test successful"
    echo "Response: $(cat /tmp/tool_response.json | jq '.result.content[0].text' | jq .)"
else
    echo "❌ Type-safe batch_read_file tool test failed"
    exit 1
fi

echo ""
echo "🎉 All tests passed! MCP server with type-safe BatchReadFile is working correctly."

# Clean up temp files
rm -f /tmp/init_response.json /tmp/tools_response.json /tmp/tool_response.json