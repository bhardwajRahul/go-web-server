# API Reference

HTTP endpoints and HTMX integration for the Modern Go Stack web server.

## Base URL

```
http://localhost:8080
```

## System Endpoints

### Health Check

**GET /health**

Returns comprehensive system health with database connectivity check.

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

**Response Status:**

- `200 OK` - All systems healthy
- `503 Service Unavailable` - System degraded

### Metrics (Optional)

**GET /metrics**

Prometheus metrics endpoint (enabled with `FEATURES_ENABLE_METRICS=true`).

**Metrics Included:**

- HTTP request metrics (duration, status, count)
- Database connection and query metrics  
- HTMX interaction metrics
- CSRF protection metrics
- Business metrics (user operations)

### Static Assets

**GET /static/***

Embedded static files served directly from binary:

- `/static/css/pico.min.css` - Pico.css v2 framework
- `/static/css/theme.css` - Custom theme variables
- `/static/css/animations.css` - Smooth transitions
- `/static/css/components.css` - Custom components
- `/static/css/layout.css` - Layout utilities
- `/static/js/htmx.min.js` - HTMX 2.x library
- `/static/favicon.ico` - Application favicon

## Page Endpoints

### Home Page

**GET /**

Main landing page showcasing the Modern Go Stack features.

**Features:**

- Interactive HTMX demo area
- Technology stack overview
- Live system information
- Theme switching (dark/light/auto)

**HTMX Integration:**

- Smooth page transitions
- Dynamic content loading
- Real-time demos

### Demo Content

**GET /demo**

Interactive demo showcasing HTMX functionality.

**Response:** HTML fragment for HTMX swap

```html
<article class="fade-in">
  <header>
    <h4>Live Demo Results</h4>
  </header>
  <p><strong>HTMX + Go + Templ working together!</strong></p>
  <!-- Additional demo content -->
</article>
```

## Authentication Endpoints

### Login Page

**GET /auth/login**

Login form with CSRF protection.

**Features:**

- CSRF token integration
- HTMX form submission
- Automatic redirect if authenticated
- Theme-aware styling

**HTMX Support:**

- Partial page updates
- Smooth transitions
- Error handling

### Register Page

**GET /auth/register**

User registration form with validation.

**Features:**

- Comprehensive input validation
- Password strength requirements
- Email format validation
- Bio and avatar URL support

### User Authentication

**POST /auth/login**

Authenticate user and create JWT session.

**Request Body:**

```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

**Response:**

- `200 OK` - Login successful, JWT cookie set
- `400 Bad Request` - Validation errors
- `401 Unauthorized` - Invalid credentials

**CSRF Protection:** Required for all POST requests

**Demo Mode Note:** Current implementation bypasses password validation for existing sample users.

### User Registration

**POST /auth/register**

Create new user account with automatic login.

**Request Body:**

```json
{
  "email": "user@example.com",
  "name": "John Doe",
  "password": "StrongPass123",
  "confirm_password": "StrongPass123",
  "bio": "Optional bio text",
  "avatar_url": "https://example.com/avatar.jpg"
}
```

**Validation Rules:**

- Email: Valid email format, unique
- Name: 2-100 characters
- Password: 8+ chars with uppercase, lowercase, numbers
- Bio: Max 500 characters (optional)
- Avatar URL: Valid URL format (optional)

**Response:**

- `200 OK` - Registration successful, user logged in
- `400 Bad Request` - Validation errors
- `409 Conflict` - Email already exists

### Logout

**POST /auth/logout**

Clear JWT authentication cookie.

**Response:**

- `200 OK` - Logout successful
- Redirects to `/auth/login`

**CSRF Protection:** Required

## User Management Endpoints

### Users Page

**GET /users**

User management interface with CRUD operations.

**Features:**

- User list with real-time updates
- Create/edit/delete functionality
- Search and filtering
- Status management (active/inactive)

### User List (HTMX)

**GET /users/list**

Returns user list as HTML fragment for HTMX updates.

**Response:** Table rows with user data

```html
<tr>
  <td>John Doe</td>
  <td>john@example.com</td>
  <td><span class="active">Active</span></td>
  <td>
    <button hx-get="/users/1/edit">Edit</button>
    <button hx-delete="/users/1">Delete</button>
  </td>
</tr>
```

### User Form (HTMX)

**GET /users/form**

New user creation form.

**GET /users/:id/edit**

Edit existing user form.

**Response:** Form HTML with pre-populated data (for edit)

### Create User

**POST /users**

Create new user.

**Request Body:**

```json
{
  "email": "user@example.com",
  "name": "John Doe", 
  "bio": "Optional bio",
  "avatar_url": "https://example.com/avatar.jpg"
}
```

**Response:**

- `201 Created` - User created successfully
- `400 Bad Request` - Validation errors
- `409 Conflict` - Email already exists

**CSRF Protection:** Required

### Update User

**PUT /users/:id**

Update existing user.

**Path Parameters:**

- `id` - User ID (integer)

**Request Body:** Same as create user

**Response:**

- `200 OK` - User updated successfully
- `400 Bad Request` - Validation errors
- `404 Not Found` - User not found

**CSRF Protection:** Required

### Deactivate User

**PATCH /users/:id/deactivate**

Mark user as inactive (soft delete).

**Response:**

- `200 OK` - User deactivated
- `404 Not Found` - User not found

**CSRF Protection:** Required

### Delete User

**DELETE /users/:id**

Permanently delete user.

**Response:**

- `200 OK` - User deleted successfully
- `404 Not Found` - User not found

**CSRF Protection:** Required

## API Endpoints

### User Count

**GET /api/users/count**

Get total count of active users.

**Response:**

```json
{
  "count": 42
}
```

## Protected Routes

### User Profile

**GET /profile**

User profile page (authentication required).

**Authentication:** JWT cookie required

**Features:**

- Display current user information
- Profile editing capabilities
- Session management

## HTMX Integration

### Request Headers

HTMX automatically includes these headers:

```
HX-Request: true
HX-Target: #element-id
HX-Trigger: button
HX-Current-URL: /current/path
```

### CSRF Token Handling

HTMX requests automatically include CSRF tokens via JavaScript:

```javascript
// Automatically added to all HTMX requests
headers['X-CSRF-Token'] = currentCSRFToken;
```

### Response Headers

Server includes HTMX-specific response headers:

```
HX-Redirect: /new/path      // Client-side redirect
HX-Refresh: true            // Refresh entire page
X-CSRF-Token: new_token     // Updated CSRF token
```

### Error Handling

HTMX error responses include structured error information:

```json
{
  "type": "validation", 
  "error": "Bad Request",
  "message": "Validation failed",
  "details": [
    {"field": "email", "message": "invalid email format"}
  ],
  "code": 400,
  "request_id": "req-123456"
}
```

## Security Features

### CSRF Protection

All state-changing operations (POST, PUT, PATCH, DELETE) require CSRF tokens:

**Form-based:**

```html
<input type="hidden" name="csrf_token" value="{{.CSRFToken}}" />
```

**Header-based (HTMX):**

```html
<button hx-post="/users" hx-headers='{"X-CSRF-Token": "{{.CSRFToken}}"}'>
```

**Automatic (JavaScript):**
HTMX requests automatically include tokens via configured JavaScript.

### Rate Limiting

All endpoints protected by rate limiting:

- **Limit:** 20 requests per minute per IP address
- **Response:** 429 Too Many Requests when exceeded
- **Headers:** Rate limit information in response headers

### Input Validation

All request payloads validated:

- Email format validation
- String length limits  
- Required field validation
- Custom business rules

### Authentication

JWT-based authentication with:

- HTTPOnly secure cookies
- 24-hour token expiration
- Automatic token refresh
- Secure logout

## Example Usage

### HTMX User Creation

```html
<form hx-post="/users" hx-target="#user-list" hx-swap="outerHTML">
  <input type="hidden" name="csrf_token" value="{{.CSRFToken}}" />
  <input type="email" name="email" required />
  <input type="text" name="name" required />
  <textarea name="bio"></textarea>
  <button type="submit">Create User</button>
</form>
```

### JavaScript Fetch with CSRF

```javascript
fetch('/api/users/count', {
  headers: {
    'X-CSRF-Token': document.querySelector('meta[name="csrf-token"]').content
  }
})
.then(response => response.json())
.then(data => console.log('User count:', data.count));
```

### cURL Examples

**Health Check:**

```bash
curl http://localhost:8080/health
```

**User Login:**

```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -H "X-CSRF-Token: your-csrf-token" \
  -d '{"email":"test@example.com","password":"password"}'
```

This API provides a complete web application interface with modern security practices, HTMX integration, and comprehensive error handling.
