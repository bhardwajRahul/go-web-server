# API Reference

HTTP endpoints and HTMX integration for the Modern Go Stack web server.

## Base URL

```
http://localhost:8080
```

## Core Concepts

### Authentication

No authentication required - this is a demo application showcasing the Modern Go Stack.

### Request Headers

**Required for state-changing operations:**

- `X-CSRF-Token: <token>` (POST/PUT/PATCH/DELETE)

**Common response headers:**

- `X-Request-ID: <uuid>` (request tracing)
- `HX-Trigger: <event>` (HTMX custom events)

### Error Format

Structured error responses with enhanced context:

```json
{
  "type": "validation|not_found|internal|csrf",
  "error": "Bad Request", 
  "message": "Detailed error description",
  "code": 400,
  "path": "/api/endpoint",
  "method": "POST",
  "request_id": "uuid",
  "timestamp": "server-time"
}
```

## System Endpoints

### Health Check

**GET /health**

Comprehensive health check with database connectivity:

```json
{
  "status": "ok|warning|degraded|error",
  "timestamp": "2024-01-01T12:00:00Z",
  "service": "go-web-server", 
  "version": "1.0.0",
  "uptime": "1h23m45s",
  "checks": {
    "database": "ok|warning|error",
    "database_connections": "ok|warning|error",
    "memory": "ok"
  }
}
```

**Status Codes:**

- `200` - All systems operational
- `206` - Partial Content (degraded/warning state)  
- `503` - Service Unavailable (error state)

### Metrics (Development)

**GET /metrics**

Prometheus metrics endpoint (enabled with `FEATURES_ENABLE_METRICS=true`):

**Key Metrics:**

- `http_requests_total` - Request count by method/path/status
- `http_request_duration_seconds` - Request latency histograms
- `database_connections_active` - Active DB connections
- `csrf_tokens_generated_total` - CSRF token generation rate
- `users_created_total` - Business metrics

### Static Assets

**GET /static/\***

Embedded static files served from memory:

- `/static/css/pico.min.css` - Pico.css v2 framework
- `/static/js/htmx.min.js` - HTMX 2.x library
- `/static/favicon.ico` - Application favicon

## Page Endpoints

### Home Page

**GET /**

Main landing page demonstrating Modern Go Stack features.

**HTMX Support:**

- Returns partial content when `HX-Request: true` header present
- Includes live demo functionality

**GET /demo**

Interactive HTMX demonstration endpoint:

```json
{
  "message": "Demo successful! Content loaded with HTMX",
  "features": ["Server-side rendering", "Dynamic loading"],
  "server_time": "3:04:05 PM MST",
  "request_id": "uuid"
}
```

### User Management

**GET /users**

User management interface with real-time updates.

**Features:**

- Live user count updates
- Dynamic table updates via HTMX
- Modal form interactions

## User API

### List Users (HTMX Fragment)

**GET /users/list**

Returns HTML table fragment with active users for HTMX updates.

**Response:** HTML `<table>` element with user rows

### User Count (HTMX Fragment)  

**GET /api/users/count**

Returns plain text user count for real-time updates.

**Response:** `42` (plain text number)

### User Forms

**GET /users/form**
New user creation form.

**GET /users/:id/edit**
Edit form pre-populated with user data.

**Headers:** Both endpoints set `X-CSRF-Token` for form submission.

### Create User

**POST /users**

Creates a new user with validation and CSRF protection.

**Form Data:**

```
name=John Doe (required)
email=john@example.com (required)  
bio=Optional biography
avatar_url=https://example.com/avatar.jpg (optional)
csrf_token=abc123... (required)
```

**Success Response:** Updated user list HTML fragment
**HTMX Events:** Triggers `userCreated` event

### Update User  

**PUT /users/:id**

Updates existing user (email cannot be changed).

**Form Data:**

```
name=Updated Name (required)
bio=Updated biography (optional)
avatar_url=New avatar URL (optional) 
csrf_token=abc123... (required)
```

**Success Response:** Updated user list HTML fragment
**HTMX Events:** Triggers `userUpdated` event

### Deactivate User

**PATCH /users/:id/deactivate**

Soft delete - sets `is_active = false`.

**Headers:** `X-CSRF-Token` required

**Success Response:** Updated user row HTML fragment
**HTMX Events:** Triggers `userDeactivated` event

### Delete User  

**DELETE /users/:id**

Permanent user deletion.

**Headers:** `X-CSRF-Token` required

**Success Response:** `200 OK` (empty body for HTMX removal)
**HTMX Events:** Triggers `userDeleted` event

## HTMX Integration

### Custom Events

The application triggers custom HTMX events for reactive UI updates:

- `userCreated` - After successful user creation
- `userUpdated` - After user information update  
- `userDeactivated` - After user deactivation
- `userDeleted` - After permanent deletion

### CSRF Protection Pattern

1. GET requests automatically set CSRF cookie
2. Forms include hidden `csrf_token` field  
3. HTMX requests use `X-CSRF-Token` header
4. Server validates and rotates tokens on each request

### Example HTMX Usage

**Form Submission:**

```html
<form hx-post="/users" hx-target="#user-list" hx-swap="innerHTML">
  <input type="hidden" name="csrf_token" value="{{.CSRFToken}}" />
  <input name="name" placeholder="Full Name" required />
  <input name="email" placeholder="Email Address" required />
  <button type="submit">Create User</button>
</form>
```

**Dynamic Updates:**

```html  
<button 
  hx-delete="/users/123"
  hx-confirm="Delete this user?"
  hx-target="#user-123" 
  hx-swap="outerHTML"
  hx-headers='{"X-CSRF-Token": "{{.CSRFToken}}"}'
>
  Delete
</button>
```

**Auto-updating Content:**

```html
<span 
  id="user-count"
  hx-get="/api/users/count"
  hx-trigger="load, userCreated from:body, userDeleted from:body"
  hx-swap="innerHTML"
>
  Loading...
</span>
```

## Security Features

### Request Protection

- **Rate Limiting:** 20 requests/minute per IP address
- **Input Sanitization:** XSS and SQL injection prevention  
- **CSRF Protection:** All state-changing operations protected
- **Request Tracing:** Unique ID per request for monitoring

### Security Headers

Comprehensive security headers applied to all responses:

- `X-Content-Type-Options: nosniff`
- `X-Frame-Options: DENY`
- `X-XSS-Protection: 1; mode=block`
- `Strict-Transport-Security` (production)
- `Content-Security-Policy` (HTMX-compatible)

### Error Handling

- **Development:** Detailed error information with stack traces
- **Production:** Sanitized responses without sensitive data
- **Logging:** All errors logged with request context

## Response Patterns

### Successful Operations

- **200 OK** - Standard success
- **201 Created** - Resource created (not used in current API)
- **204 No Content** - Success with no response body

### Client Errors  

- **400 Bad Request** - Validation errors
- **403 Forbidden** - CSRF validation failure
- **404 Not Found** - Resource not found
- **405 Method Not Allowed** - Invalid HTTP method
- **429 Too Many Requests** - Rate limit exceeded

### Server Errors

- **500 Internal Server Error** - Unexpected server error
- **503 Service Unavailable** - Health check failure

This API is designed for the Modern Go Stack with HTMX-first interactions, comprehensive security, and excellent developer experience.
