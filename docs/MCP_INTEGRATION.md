# Contact Management MCP Server

This document describes the Model Context Protocol (MCP) server implementation for the Contact Management microservice, enabling AI assistants to interact with the contact management system.

## Overview

The MCP server provides AI assistants with tools to:
- Create, read, update, and delete contacts
- Search and filter contacts
- Export contact data
- Access analytics and metrics
- Perform bulk operations

## Installation & Setup

### Prerequisites
- Go 1.21+
- MySQL database
- Contact Management microservice configured

### Configuration

1. **Environment Variables**: Set up the same environment variables as the main service:
```bash
DB_HOST=localhost
DB_PORT=3306
DB_NAME=mejona_contacts
DB_USER=root
DB_PASSWORD=your_password
APP_DEBUG=false
```

2. **MCP Client Configuration**: Add to your MCP client config (e.g., Claude Desktop):
```json
{
  "mcpServers": {
    "contact-management": {
      "command": "go",
      "args": ["run", "./cmd/mcp-server/main.go"],
      "cwd": "/path/to/contact-service",
      "env": {
        "DB_HOST": "localhost",
        "DB_PORT": "3306",
        "DB_NAME": "mejona_contacts",
        "DB_USER": "root",
        "DB_PASSWORD": "your_password"
      }
    }
  }
}
```

### Running the MCP Server

```bash
# From the contact-service directory
cd cmd/mcp-server
go run main.go
```

## Available Tools

### 1. create_contact
Creates a new contact in the system.

**Parameters:**
- `name` (required): Full name of the contact
- `email` (required): Email address
- `phone` (optional): Phone number
- `company` (optional): Company name
- `position` (optional): Job position/title
- `notes` (optional): Additional notes

**Example:**
```json
{
  "name": "John Doe",
  "email": "john.doe@example.com",
  "phone": "+1-555-123-4567",
  "company": "Acme Corp",
  "position": "Manager",
  "notes": "Interested in web development services"
}
```

### 2. search_contacts
Search for contacts by various criteria.

**Parameters:**
- `query` (optional): Search query (searches name, email, company)
- `status` (optional): Filter by contact status
- `limit` (optional): Maximum number of results (default: 10)

**Example:**
```json
{
  "query": "john",
  "status": "new",
  "limit": 20
}
```

### 3. get_contact
Get detailed information about a specific contact.

**Parameters:**
- `id` (optional): Contact ID
- `email` (optional): Contact email address

Note: Either `id` or `email` must be provided.

**Example:**
```json
{
  "id": 123
}
```

### 4. update_contact
Update an existing contact's information.

**Parameters:**
- `id` (required): Contact ID
- `name` (optional): Updated name
- `email` (optional): Updated email
- `phone` (optional): Updated phone
- `company` (optional): Updated company
- `position` (optional): Updated position
- `status` (optional): Updated status
- `notes` (optional): Updated notes

**Example:**
```json
{
  "id": 123,
  "status": "qualified",
  "notes": "Follow up next week"
}
```

### 5. delete_contact
Delete a contact from the system.

**Parameters:**
- `id` (required): Contact ID to delete

**Example:**
```json
{
  "id": 123
}
```

### 6. get_analytics
Get contact analytics and metrics.

**Parameters:**
- `start_date` (optional): Start date (YYYY-MM-DD format)
- `end_date` (optional): End date (YYYY-MM-DD format)
- `granularity` (optional): Data granularity (daily, weekly, monthly)

**Example:**
```json
{
  "start_date": "2024-01-01",
  "end_date": "2024-01-31",
  "granularity": "weekly"
}
```

### 7. export_contacts
Export contacts to CSV or JSON format.

**Parameters:**
- `format` (optional): Export format ("csv" or "json", default: "csv")
- `status` (optional): Filter by status
- `limit` (optional): Maximum records to export

**Example:**
```json
{
  "format": "csv",
  "status": "qualified",
  "limit": 500
}
```

## Usage Examples

### AI Assistant Interactions

**Creating a contact:**
> "Create a new contact for Sarah Johnson at sarah@techcorp.com, she works at TechCorp as a CTO and is interested in our AI solutions."

**Searching contacts:**
> "Find all contacts from TechCorp that have a 'qualified' status."

**Updating contact status:**
> "Update contact ID 123 to have status 'converted' and add a note that they signed a contract."

**Getting analytics:**
> "Show me contact analytics for the last month with weekly granularity."

**Bulk operations:**
> "Export all qualified contacts to CSV format."

## Error Handling

The MCP server provides detailed error messages for:
- Invalid parameters
- Database connection issues
- Missing required fields
- Record not found errors
- Permission errors

## Security Considerations

1. **Database Access**: The MCP server has full database access - ensure it runs in a secure environment
2. **Input Validation**: All inputs are validated before database operations
3. **Error Disclosure**: Error messages are sanitized to prevent information leakage
4. **Environment Variables**: Store sensitive configuration in environment variables

## Development

### Extending Functionality

To add new tools:

1. Add the tool definition to `handleToolsList()`
2. Implement the execution function following the pattern `execute{ToolName}()`
3. Add the case to `handleToolCall()`

### Testing

```bash
# Test the MCP server with sample requests
echo '{"method":"tools/list","params":{},"id":"1"}' | go run cmd/mcp-server/main.go
```

### Integration with Other Services

The MCP server can be extended to integrate with other microservices in the admin dashboard:
- User management service
- Blog management service  
- HR management service
- Analytics service

## Troubleshooting

### Common Issues

1. **Database Connection Failed**
   - Verify database credentials in environment variables
   - Ensure MySQL server is running
   - Check network connectivity

2. **Tool Not Found**
   - Verify tool name spelling
   - Check if tool is registered in `handleToolsList()`

3. **Invalid Parameters**
   - Review tool input schema requirements
   - Ensure required parameters are provided

### Logging

Enable debug logging by setting:
```bash
APP_DEBUG=true
```

## Performance Considerations

- Search operations are optimized with database indexes
- Large export operations may take time - consider implementing streaming for very large datasets
- Connection pooling is configured for optimal database performance

## Future Enhancements

Planned improvements:
- WebSocket support for real-time updates
- Batch operations for multiple contacts
- Advanced filtering and sorting options
- Integration with calendar systems for appointment scheduling
- Webhook notifications for contact updates