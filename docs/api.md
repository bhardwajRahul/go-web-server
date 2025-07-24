# API Documentation

This document provides comprehensive API documentation for the Go Web Server.

## Base URL

```
http://localhost:8080
```

## Authentication

Currently, the API does not require authentication. All endpoints are publicly accessible.

## Headers

### Required Headers

- `Content-Type: application/x-www-form-urlencoded` (for form submissions)
- `X-CSRF-Token: <token>` (for POST/PUT/PATCH/DELETE requests)

### Response Headers

- `X-Request-ID: <uuid>` (for request tracing)
- `HX-Trigger: <event>` (for HTMX events)

## Error Responses

All API errors return structured JSON responses:

```json
{
  "type": "validation|authentication|authorization|not_found|conflict|rate_limit|internal|external|timeout|csrf|sanitization",
  "error": "HTTP Status Text",
  "message": "Detailed error description",
  "details": {
    "field": "validation error details"
  },
  "code": 400,
  "path": "/api/endpoint",
  "method": "POST",
  "request_id": "uuid-string",
  "timestamp": "server-time"
}
```

## Health Check

### GET /health

Returns server health status.

**Response:**
```json
{
  "status": "ok",
  "service": "go-web-server",
  "version": "1.0.0",
  "uptime": "1h23m45s",
  "timestamp": "2024-01-01T12:00:00Z",
  "checks": {
    "database": "ok",
    "memory": "ok"
  }
}
```

## Static Assets

### GET /static/*

Serves static files (CSS, JS, images).

**Examples:**
- `/static/css/pico.min.css`
- `/static/js/htmx.min.js`

## Web Pages

### GET /

Home page with feature overview and interactive demos.

### GET /demo

Interactive demo content (HTMX partial).

### GET /users

User management page with full layout.

## User Management API

### List Users

**GET /users/list**

Returns HTML fragment with user list.

**Response:** HTML table with user data

### Get User Count

**GET /api/users/count**

Returns the number of active users.

**Response:** Plain text number (e.g., `42`)

### User Form

**GET /users/form**

Returns HTML form for creating a new user.

**Response:** HTML form fragment

### Edit User Form

**GET /users/:id/edit**

Returns HTML form for editing an existing user.

**Parameters:**
- `id` (path) - User ID (integer)

**Response:** HTML form fragment populated with user data

**Errors:**
- `400` - Invalid user ID format
- `404` - User not found

### Create User

**POST /users**

Creates a new user.

**Headers:**
- `X-CSRF-Token: <token>` (required)

**Form Data:**
- `name` (required) - User's full name
- `email` (required) - User's email address
- `bio` (optional) - User biography
- `avatar_url` (optional) - URL to user's avatar image

**Response:** HTML user list fragment

**Errors:**
- `400` - Validation failed (missing required fields)
- `403` - Invalid CSRF token
- `500` - Internal server error

**Example:**
```bash
curl -X POST http://localhost:8080/users \
  -H "X-CSRF-Token: abc123" \
  -d "name=John Doe" \
  -d "email=john@example.com" \
  -d "bio=Software developer"
```

### Update User

**PUT /users/:id**

Updates an existing user.

**Parameters:**
- `id` (path) - User ID (integer)

**Headers:**
- `X-CSRF-Token: <token>` (required)

**Form Data:**
- `name` (required) - User's full name
- `bio` (optional) - User biography
- `avatar_url` (optional) - URL to user's avatar image

**Response:** HTML user list fragment

**Errors:**
- `400` - Invalid user ID format or validation failed
- `403` - Invalid CSRF token
- `404` - User not found
- `500` - Internal server error

### Deactivate User

**PATCH /users/:id/deactivate**

Deactivates a user (soft delete).

**Parameters:**
- `id` (path) - User ID (integer)

**Headers:**
- `X-CSRF-Token: <token>` (required)

**Response:** HTML user row fragment (updated)

**Errors:**
- `400` - Invalid user ID format
- `403` - Invalid CSRF token
- `404` - User not found
- `500` - Internal server error

### Delete User

**DELETE /users/:id**

Permanently deletes a user.

**Parameters:**
- `id` (path) - User ID (integer)

**Headers:**
- `X-CSRF-Token: <token>` (required)

**Response:** 204 No Content (empty response)

**Errors:**
- `400` - Invalid user ID format
- `403` - Invalid CSRF token
- `404` - User not found
- `500` - Internal server error

## HTMX Integration

This API is designed to work seamlessly with HTMX. Key features:

### Custom Events

The API triggers custom HTMX events:
- `userCreated` - After creating a user
- `userUpdated` - After updating a user  
- `userDeactivated` - After deactivating a user
- `userDeleted` - After deleting a user

### CSRF Protection

CSRF tokens are automatically managed:
- GET requests receive tokens via cookies
- Forms must include `csrf_token` field
- HTMX requests use `X-CSRF-Token` header

### Request/Response Flow

1. Initial page load sets CSRF cookie
2. HTMX requests include CSRF token in headers
3. Server validates token and processes request
4. Response includes new CSRF token for next request
5. HTMX swaps content and triggers events

## Rate Limiting

- **Limit:** 20 requests per IP per minute
- **Response:** 429 Too Many Requests
- **Headers:** Standard rate limit headers included

## Input Sanitization

All form inputs are automatically sanitized:
- HTML tags escaped
- XSS vectors removed
- SQL injection patterns blocked
- Custom sanitization rules applied

## Request Tracing

Every request receives a unique ID for tracing:
- Header: `X-Request-ID`
- Included in error responses
- Used in server logs