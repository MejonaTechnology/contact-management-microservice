# Contact Management Microservice API Documentation

## Overview

The Contact Management Microservice API provides comprehensive functionality for managing customer contacts, including CRUD operations, advanced search, assignment management, lifecycle tracking, scheduling, analytics, and system monitoring.

## API Documentation Access

### Swagger UI
- **Development**: http://localhost:8081/swagger/index.html
- **Production**: https://api.mejona.com/swagger/index.html

### OpenAPI Specification
- **YAML Format**: [swagger.yaml](./swagger.yaml)
- **JSON Format**: Auto-generated via Swagger annotations

## Authentication

### JWT Bearer Token
All protected endpoints require JWT authentication:

```bash
Authorization: Bearer <your-jwt-token>
```

### Getting Started
1. **Login**: POST `/api/v1/auth/login` with email/password
2. **Use Token**: Include the returned `access_token` in Authorization header
3. **Refresh**: Use `/api/v1/auth/refresh` when token expires

## API Endpoints Overview

### 🏥 Health & Monitoring
| Endpoint | Method | Description | Auth Required |
|----------|--------|-------------|---------------|
| `/health` | GET | Comprehensive health check | No |
| `/health/deep` | GET | Deep system analysis | No |
| `/status` | GET | Quick status overview | No |
| `/ready` | GET | Readiness probe | No |
| `/alive` | GET | Liveness probe | No |
| `/metrics` | GET | System metrics | No |

### 🔐 Authentication
| Endpoint | Method | Description | Auth Required |
|----------|--------|-------------|---------------|
| `/api/v1/auth/login` | POST | User login | No |
| `/api/v1/auth/refresh` | POST | Refresh token | No |
| `/api/v1/auth/logout` | POST | User logout | Yes |
| `/api/v1/auth/profile` | GET | Get user profile | Yes |
| `/api/v1/auth/change-password` | POST | Change password | Yes |
| `/api/v1/auth/validate` | GET | Validate token | Yes |

### 👥 Contacts
| Endpoint | Method | Description | Auth Required |
|----------|--------|-------------|---------------|
| `/api/v1/contacts` | GET | List contacts | Yes |
| `/api/v1/contacts` | POST | Create contact | Yes |
| `/api/v1/contacts/{id}` | GET | Get contact | Yes |
| `/api/v1/contacts/{id}` | PUT | Update contact | Yes |
| `/api/v1/contacts/{id}` | DELETE | Delete contact | Yes |
| `/api/v1/contacts/{id}/status` | PUT | Update status | Yes |
| `/api/v1/contacts/search` | GET | Search contacts | Yes |
| `/api/v1/public/contact` | POST | Submit contact | No |

### 🔍 Advanced Search
| Endpoint | Method | Description | Auth Required |
|----------|--------|-------------|---------------|
| `/api/v1/search/contacts/advanced` | GET | Advanced search | Yes |
| `/api/v1/search/suggestions` | GET | Search suggestions | Yes |
| `/api/v1/search/saved` | GET | Saved searches | Yes |
| `/api/v1/search/saved` | POST | Save search | Yes |
| `/api/v1/search/saved/{id}` | DELETE | Delete saved search | Yes |
| `/api/v1/search/saved/{id}/execute` | GET | Execute saved search | Yes |

### 📋 Assignments
| Endpoint | Method | Description | Auth Required |
|----------|--------|-------------|---------------|
| `/api/v1/assignments/auto/{id}` | POST | Auto assign contact | Yes |
| `/api/v1/assignments/manual` | POST | Manual assignment | Yes |
| `/api/v1/assignments/bulk` | POST | Bulk assignments | Yes |
| `/api/v1/assignments/unassign/{id}` | POST | Unassign contact | Yes |
| `/api/v1/assignments/reassign/{id}` | POST | Reassign contact | Yes |
| `/api/v1/assignments/workload/{id}` | GET | Get user workload | Yes |
| `/api/v1/assignments/my-workload` | GET | Get my workload | Yes |
| `/api/v1/assignments/workloads` | GET | Get all workloads | Yes |
| `/api/v1/assignments/history/{id}` | GET | Assignment history | Yes |
| `/api/v1/assignments/my-assignments` | GET | My assignments | Yes |
| `/api/v1/assignments/{id}/accept` | POST | Accept assignment | Yes |

### 🔄 Lifecycle Management
| Endpoint | Method | Description | Auth Required |
|----------|--------|-------------|---------------|
| `/api/v1/lifecycle/score` | POST | Score contact | Yes |
| `/api/v1/lifecycle/score/{id}` | POST | Score contact by ID | Yes |
| `/api/v1/lifecycle/status/change` | POST | Change status | Yes |
| `/api/v1/lifecycle/status/{id}/change` | POST | Change status by ID | Yes |
| `/api/v1/lifecycle/status/bulk-change` | POST | Bulk status change | Yes |
| `/api/v1/lifecycle/{id}` | GET | Get lifecycle info | Yes |
| `/api/v1/lifecycle/{id}/events` | GET | Get lifecycle events | Yes |
| `/api/v1/lifecycle/{id}/analyze` | GET | Analyze scoring | Yes |

### 📅 Scheduling
| Endpoint | Method | Description | Auth Required |
|----------|--------|-------------|---------------|
| `/api/v1/appointments` | POST | Create appointment | Yes |
| `/api/v1/appointments/{id}` | GET | Get appointment | Yes |
| `/api/v1/appointments/{id}` | PUT | Update appointment | Yes |
| `/api/v1/appointments/{id}/status` | PUT | Update status | Yes |
| `/api/v1/appointments/{id}/reschedule` | POST | Reschedule | Yes |
| `/api/v1/appointments/{id}/cancel` | POST | Cancel appointment | Yes |
| `/api/v1/appointments/user` | GET | User appointments | Yes |
| `/api/v1/appointments/my` | GET | My appointments | Yes |
| `/api/v1/appointments/today` | GET | Today's appointments | Yes |
| `/api/v1/appointments/upcoming` | GET | Upcoming appointments | Yes |
| `/api/v1/appointments/available-slots` | POST | Find available slots | Yes |
| `/api/v1/appointments/availability` | GET | Get availability | Yes |

### 📊 Analytics
| Endpoint | Method | Description | Auth Required |
|----------|--------|-------------|---------------|
| `/api/v1/analytics/contacts` | GET | Contact analytics | Yes |
| `/api/v1/analytics/appointments` | GET | Appointment analytics | Yes |
| `/api/v1/analytics/performance` | GET | User performance | Yes |
| `/api/v1/analytics/conversion` | GET | Conversion metrics | Yes |
| `/api/v1/analytics/sources` | GET | Source analytics | Yes |
| `/api/v1/analytics/response-times` | GET | Response time metrics | Yes |
| `/api/v1/analytics/realtime` | GET | Realtime metrics | Yes |
| `/api/v1/analytics/dashboard` | GET | Dashboard summary | Yes |
| `/api/v1/analytics/business-intelligence` | GET | Business intelligence | Yes |
| `/api/v1/analytics/export` | GET | Export analytics | Yes |

### 📊 Monitoring
| Endpoint | Method | Description | Auth Required |
|----------|--------|-------------|---------------|
| `/api/v1/monitoring/health` | GET | System health | Yes |
| `/api/v1/monitoring/errors` | GET | Error statistics | Yes |
| `/api/v1/monitoring/metrics` | GET | System metrics | Yes |
| `/api/v1/monitoring/alerts` | GET | Active alerts | Yes |
| `/api/v1/monitoring/alerts` | POST | Create alert | Yes |
| `/api/v1/monitoring/alerts/{id}/acknowledge` | POST | Acknowledge alert | Yes |
| `/api/v1/monitoring/errors/track` | POST | Track error | Yes |
| `/api/v1/monitoring/metrics/track` | POST | Track metric | Yes |
| `/api/v1/monitoring/logs` | GET | System logs | Yes |
| `/api/v1/monitoring/logs/level` | POST | Set log level | Yes |

### ⚙️ Administration
| Endpoint | Method | Description | Auth Required |
|----------|--------|-------------|---------------|
| `/api/v1/admin/contact-types` | GET | Get contact types | Admin |
| `/api/v1/admin/contact-types` | POST | Create contact type | Admin |
| `/api/v1/admin/contact-types/{id}` | PUT | Update contact type | Admin |
| `/api/v1/admin/contact-types/{id}` | DELETE | Delete contact type | Admin |
| `/api/v1/admin/contact-sources` | GET | Get contact sources | Admin |
| `/api/v1/admin/contact-sources` | POST | Create contact source | Admin |
| `/api/v1/admin/contact-sources/{id}` | PUT | Update contact source | Admin |
| `/api/v1/admin/contact-sources/{id}` | DELETE | Delete contact source | Admin |

## Request/Response Format

### Standard Response Structure
```json
{
  "success": true,
  "message": "Operation completed successfully",
  "data": {
    // Response data here
  },
  "timestamp": "2024-01-15T10:30:00Z"
}
```

### Error Response Structure
```json
{
  "success": false,
  "message": "Error description",
  "error_code": "VALIDATION_ERROR",
  "details": {
    // Additional error details
  },
  "timestamp": "2024-01-15T10:30:00Z"
}
```

### Validation Error Response
```json
{
  "success": false,
  "message": "Validation failed",
  "error_code": "VALIDATION_ERROR",
  "field_errors": [
    {
      "field": "email",
      "message": "email must be a valid email address",
      "code": "email",
      "value": "invalid-email"
    }
  ],
  "timestamp": "2024-01-15T10:30:00Z"
}
```

## Pagination

List endpoints support pagination:

### Query Parameters
- `page`: Page number (default: 1)
- `limit`: Items per page (default: 10, max: 100)
- `sort`: Sort field (default: created_at)
- `order`: Sort order (asc/desc, default: desc)

### Response Format
```json
{
  "success": true,
  "message": "Data retrieved successfully",
  "data": {
    "items": [...],
    "pagination": {
      "current_page": 1,
      "per_page": 10,
      "total": 150,
      "last_page": 15,
      "has_next": true,
      "has_prev": false
    }
  }
}
```

## Filtering & Search

### Basic Filtering
Use query parameters for basic filtering:
```
GET /api/v1/contacts?status=new&type=sales_inquiry&page=1&limit=20
```

### Advanced Search
Use the advanced search endpoint for complex queries:
```json
POST /api/v1/search/contacts/advanced
{
  "filters": {
    "status": ["new", "contacted"],
    "created_after": "2024-01-01T00:00:00Z",
    "company": {
      "operator": "contains",
      "value": "tech"
    }
  },
  "sort": {
    "field": "created_at",
    "order": "desc"
  }
}
```

## Rate Limiting

API endpoints are rate-limited:

- **Public endpoints**: 100 requests per minute
- **Authenticated endpoints**: 1000 requests per minute  
- **Admin endpoints**: 500 requests per minute

Rate limit headers:
```
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 999
X-RateLimit-Reset: 1642248000
```

## Error Codes

| Code | Description |
|------|-------------|
| `VALIDATION_ERROR` | Request validation failed |
| `AUTHENTICATION_ERROR` | Authentication required or failed |
| `AUTHORIZATION_ERROR` | Insufficient permissions |
| `NOT_FOUND` | Resource not found |
| `CONFLICT` | Resource conflict (duplicate) |
| `RATE_LIMIT_EXCEEDED` | Rate limit exceeded |
| `INTERNAL_ERROR` | Internal server error |
| `DATABASE_ERROR` | Database operation failed |
| `TIMEOUT_ERROR` | Request timeout |

## SDKs and Client Libraries

### Available SDKs
- **Go**: [mejona-go-sdk](https://github.com/mejona/mejona-go-sdk)
- **JavaScript/TypeScript**: [mejona-js-sdk](https://github.com/mejona/mejona-js-sdk)
- **Python**: [mejona-python-sdk](https://github.com/mejona/mejona-python-sdk)
- **PHP**: [mejona-php-sdk](https://github.com/mejona/mejona-php-sdk)

### Example Usage (JavaScript)
```javascript
import { MejonaClient } from '@mejona/sdk';

const client = new MejonaClient({
  baseURL: 'http://localhost:8081',
  apiKey: 'your-jwt-token'
});

// Create a contact
const contact = await client.contacts.create({
  name: 'John Doe',
  email: 'john@example.com',
  company: 'Acme Corp'
});

// List contacts
const contacts = await client.contacts.list({
  page: 1,
  limit: 20,
  status: 'new'
});
```

## Support

### Documentation
- **API Reference**: [Swagger UI](http://localhost:8081/swagger/index.html)
- **GitHub Repository**: [mejona-contact-service](https://github.com/mejona/contact-service)
- **Developer Portal**: [developers.mejona.com](https://developers.mejona.com)

### Contact Support
- **Email**: support@mejona.com
- **Documentation Issues**: [GitHub Issues](https://github.com/mejona/contact-service/issues)
- **Developer Community**: [Discord](https://discord.gg/mejona)

### SLA & Availability
- **Uptime**: 99.9% guaranteed
- **Response Time**: < 200ms average
- **Support Hours**: 24/7 for critical issues
- **Maintenance Windows**: Sundays 2-4 AM UTC

---

**© 2024 Mejona Technology LLP. All rights reserved.**