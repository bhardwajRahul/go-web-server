# API Reference

HTTP endpoints and HTMX integration for the Go Web Server.

## Base URL

```
http://localhost:8080
```

## Authentication

No authentication required - this is a demo application.

## Headers

**Required for state-changing operations:**

- `X-CSRF-Token: <token>` (POST/PUT/PATCH/DELETE)

**Response headers:**

- `X-Request-ID: <uuid>` (request tracing)
- `HX-Trigger: <event>` (HTMX events)

## Error Format

```json
{
  "type": "validation|not_found|internal|csrf",
  "error": "Bad Request",
  "message": "Detailed error description",
  "code": 400,
  "request_id": "uuid",
  "timestamp": "server-time"
}
```

## Core Endpoints

### Health Check

**GET /health**

```json
{
  "status": "ok",
  "service": "go-web-server",
  "version": "1.0.0",
  "uptime": "1h23m45s",
  "checks": { "database": "ok", "memory": "ok" }
}
```

### Static Files

**GET /static/\***

- `/static/css/pico.min.css` - Pico.css framework
- `/static/js/htmx.min.js` - HTMX library

## Web Pages

### Home

**GET /** - Home page with feature demos
**GET /demo** - Interactive HTMX demo content

### Users

**GET /users** - User management page

## User API

### List Users (HTMX Fragment)

**GET /users/list**

Returns HTML table with active users.

### User Count

**GET /api/users/count**

Returns plain text count: `42`

### User Forms

**GET /users/form** - New user form
**GET /users/:id/edit** - Edit user form (populated)

### Create User

**POST /users**

**Form Data:**

- `name` (required) - Full name
- `email` (required) - Email address
- `bio` (optional) - Biography
- `avatar_url` (optional) - Avatar URL

**Headers:** `X-CSRF-Token` required

**Response:** Updated user list HTML

### Update User

**PUT /users/:id**

**Form Data:**

- `name` (required) - Full name
- `bio` (optional) - Biography
- `avatar_url` (optional) - Avatar URL

**Headers:** `X-CSRF-Token` required

### Deactivate User

**PATCH /users/:id/deactivate**

Soft delete - sets `is_active = false`

**Response:** Updated user row HTML

### Delete User

**DELETE /users/:id**

Permanent deletion.

**Response:** `204 No Content`

## HTMX Integration

### Custom Events

- `userCreated` - After creating user
- `userUpdated` - After updating user
- `userDeactivated` - After deactivating user
- `userDeleted` - After deleting user

### CSRF Protection

1. GET requests set CSRF cookie
2. Forms include `csrf_token` field
3. HTMX uses `X-CSRF-Token` header
4. Server validates and rotates tokens

### Example HTMX Usage

```html
<!-- Form submission -->
<form hx-post="/users" hx-target="#user-list">
  <input type="hidden" name="csrf_token" value="{{token}}" />
  <input name="name" required />
  <button type="submit">Create</button>
</form>

<!-- Delete button -->
<button
  hx-delete="/users/123"
  hx-confirm="Delete user?"
  hx-target="#user-123"
  hx-swap="outerHTML"
>
  Delete
</button>
```

## Security Features

- **Rate Limiting**: 20 requests/minute per IP
- **Input Sanitization**: XSS and SQL injection protection
- **CSRF Protection**: All state-changing operations
- **Request Tracing**: Unique ID per request
- **Error Sanitization**: No sensitive data in responses
