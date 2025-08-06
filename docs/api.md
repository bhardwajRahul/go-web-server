# API Reference

HTTP endpoints and HTMX integration for the Modern Go Stack web server.

## Base URL

```
http://localhost:8080
```

## System Endpoints

### Health Check

**GET /health**

Returns system health with database connectivity check.

```json
{
  "status": "ok|warning|error",
  "timestamp": "2024-01-01T12:00:00Z",
  "service": "go-web-server",
  "version": "1.0.0",
  "uptime": "1h23m",
  "checks": {
    "database": "ok",
    "database_connections": "ok"
  }
}
```

### Metrics (Development)

**GET /metrics**

Prometheus metrics endpoint (enabled with `FEATURES_ENABLE_METRICS=true`).

### Static Assets

**GET /static/***

Embedded static files:

- `/static/css/pico.min.css` - Pico.css v2 framework
- `/static/js/htmx.min.js` - HTMX 2.0.6 library
- `/static/favicon.ico` - Application favicon

## Page Endpoints

### Home Page

**GET /**

Main landing page with HTMX demo functionality.

### User Management

**GET /users**

User management interface with real-time updates.

## User API

### List Users (HTMX Fragment)

**GET /users/list**

Returns HTML table fragment for HTMX updates.

### User Count

**GET /api/users/count**

Returns plain text user count: `42`

### User Forms

**GET /users/form** - New user creation form
**GET /users/:id/edit** - Edit form with user data

Both set `X-CSRF-Token` header for form submission.

### Create User

**POST /users**

Creates user with CSRF protection.

**Form Data:**

```
name=John Doe (required)
email=john@example.com (required)
bio=Optional bio text
avatar_url=https://example.com/avatar.jpg (optional)
csrf_token=abc123... (required)
```

**Success:** Returns updated user list HTML
**Events:** Triggers `userCreated` HTMX event

### Update User

**PUT /users/:id**

Updates existing user (email cannot be changed).

**Form Data:** Same as create (name required, csrf_token required)
**Success:** Returns updated user list HTML
**Events:** Triggers `userUpdated` HTMX event

### Deactivate User

**PATCH /users/:id/deactivate**

Soft delete (sets `is_active = false`).

**Headers:** `X-CSRF-Token` required
**Success:** Returns updated user row HTML
**Events:** Triggers `userDeactivated` HTMX event

### Delete User

**DELETE /users/:id**

Permanent user deletion.

**Headers:** `X-CSRF-Token` required
**Success:** `200 OK` empty response (removes row)
**Events:** Triggers `userDeleted` HTMX event

## CSRF Protection

### Token Flow

1. GET requests automatically set CSRF cookie
2. Forms include hidden `csrf_token` field
3. HTMX requests use `X-CSRF-Token` header
4. Server validates and rotates tokens

### Example Usage

**Form:**

```html
<form hx-post="/users" hx-target="#user-list">
  <input type="hidden" name="csrf_token" value="{{.CSRFToken}}" />
  <input name="name" required />
  <button type="submit">Create</button>
</form>
```

**HTMX Button:**

```html
<button 
  hx-delete="/users/123"
  hx-confirm="Delete user?"
  hx-target="#user-123"
  hx-swap="outerHTML"
>Delete</button>
```

## Error Responses

```json
{
  "type": "validation|not_found|internal|csrf",
  "error": "Bad Request",
  "message": "Validation failed",
  "code": 400,
  "path": "/users",
  "method": "POST",
  "request_id": "uuid",
  "timestamp": "2024-01-01T12:00:00Z"
}
```

## Security Features

- **Rate Limiting:** 20 requests/minute per IP
- **CSRF Protection:** All state-changing operations protected
- **Input Sanitization:** XSS and SQL injection prevention
- **Security Headers:** CSP, HSTS, X-Frame-Options
- **Request Tracing:** Unique ID per request
