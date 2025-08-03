package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// MCPRequest represents an MCP request for testing
type MCPRequest struct {
	Method string                 `json:"method"`
	Params map[string]interface{} `json:"params"`
	ID     string                 `json:"id"`
}

// MCPResponse represents an MCP response
type MCPResponse struct {
	Result interface{} `json:"result,omitempty"`
	Error  *MCPError   `json:"error,omitempty"`
	ID     string      `json:"id"`
}

// MCPError represents an MCP error
type MCPError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func main() {
	fmt.Println("🚀 Testing Contact Management MCP Server")
	fmt.Println("========================================")

	// Test cases
	testCases := []struct {
		name    string
		request MCPRequest
	}{
		{
			name: "Initialize Server",
			request: MCPRequest{
				Method: "initialize",
				Params: map[string]interface{}{
					"protocolVersion": "2024-11-05",
				},
				ID: "init-1",
			},
		},
		{
			name: "List Available Tools",
			request: MCPRequest{
				Method: "tools/list",
				Params: map[string]interface{}{},
				ID:     "tools-1",
			},
		},
		{
			name: "Create Test Contact",
			request: MCPRequest{
				Method: "tools/call",
				Params: map[string]interface{}{
					"name": "create_contact",
					"arguments": map[string]interface{}{
						"name":     "Test User",
						"email":    "test@example.com",
						"phone":    "+1-555-TEST",
						"company":  "Test Company",
						"position": "Test Manager",
						"notes":    "Created via MCP test",
					},
				},
				ID: "create-1",
			},
		},
		{
			name: "Search Contacts",
			request: MCPRequest{
				Method: "tools/call",
				Params: map[string]interface{}{
					"name": "search_contacts",
					"arguments": map[string]interface{}{
						"query": "test",
						"limit": 5,
					},
				},
				ID: "search-1",
			},
		},
		{
			name: "Get Analytics",
			request: MCPRequest{
				Method: "tools/call",
				Params: map[string]interface{}{
					"name": "get_analytics",
					"arguments": map[string]interface{}{
						"granularity": "daily",
					},
				},
				ID: "analytics-1",
			},
		},
	}

	// Check if MCP server binary exists or can be built
	fmt.Println("📦 Building MCP server...")
	buildCmd := exec.Command("go", "build", "-o", "mcp-server", "./cmd/mcp-server")
	if err := buildCmd.Run(); err != nil {
		fmt.Printf("❌ Failed to build MCP server: %v\n", err)
		fmt.Println("💡 Make sure you're running this from the contact-service root directory")
		return
	}
	fmt.Println("✅ MCP server built successfully")

	// Start MCP server process
	fmt.Println("🔄 Starting MCP server...")
	cmd := exec.Command("./mcp-server")
	
	stdin, err := cmd.StdinPipe()
	if err != nil {
		fmt.Printf("❌ Failed to create stdin pipe: %v\n", err)
		return
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Printf("❌ Failed to create stdout pipe: %v\n", err)
		return
	}

	if err := cmd.Start(); err != nil {
		fmt.Printf("❌ Failed to start MCP server: %v\n", err)
		return
	}

	// Give the server time to start
	time.Sleep(2 * time.Second)
	fmt.Println("✅ MCP server started")

	// Create a scanner to read responses
	scanner := bufio.NewScanner(stdout)
	
	// Run test cases
	for i, testCase := range testCases {
		fmt.Printf("\n🧪 Test %d: %s\n", i+1, testCase.name)
		
		// Send request
		requestJSON, _ := json.Marshal(testCase.request)
		fmt.Printf("📤 Request: %s\n", string(requestJSON))
		
		_, err := stdin.Write(append(requestJSON, '\n'))
		if err != nil {
			fmt.Printf("❌ Failed to send request: %v\n", err)
			continue
		}

		// Read response with timeout
		responseChan := make(chan string, 1)
		go func() {
			if scanner.Scan() {
				responseChan <- scanner.Text()
			}
		}()

		select {
		case response := <-responseChan:
			fmt.Printf("📥 Response: %s\n", response)
			
			// Parse and validate response
			var mcpResponse MCPResponse
			if err := json.Unmarshal([]byte(response), &mcpResponse); err != nil {
				fmt.Printf("⚠️  Failed to parse response JSON: %v\n", err)
			} else {
				if mcpResponse.Error != nil {
					fmt.Printf("❌ Error: Code %d - %s\n", mcpResponse.Error.Code, mcpResponse.Error.Message)
				} else {
					fmt.Println("✅ Success")
				}
			}
		case <-time.After(10 * time.Second):
			fmt.Println("⏰ Response timeout")
		}
	}

	// Clean up
	fmt.Println("\n🧹 Cleaning up...")
	stdin.Close()
	cmd.Process.Kill()
	cmd.Wait()

	// Remove test binary
	os.Remove("mcp-server")

	fmt.Println("\n🎉 MCP Server test completed!")
	fmt.Println("\n💡 Tips for real usage:")
	fmt.Println("   • Ensure database is running and configured")
	fmt.Println("   • Add proper error handling for production")
	fmt.Println("   • Configure MCP client (like Claude Desktop) to use this server")
	fmt.Println("   • See docs/MCP_INTEGRATION.md for detailed setup instructions")
}

// Helper function to format JSON output
func formatJSON(data interface{}) string {
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Sprintf("Error formatting JSON: %v", err)
	}
	return string(jsonBytes)
}

// Interactive test mode
func runInteractiveTest() {
	fmt.Println("\n🔧 Interactive MCP Test Mode")
	fmt.Println("============================")
	fmt.Println("Enter MCP requests (JSON format), or 'quit' to exit:")

	reader := bufio.NewReader(os.Stdin)
	
	for {
		fmt.Print("MCP> ")
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Error reading input: %v\n", err)
			continue
		}

		input = strings.TrimSpace(input)
		if input == "quit" || input == "exit" {
			break
		}

		// Try to parse as JSON
		var request MCPRequest
		if err := json.Unmarshal([]byte(input), &request); err != nil {
			fmt.Printf("Invalid JSON: %v\n", err)
			fmt.Println("Example: {\"method\":\"tools/list\",\"params\":{},\"id\":\"1\"}")
			continue
		}

		fmt.Printf("Parsed request: %+v\n", request)
	}

	fmt.Println("Goodbye! 👋")
}