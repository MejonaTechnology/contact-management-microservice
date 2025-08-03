package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"contact-service/internal/models"
	"contact-service/internal/repository"
	"contact-service/internal/services"
	"contact-service/pkg/database"
	"contact-service/pkg/logger"

	"github.com/joho/godotenv"
)

// MCPServer represents the Model Context Protocol server
type MCPServer struct {
	contactRepo   repository.ContactRepository
	userRepo      repository.UserRepository
	contactService *services.ContactService
	bulkService   *services.BulkService
	analyticsService *services.AnalyticsService
}

// MCPRequest represents an incoming MCP request
type MCPRequest struct {
	Method string                 `json:"method"`
	Params map[string]interface{} `json:"params"`
	ID     string                 `json:"id"`
}

// MCPResponse represents an outgoing MCP response
type MCPResponse struct {
	Result interface{} `json:"result,omitempty"`
	Error  *MCPError   `json:"error,omitempty"`
	ID     string      `json:"id"`
}

// MCPError represents an MCP error
type MCPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Tool represents an available MCP tool
type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

// ToolResult represents the result of a tool execution
type ToolResult struct {
	Content []ContentBlock `json:"content"`
	IsError bool          `json:"isError,omitempty"`
}

// ContentBlock represents a content block in the response
type ContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Initialize logger
	logger.InitLogger()

	// Initialize database
	if err := database.InitDB(); err != nil {
		log.Fatal("Failed to initialize database:", err)
	}

	// Initialize repositories and services
	db := database.GetDB()
	contactRepo := repository.NewContactRepository(db)
	userRepo := repository.NewUserRepository(db)
	
	contactService := services.NewContactService(contactRepo, userRepo)
	bulkService := services.NewBulkService(contactRepo, userRepo)
	analyticsService := services.NewAnalyticsService(db)

	server := &MCPServer{
		contactRepo:      contactRepo,
		userRepo:         userRepo,
		contactService:   contactService,
		bulkService:      bulkService,
		analyticsService: analyticsService,
	}

	log.Println("MCP Server starting for Contact Management System")
	server.Start()
}

// Start starts the MCP server
func (s *MCPServer) Start() {
	// MCP servers typically communicate via stdio
	// For this implementation, we'll create a simple JSON-RPC interface
	
	for {
		var request MCPRequest
		decoder := json.NewDecoder(os.Stdin)
		
		if err := decoder.Decode(&request); err != nil {
			log.Printf("Error decoding request: %v", err)
			continue
		}

		response := s.handleRequest(request)
		
		encoder := json.NewEncoder(os.Stdout)
		if err := encoder.Encode(response); err != nil {
			log.Printf("Error encoding response: %v", err)
		}
	}
}

// handleRequest handles incoming MCP requests
func (s *MCPServer) handleRequest(request MCPRequest) MCPResponse {
	switch request.Method {
	case "initialize":
		return s.handleInitialize(request)
	case "tools/list":
		return s.handleToolsList(request)
	case "tools/call":
		return s.handleToolCall(request)
	default:
		return MCPResponse{
			Error: &MCPError{
				Code:    -32601,
				Message: "Method not found",
			},
			ID: request.ID,
		}
	}
}

// handleInitialize handles the initialize request
func (s *MCPServer) handleInitialize(request MCPRequest) MCPResponse {
	return MCPResponse{
		Result: map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]interface{}{
				"tools": map[string]interface{}{},
			},
			"serverInfo": map[string]interface{}{
				"name":    "contact-management-server",
				"version": "1.0.0",
			},
		},
		ID: request.ID,
	}
}

// handleToolsList returns the list of available tools
func (s *MCPServer) handleToolsList(request MCPRequest) MCPResponse {
	tools := []Tool{
		{
			Name:        "create_contact",
			Description: "Create a new contact in the system",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Full name of the contact",
					},
					"email": map[string]interface{}{
						"type":        "string",
						"description": "Email address of the contact",
					},
					"phone": map[string]interface{}{
						"type":        "string",
						"description": "Phone number of the contact",
					},
					"company": map[string]interface{}{
						"type":        "string",
						"description": "Company name",
					},
					"position": map[string]interface{}{
						"type":        "string",
						"description": "Job position/title",
					},
					"notes": map[string]interface{}{
						"type":        "string",
						"description": "Additional notes about the contact",
					},
				},
				"required": []string{"name", "email"},
			},
		},
		{
			Name:        "search_contacts",
			Description: "Search for contacts by various criteria",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query": map[string]interface{}{
						"type":        "string",
						"description": "Search query (name, email, company)",
					},
					"status": map[string]interface{}{
						"type":        "string",
						"description": "Filter by contact status",
					},
					"limit": map[string]interface{}{
						"type":        "integer",
						"description": "Maximum number of results (default: 10)",
					},
				},
			},
		},
		{
			Name:        "get_contact",
			Description: "Get detailed information about a specific contact",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"id": map[string]interface{}{
						"type":        "integer",
						"description": "Contact ID",
					},
					"email": map[string]interface{}{
						"type":        "string",
						"description": "Contact email address",
					},
				},
			},
		},
		{
			Name:        "update_contact",
			Description: "Update an existing contact's information",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"id": map[string]interface{}{
						"type":        "integer",
						"description": "Contact ID",
					},
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Updated name",
					},
					"email": map[string]interface{}{
						"type":        "string",
						"description": "Updated email",
					},
					"phone": map[string]interface{}{
						"type":        "string",
						"description": "Updated phone",
					},
					"company": map[string]interface{}{
						"type":        "string",
						"description": "Updated company",
					},
					"position": map[string]interface{}{
						"type":        "string",
						"description": "Updated position",
					},
					"status": map[string]interface{}{
						"type":        "string",
						"description": "Updated status",
					},
					"notes": map[string]interface{}{
						"type":        "string",
						"description": "Updated notes",
					},
				},
				"required": []string{"id"},
			},
		},
		{
			Name:        "delete_contact",
			Description: "Delete a contact from the system",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"id": map[string]interface{}{
						"type":        "integer",
						"description": "Contact ID to delete",
					},
				},
				"required": []string{"id"},
			},
		},
		{
			Name:        "get_analytics",
			Description: "Get contact analytics and metrics",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"start_date": map[string]interface{}{
						"type":        "string",
						"description": "Start date (YYYY-MM-DD)",
					},
					"end_date": map[string]interface{}{
						"type":        "string",
						"description": "End date (YYYY-MM-DD)",
					},
					"granularity": map[string]interface{}{
						"type":        "string",
						"description": "Data granularity (daily, weekly, monthly)",
					},
				},
			},
		},
		{
			Name:        "export_contacts",
			Description: "Export contacts to CSV or JSON format",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"format": map[string]interface{}{
						"type":        "string",
						"description": "Export format (csv or json)",
					},
					"status": map[string]interface{}{
						"type":        "string",
						"description": "Filter by status",
					},
					"limit": map[string]interface{}{
						"type":        "integer",
						"description": "Maximum records to export",
					},
				},
			},
		},
	}

	return MCPResponse{
		Result: map[string]interface{}{
			"tools": tools,
		},
		ID: request.ID,
	}
}

// handleToolCall handles tool execution requests
func (s *MCPServer) handleToolCall(request MCPRequest) MCPResponse {
	toolName, ok := request.Params["name"].(string)
	if !ok {
		return MCPResponse{
			Error: &MCPError{
				Code:    -32602,
				Message: "Invalid tool name",
			},
			ID: request.ID,
		}
	}

	arguments, ok := request.Params["arguments"].(map[string]interface{})
	if !ok {
		arguments = make(map[string]interface{})
	}

	var result ToolResult
	var err error

	switch toolName {
	case "create_contact":
		result, err = s.executeCreateContact(arguments)
	case "search_contacts":
		result, err = s.executeSearchContacts(arguments)
	case "get_contact":
		result, err = s.executeGetContact(arguments)
	case "update_contact":
		result, err = s.executeUpdateContact(arguments)
	case "delete_contact":
		result, err = s.executeDeleteContact(arguments)
	case "get_analytics":
		result, err = s.executeGetAnalytics(arguments)
	case "export_contacts":
		result, err = s.executeExportContacts(arguments)
	default:
		return MCPResponse{
			Error: &MCPError{
				Code:    -32601,
				Message: "Unknown tool: " + toolName,
			},
			ID: request.ID,
		}
	}

	if err != nil {
		return MCPResponse{
			Error: &MCPError{
				Code:    -32603,
				Message: err.Error(),
			},
			ID: request.ID,
		}
	}

	return MCPResponse{
		Result: result,
		ID:     request.ID,
	}
}

// executeCreateContact creates a new contact
func (s *MCPServer) executeCreateContact(args map[string]interface{}) (ToolResult, error) {
	name, _ := args["name"].(string)
	email, _ := args["email"].(string)
	phone, _ := args["phone"].(string)
	company, _ := args["company"].(string)
	position, _ := args["position"].(string)
	notes, _ := args["notes"].(string)

	if name == "" || email == "" {
		return ToolResult{IsError: true}, fmt.Errorf("name and email are required")
	}

	contact := &models.Contact{
		Name:     name,
		Email:    email,
		Phone:    phone,
		Company:  company,
		Position: position,
		Notes:    notes,
		Status:   "new",
	}

	if err := s.contactRepo.Create(contact); err != nil {
		return ToolResult{IsError: true}, fmt.Errorf("failed to create contact: %v", err)
	}

	result := fmt.Sprintf("Successfully created contact: %s (ID: %d)", contact.Name, contact.ID)
	return ToolResult{
		Content: []ContentBlock{{Type: "text", Text: result}},
	}, nil
}

// executeSearchContacts searches for contacts
func (s *MCPServer) executeSearchContacts(args map[string]interface{}) (ToolResult, error) {
	query, _ := args["query"].(string)
	status, _ := args["status"].(string)
	limitFloat, _ := args["limit"].(float64)
	limit := int(limitFloat)
	
	if limit == 0 {
		limit = 10
	}

	params := repository.ContactListParams{
		Page:   1,
		Limit:  limit,
		Search: query,
		Status: status,
		Sort:   "created_at",
		Order:  "desc",
	}

	contacts, total, err := s.contactRepo.List(params)
	if err != nil {
		return ToolResult{IsError: true}, fmt.Errorf("failed to search contacts: %v", err)
	}

	result := fmt.Sprintf("Found %d contacts (showing %d):\n\n", int(total), len(contacts))
	for _, contact := range contacts {
		result += fmt.Sprintf("• %s (%s) - %s - Status: %s\n", 
			contact.Name, contact.Email, contact.Company, contact.Status)
	}

	return ToolResult{
		Content: []ContentBlock{{Type: "text", Text: result}},
	}, nil
}

// executeGetContact gets a specific contact
func (s *MCPServer) executeGetContact(args map[string]interface{}) (ToolResult, error) {
	var contact *models.Contact
	var err error

	if idFloat, ok := args["id"].(float64); ok {
		id := uint(idFloat)
		contact, err = s.contactRepo.GetByID(id)
	} else if email, ok := args["email"].(string); ok && email != "" {
		contact, err = s.contactRepo.GetByEmail(email)
	} else {
		return ToolResult{IsError: true}, fmt.Errorf("either id or email is required")
	}

	if err != nil {
		return ToolResult{IsError: true}, fmt.Errorf("failed to get contact: %v", err)
	}

	result := fmt.Sprintf(`Contact Details:
• ID: %d
• Name: %s
• Email: %s
• Phone: %s
• Company: %s
• Position: %s
• Status: %s
• Notes: %s
• Created: %s
• Updated: %s`,
		contact.ID,
		contact.Name,
		contact.Email,
		contact.Phone,
		contact.Company,
		contact.Position,
		contact.Status,
		contact.Notes,
		contact.CreatedAt.Format("2006-01-02 15:04:05"),
		contact.UpdatedAt.Format("2006-01-02 15:04:05"),
	)

	return ToolResult{
		Content: []ContentBlock{{Type: "text", Text: result}},
	}, nil
}

// executeUpdateContact updates a contact
func (s *MCPServer) executeUpdateContact(args map[string]interface{}) (ToolResult, error) {
	idFloat, ok := args["id"].(float64)
	if !ok {
		return ToolResult{IsError: true}, fmt.Errorf("contact id is required")
	}
	
	id := uint(idFloat)
	contact, err := s.contactRepo.GetByID(id)
	if err != nil {
		return ToolResult{IsError: true}, fmt.Errorf("contact not found: %v", err)
	}

	// Update fields if provided
	if name, ok := args["name"].(string); ok && name != "" {
		contact.Name = name
	}
	if email, ok := args["email"].(string); ok && email != "" {
		contact.Email = email
	}
	if phone, ok := args["phone"].(string); ok {
		contact.Phone = phone
	}
	if company, ok := args["company"].(string); ok {
		contact.Company = company
	}
	if position, ok := args["position"].(string); ok {
		contact.Position = position
	}
	if status, ok := args["status"].(string); ok && status != "" {
		contact.Status = status
	}
	if notes, ok := args["notes"].(string); ok {
		contact.Notes = notes
	}

	if err := s.contactRepo.Update(contact); err != nil {
		return ToolResult{IsError: true}, fmt.Errorf("failed to update contact: %v", err)
	}

	result := fmt.Sprintf("Successfully updated contact: %s (ID: %d)", contact.Name, contact.ID)
	return ToolResult{
		Content: []ContentBlock{{Type: "text", Text: result}},
	}, nil
}

// executeDeleteContact deletes a contact
func (s *MCPServer) executeDeleteContact(args map[string]interface{}) (ToolResult, error) {
	idFloat, ok := args["id"].(float64)
	if !ok {
		return ToolResult{IsError: true}, fmt.Errorf("contact id is required")
	}
	
	id := uint(idFloat)
	
	// Get contact details before deletion for confirmation
	contact, err := s.contactRepo.GetByID(id)
	if err != nil {
		return ToolResult{IsError: true}, fmt.Errorf("contact not found: %v", err)
	}

	if err := s.contactRepo.Delete(id); err != nil {
		return ToolResult{IsError: true}, fmt.Errorf("failed to delete contact: %v", err)
	}

	result := fmt.Sprintf("Successfully deleted contact: %s (ID: %d)", contact.Name, contact.ID)
	return ToolResult{
		Content: []ContentBlock{{Type: "text", Text: result}},
	}, nil
}

// executeGetAnalytics gets analytics data
func (s *MCPServer) executeGetAnalytics(args map[string]interface{}) (ToolResult, error) {
	startDateStr, _ := args["start_date"].(string)
	endDateStr, _ := args["end_date"].(string)
	granularity, _ := args["granularity"].(string)

	// Default to last 30 days if no dates provided
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -30)

	if startDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", startDateStr); err == nil {
			startDate = parsed
		}
	}
	if endDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", endDateStr); err == nil {
			endDate = parsed
		}
	}
	if granularity == "" {
		granularity = "daily"
	}

	request := &models.AnalyticsRequest{
		StartDate:   startDate,
		EndDate:     endDate,
		Granularity: granularity,
		Metrics:     []string{"contacts", "appointments"},
	}

	// Note: This would need the analytics service method to be implemented
	// For now, we'll provide a simplified response
	result := fmt.Sprintf(`Analytics Summary (%s to %s):

Contact Metrics:
• Period: %s
• Granularity: %s

Note: Detailed analytics implementation depends on the specific analytics service methods being available.
For full analytics, use the REST API endpoints directly.`,
		startDate.Format("2006-01-02"),
		endDate.Format("2006-01-02"),
		startDate.Format("2006-01-02")+" to "+endDate.Format("2006-01-02"),
		granularity,
	)

	return ToolResult{
		Content: []ContentBlock{{Type: "text", Text: result}},
	}, nil
}

// executeExportContacts exports contacts
func (s *MCPServer) executeExportContacts(args map[string]interface{}) (ToolResult, error) {
	format, _ := args["format"].(string)
	status, _ := args["status"].(string)
	limitFloat, _ := args["limit"].(float64)
	limit := int(limitFloat)

	if format == "" {
		format = "csv"
	}
	if limit == 0 {
		limit = 1000
	}

	exportRequest := services.ExportRequest{
		Format:    services.ExportFormat(format),
		SortBy:    "created_at",
		SortOrder: "desc",
		Limit:     limit,
		Filters:   make(map[string]interface{}),
	}

	if status != "" {
		exportRequest.Filters["status"] = status
	}

	var data []byte
	var err error

	switch format {
	case "csv":
		data, err = s.bulkService.ExportContactsToCSV(exportRequest)
	case "json":
		data, err = s.bulkService.ExportContactsToJSON(exportRequest)
	default:
		return ToolResult{IsError: true}, fmt.Errorf("unsupported format: %s (use csv or json)", format)
	}

	if err != nil {
		return ToolResult{IsError: true}, fmt.Errorf("failed to export contacts: %v", err)
	}

	result := fmt.Sprintf("Successfully exported %d bytes in %s format. Export filters: status=%s, limit=%d",
		len(data), format, status, limit)

	// In a real implementation, you might want to save this to a file
	// or provide a way to retrieve the exported data
	return ToolResult{
		Content: []ContentBlock{{Type: "text", Text: result}},
	}, nil
}