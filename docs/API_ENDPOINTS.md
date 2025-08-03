# Contact Management Microservice - Complete API Reference

**Base URL**: `http://localhost:8081`  
**API Version**: `v1`  
**Authentication**: Bearer JWT Token (except public endpoints)

## 📋 Complete API Endpoints Overview

### **Status**: ⚠️ **Some compilation issues need to be resolved**

**Total Endpoints**: 89 endpoints across 11 modules
- ✅ **Working**: Health checks, MCP server
- ⚠️ **Needs fixes**: Analytics service type mismatches, missing handler implementations

---

## 🏥 Health Check Endpoints (6 endpoints)

### 1. Basic Health Check
```
GET /health
```
**Response**:
```json
{
  "status": "healthy",
  "timestamp": "2024-01-15T10:30:00Z",
  "version": "1.0.0",
  "uptime": "2h45m30s"
}
```

### 2. Deep Health Check
```
GET /health/deep
```
**Response**:
```json
{
  "status": "healthy",
  "timestamp": "2024-01-15T10:30:00Z",
  "services": {
    "database": {
      "status": "healthy",
      "response_time": "2ms",
      "connections": {
        "active": 5,
        "idle": 10,
        "max": 100
      }
    },
    "cache": {
      "status": "healthy",
      "response_time": "1ms"
    }
  },
  "system": {
    "memory_usage": "45%",
    "cpu_usage": "12%",
    "disk_usage": "67%"
  }
}
```

### 3. Readiness Check
```
GET /ready
```

### 4. Liveness Check
```
GET /alive
```

### 5. Status Check
```
GET /status
```

### 6. Metrics Check
```
GET /metrics
```

---

## 🔐 Authentication Endpoints (6 endpoints)

### 1. User Login
```
POST /api/v1/auth/login
```
**Request**:
```json
{
  "email": "admin@mejona.com",
  "password": "secure_password"
}
```
**Response**:
```json
{
  "success": true,
  "message": "Login successful",
  "data": {
    "user": {
      "id": 1,
      "email": "admin@mejona.com",
      "name": "Admin User",
      "role": "admin"
    },
    "tokens": {
      "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
      "refresh_token": "eyJhbGciOiJIUzI1NiIsInN5cCI6IkpXVCJ9...",
      "expires_in": 3600
    }
  }
}
```

### 2. Refresh Token
```
POST /api/v1/auth/refresh
```

### 3. Logout
```
POST /api/v1/auth/logout
Authorization: Bearer <token>
```

### 4. Get Profile
```
GET /api/v1/auth/profile
Authorization: Bearer <token>
```

### 5. Change Password
```
POST /api/v1/auth/change-password
Authorization: Bearer <token>
```

### 6. Validate Token
```
GET /api/v1/auth/validate
Authorization: Bearer <token>
```

---

## 👥 Contact Management Endpoints (8 endpoints)

### 1. List Contacts
```
GET /api/v1/contacts?page=1&limit=10&status=active&sort=created_at&order=desc
Authorization: Bearer <token>
```
**Response**:
```json
{
  "success": true,
  "message": "Contacts retrieved successfully",
  "data": {
    "contacts": [
      {
        "id": 1,
        "name": "John Doe",
        "email": "john.doe@example.com",
        "phone": "+1-555-123-4567",
        "company": "Acme Corp",
        "position": "Manager",
        "status": "qualified",
        "created_at": "2024-01-15T10:30:00Z",
        "updated_at": "2024-01-15T12:45:00Z"
      }
    ],
    "pagination": {
      "current_page": 1,
      "per_page": 10,
      "total": 156,
      "total_pages": 16
    }
  }
}
```

### 2. Create Contact
```
POST /api/v1/contacts
Authorization: Bearer <token>
```
**Request**:
```json
{
  "name": "Jane Smith",
  "email": "jane.smith@techcorp.com",
  "phone": "+1-555-987-6543",
  "company": "TechCorp",
  "position": "CTO",
  "status": "new",
  "notes": "Interested in AI solutions"
}
```
**Response**:
```json
{
  "success": true,
  "message": "Contact created successfully",
  "data": {
    "id": 157,
    "name": "Jane Smith",
    "email": "jane.smith@techcorp.com",
    "phone": "+1-555-987-6543",
    "company": "TechCorp",
    "position": "CTO", 
    "status": "new",
    "notes": "Interested in AI solutions",
    "created_at": "2024-01-15T14:30:00Z",
    "updated_at": "2024-01-15T14:30:00Z"
  }
}
```

### 3. Get Contact
```
GET /api/v1/contacts/1
Authorization: Bearer <token>
```

### 4. Update Contact
```
PUT /api/v1/contacts/1
Authorization: Bearer <token>
```

### 5. Delete Contact
```
DELETE /api/v1/contacts/1
Authorization: Bearer <token>
```

### 6. Update Contact Status
```
PUT /api/v1/contacts/1/status
Authorization: Bearer <token>
```
**Request**:
```json
{
  "status": "qualified",
  "notes": "Passed initial screening"
}
```

### 7. Search Contacts
```
GET /api/v1/contacts/search?q=john&status=qualified&company=acme
Authorization: Bearer <token>
```

### 8. Get Contact Appointments
```
GET /api/v1/contacts/1/appointments
Authorization: Bearer <token>
```

---

## 📤 Bulk Operations Endpoints (6 endpoints)

### 1. Import Contacts from CSV
```
POST /api/v1/bulk/contacts/import
Authorization: Bearer <token>
Content-Type: multipart/form-data
```
**Request**: CSV file upload
**Response**:
```json
{
  "success": true,
  "message": "Import completed with 2 errors out of 100 records",
  "data": {
    "total_records": 100,
    "success_count": 98,
    "error_count": 2,
    "imported_ids": [158, 159, 160, ...],
    "processing_time": "2.5s",
    "errors": [
      {
        "row": 45,
        "field": "email",
        "value": "invalid-email",
        "message": "Invalid email format"
      }
    ],
    "summary": {
      "success_rate": 98.0,
      "processing_speed": "40 records/second",
      "duplicate_emails": 1,
      "validation_errors": 1
    }
  }
}
```

### 2. Export Contacts
```
GET /api/v1/bulk/contacts/export?format=csv&status=qualified&limit=1000
Authorization: Bearer <token>
```
**Response**: CSV file download

### 3. Get Import Template
```
GET /api/v1/bulk/contacts/template
```
**Response**: CSV template file

### 4. Bulk Update Contacts
```
POST /api/v1/bulk/contacts/update
Authorization: Bearer <token>
```
**Request**:
```json
{
  "contact_ids": [1, 2, 3, 4, 5],
  "updates": {
    "status": "qualified",
    "notes": "Bulk updated during review"
  },
  "conditions": {
    "status": "new"
  }
}
```

### 5. Bulk Delete Contacts
```
POST /api/v1/bulk/contacts/delete
Authorization: Bearer <token>
```

### 6. Bulk Operation Status
```
GET /api/v1/bulk/contacts/status
Authorization: Bearer <token>
```

---

## 🔍 Advanced Search Endpoints (6 endpoints)

### 1. Advanced Search
```
GET /api/v1/search/contacts/advanced
Authorization: Bearer <token>
```
**Query Parameters**:
- `name`, `email`, `company`, `position`
- `status`, `created_after`, `created_before`
- `has_phone`, `has_notes`, `assigned_to`
- `lead_score_min`, `lead_score_max`

### 2. Search Suggestions
```
GET /api/v1/search/suggestions?q=john
Authorization: Bearer <token>
```

### 3. Saved Searches
```
GET /api/v1/search/saved
Authorization: Bearer <token>
```

### 4. Save Search
```
POST /api/v1/search/saved
Authorization: Bearer <token>
```

### 5. Delete Saved Search
```
DELETE /api/v1/search/saved/1
Authorization: Bearer <token>
```

### 6. Execute Saved Search
```
GET /api/v1/search/saved/1/execute
Authorization: Bearer <token>
```

---

## 🌐 Public Endpoints (1 endpoint)

### 1. Submit Contact (No Authentication)
```
POST /api/v1/public/contact
```
**Request**:
```json
{
  "name": "Lead Contact",
  "email": "lead@company.com",
  "phone": "+1-555-000-0000",
  "company": "New Company",
  "message": "Interested in your services",
  "source": "website"
}
```
**Response**:
```json
{
  "success": true,
  "message": "Contact submitted successfully. We'll get back to you soon!",
  "data": {
    "reference_id": "CNT-2024-001234",
    "status": "submitted"
  }
}
```

---

## ⚙️ Admin Configuration Endpoints (8 endpoints)

### 1. Contact Types Management
```
GET /api/v1/admin/contact-types
POST /api/v1/admin/contact-types
PUT /api/v1/admin/contact-types/1
DELETE /api/v1/admin/contact-types/1
Authorization: Bearer <admin-token>
```

### 2. Contact Sources Management
```
GET /api/v1/admin/contact-sources
POST /api/v1/admin/contact-sources  
PUT /api/v1/admin/contact-sources/1
DELETE /api/v1/admin/contact-sources/1
Authorization: Bearer <admin-token>
```

---

## 👤 Assignment Management Endpoints (11 endpoints)

### 1. Auto Assign Contact
```
POST /api/v1/assignments/auto/1
Authorization: Bearer <token>
```

### 2. Manual Assignment
```
POST /api/v1/assignments/manual
Authorization: Bearer <token>
```
**Request**:
```json
{
  "contact_id": 1,
  "user_id": 5,
  "reason": "Subject matter expertise"
}
```

### 3. Bulk Assignment
```
POST /api/v1/assignments/bulk
Authorization: Bearer <token>
```

### 4. Unassign Contact
```
POST /api/v1/assignments/unassign/1
Authorization: Bearer <token>
```

### 5. Reassign Contact
```
POST /api/v1/assignments/reassign/1
Authorization: Bearer <token>
```

### 6. Get User Workload
```
GET /api/v1/assignments/workload/5
Authorization: Bearer <token>
```

### 7. Get My Workload
```
GET /api/v1/assignments/my-workload
Authorization: Bearer <token>
```

### 8. Get All Workloads
```
GET /api/v1/assignments/workloads
Authorization: Bearer <token>
```

### 9. Assignment History
```
GET /api/v1/assignments/history/1
Authorization: Bearer <token>
```

### 10. My Assignments
```
GET /api/v1/assignments/my-assignments
Authorization: Bearer <token>
```

### 11. Accept Assignment
```
POST /api/v1/assignments/1/accept
Authorization: Bearer <token>
```

---

## 📋 Assignment Rules (Admin) Endpoints (7 endpoints)

### 1. Create Assignment Rule
```
POST /api/v1/assignment-rules
Authorization: Bearer <admin-token>
```

### 2. Get Assignment Rules
```
GET /api/v1/assignment-rules
Authorization: Bearer <admin-token>
```

### 3. Get Assignment Rule
```
GET /api/v1/assignment-rules/1
Authorization: Bearer <admin-token>
```

### 4. Update Assignment Rule
```
PUT /api/v1/assignment-rules/1
Authorization: Bearer <admin-token>
```

### 5. Delete Assignment Rule
```
DELETE /api/v1/assignment-rules/1
Authorization: Bearer <admin-token>
```

### 6. Toggle Assignment Rule
```
POST /api/v1/assignment-rules/1/toggle
Authorization: Bearer <admin-token>
```

### 7. Test Assignment Rule
```
POST /api/v1/assignment-rules/1/test
Authorization: Bearer <admin-token>
```

---

## 🔄 Lifecycle Management Endpoints (13 endpoints)

### 1. Score Contact
```
POST /api/v1/lifecycle/score
Authorization: Bearer <token>
```

### 2. Score Contact by ID
```
POST /api/v1/lifecycle/score/1
Authorization: Bearer <token>
```

### 3. Change Contact Status
```
POST /api/v1/lifecycle/status/change
Authorization: Bearer <token>
```

### 4. Change Status by ID
```
POST /api/v1/lifecycle/status/1/change
Authorization: Bearer <token>
```

### 5. Bulk Status Change
```
POST /api/v1/lifecycle/status/bulk-change
Authorization: Bearer <token>
```

### 6. Get Contact Lifecycle
```
GET /api/v1/lifecycle/1
Authorization: Bearer <token>
```

### 7. Get Lifecycle Events
```
GET /api/v1/lifecycle/1/events
Authorization: Bearer <token>
```
**Response**:
```json
{
  "success": true,
  "data": {
    "contact_id": 1,
    "events": [
      {
        "id": 1,
        "event_type": "status_change",
        "from_status": "new",
        "to_status": "qualified",
        "timestamp": "2024-01-15T10:30:00Z",
        "user_id": 2,
        "notes": "Passed initial screening"
      },
      {
        "id": 2,
        "event_type": "score_update",
        "old_score": 65,
        "new_score": 85,
        "timestamp": "2024-01-15T14:22:00Z",
        "reason": "Engagement increased"
      }
    ]
  }
}
```

### 8. Analyze Scoring
```
GET /api/v1/lifecycle/1/analyze
Authorization: Bearer <token>
```

### 9-13. Scoring & Transition Rules (Admin Only)
- Create/Get/Update/Delete Scoring Rules
- Create/Get/Update/Delete Transition Rules

---

## 📅 Appointment Management Endpoints (12 endpoints)

### 1. Create Appointment
```
POST /api/v1/appointments
Authorization: Bearer <token>
```
**Request**:
```json
{
  "contact_id": 1,
  "title": "Sales Consultation",
  "description": "Discuss AI solutions",
  "start_time": "2024-01-20T14:00:00Z",
  "end_time": "2024-01-20T15:00:00Z",
  "type": "consultation",
  "meeting_type": "video_call",
  "meeting_url": "https://meet.google.com/abc-defg-hij"
}
```

### 2. Get Appointment
```
GET /api/v1/appointments/1
Authorization: Bearer <token>
```

### 3. Update Appointment
```
PUT /api/v1/appointments/1
Authorization: Bearer <token>
```

### 4. Update Appointment Status
```
PUT /api/v1/appointments/1/status
Authorization: Bearer <token>
```

### 5. Reschedule Appointment
```
POST /api/v1/appointments/1/reschedule
Authorization: Bearer <token>
```

### 6. Cancel Appointment
```
POST /api/v1/appointments/1/cancel
Authorization: Bearer <token>
```

### 7. Get User Appointments
```
GET /api/v1/appointments/user?user_id=2&date=2024-01-20
Authorization: Bearer <token>
```

### 8. Get My Appointments
```
GET /api/v1/appointments/my?date=2024-01-20&status=confirmed
Authorization: Bearer <token>
```

### 9. Today's Appointments
```
GET /api/v1/appointments/today
Authorization: Bearer <token>
```

### 10. Upcoming Appointments
```
GET /api/v1/appointments/upcoming?days=7
Authorization: Bearer <token>
```

### 11. Find Available Slots
```
POST /api/v1/appointments/available-slots
Authorization: Bearer <token>
```
**Request**:
```json
{
  "user_id": 2,
  "start_date": "2024-01-20",
  "end_date": "2024-01-27",
  "duration": 60,
  "buffer_time": 15,
  "business_hours_only": true
}
```

### 12. Get User Availability
```
GET /api/v1/appointments/availability?user_id=2&date=2024-01-20
Authorization: Bearer <token>
```

---

## 📊 Analytics Endpoints (10 endpoints)

### 1. Contact Analytics
```
GET /api/v1/analytics/contacts?start_date=2024-01-01&end_date=2024-01-31&granularity=daily
Authorization: Bearer <token>
```
**Response**:
```json
{
  "success": true,
  "data": {
    "total_contacts": 1250,
    "new_contacts": 145,
    "active_contacts": 890,
    "converted_contacts": 67,
    "conversion_rate": 5.36,
    "by_status": {
      "new": 234,
      "qualified": 456,
      "contacted": 123,
      "converted": 67,
      "inactive": 370
    },
    "by_source": {
      "website": 450,
      "referral": 230,
      "social_media": 180,
      "email_campaign": 145,
      "other": 245
    },
    "timeline": [
      {
        "date": "2024-01-01",
        "new_contacts": 5,
        "conversions": 2
      }
    ]
  }
}
```

### 2. Appointment Analytics
```
GET /api/v1/analytics/appointments
Authorization: Bearer <token>
```

### 3. User Performance Analytics
```
GET /api/v1/analytics/performance?user_id=2
Authorization: Bearer <token>
```

### 4. Conversion Metrics
```
GET /api/v1/analytics/conversion
Authorization: Bearer <token>
```

### 5. Source Analytics
```
GET /api/v1/analytics/sources
Authorization: Bearer <token>
```

### 6. Response Time Metrics
```
GET /api/v1/analytics/response-times
Authorization: Bearer <token>
```

### 7. Realtime Metrics
```
GET /api/v1/analytics/realtime
Authorization: Bearer <token>
```

### 8. Dashboard Summary
```
GET /api/v1/analytics/dashboard
Authorization: Bearer <token>
```
**Response**:
```json
{
  "success": true,
  "data": {
    "quick_stats": {
      "total_contacts": 1250,
      "today_new_contacts": 12,
      "week_growth": 8.5,
      "conversion_rate": 5.36,
      "average_response_time": 2.4,
      "active_deals": 23,
      "total_revenue": 125000,
      "monthly_target": 150000,
      "target_progress": 83.33
    },
    "recent_activity": [
      {
        "type": "contact_created",
        "contact": "John Doe",
        "timestamp": "2024-01-15T14:30:00Z"
      }
    ],
    "alerts": [
      {
        "type": "high_value_lead",
        "message": "New high-value lead from TechCorp",
        "priority": "high"
      }
    ]
  }
}
```

### 9. Business Intelligence
```
GET /api/v1/analytics/business-intelligence
Authorization: Bearer <token>
```

### 10. Analytics Export
```
GET /api/v1/analytics/export?format=csv&metrics=contacts,appointments
Authorization: Bearer <token>
```

---

## 🖥️ Monitoring Endpoints (9 endpoints)

### 1. System Health
```
GET /api/v1/monitoring/health
Authorization: Bearer <token>
```
**Response**:
```json
{
  "success": true,
  "data": {
    "overall_status": "healthy",
    "services": {
      "database": {
        "status": "healthy",
        "response_time": "2ms",
        "connections": 15,
        "max_connections": 100
      },
      "api": {
        "status": "healthy",
        "requests_per_minute": 245,
        "average_response_time": "45ms"
      }
    },
    "system_resources": {
      "memory_usage": 67.5,
      "cpu_usage": 23.1,
      "disk_usage": 45.8
    }
  }
}
```

### 2. Error Statistics
```
GET /api/v1/monitoring/errors
Authorization: Bearer <token>
```

### 3. System Metrics
```
GET /api/v1/monitoring/metrics
Authorization: Bearer <token>
```

### 4. Active Alerts
```
GET /api/v1/monitoring/alerts
Authorization: Bearer <token>
```

### 5. Create Alert
```
POST /api/v1/monitoring/alerts
Authorization: Bearer <token>
```

### 6. Acknowledge Alert
```
POST /api/v1/monitoring/alerts/1/acknowledge
Authorization: Bearer <token>
```

### 7. Track Error
```
POST /api/v1/monitoring/errors/track
Authorization: Bearer <token>
```

### 8. Track Metric
```
POST /api/v1/monitoring/metrics/track
Authorization: Bearer <token>
```

### 9. Get System Logs
```
GET /api/v1/monitoring/logs?level=error&limit=100
Authorization: Bearer <token>
```

---

## 🤖 MCP Server Integration

The MCP (Model Context Protocol) server provides AI assistants with tools to interact with the contact management system:

### Available Tools:
1. **create_contact** - Create new contacts
2. **search_contacts** - Search and filter contacts  
3. **get_contact** - Get contact details
4. **update_contact** - Update contact information
5. **delete_contact** - Delete contacts
6. **get_analytics** - Contact analytics
7. **export_contacts** - Export contact data

### Usage:
```bash
# Start MCP server
make mcp-run

# Test MCP server
make mcp-test
```

---

## ⚠️ Current Issues & Status

### **Compilation Issues to Fix:**
1. **Analytics Service**: Type mismatches in GORM Count operations
2. **Missing Handlers**: Some handler functions need implementation
3. **Model Relationships**: Some model references need cleanup

### **Working Components:**
- ✅ Health check endpoints
- ✅ MCP server implementation 
- ✅ Database models and migrations
- ✅ Basic CRUD operations structure
- ✅ Authentication middleware
- ✅ Bulk operations service

### **Next Steps:**
1. Fix analytics service type issues
2. Complete missing handler implementations
3. Test all endpoints with database
4. Add comprehensive error handling
5. Complete integration testing

---

## 🚀 Getting Started

1. **Setup Database**:
```bash
make migrate-up
```

2. **Start Service**:
```bash
make run
```

3. **View Documentation**:
```bash
make docs-serve
# Visit: http://localhost:8081/swagger/index.html
```

4. **Test APIs**:
```bash
# Health check
curl http://localhost:8081/health

# Login (get token first)
curl -X POST http://localhost:8081/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@mejona.com","password":"password"}'
```

The service provides a comprehensive contact management solution with 89 endpoints covering all aspects of contact lifecycle management, from creation to analytics and AI integration.