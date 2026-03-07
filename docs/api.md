# API

Base URL: `http://localhost:8080`

## Public Routes

| Method | Path | Notes |
| --- | --- | --- |
| `GET` | `/` | Home page |
| `GET` | `/demo` | HTMX demo fragment or JSON response |
| `GET` | `/health` | Health response with database check |
| `GET` | `/auth/login` | Login page |
| `GET` | `/auth/register` | Registration page |
| `POST` | `/auth/login` | Creates a session on success |
| `POST` | `/auth/register` | Creates a user and session on success |
| `POST` | `/auth/logout` | Destroys the current session if present |
| `GET` | `/static/*` | Embedded static assets |

## Protected Routes

These routes now require an authenticated session.

| Method | Path | Notes |
| --- | --- | --- |
| `GET` | `/profile` | Profile page |
| `GET` | `/users` | User management page |
| `GET` | `/users/list` | User list fragment |
| `GET` | `/users/form` | New-user form fragment |
| `GET` | `/users/:id/edit` | Edit-user form fragment |
| `POST` | `/users` | Create user |
| `PUT` | `/users/:id` | Update user |
| `PATCH` | `/users/:id/deactivate` | Soft deactivate user |
| `DELETE` | `/users/:id` | Hard delete user |
| `GET` | `/api/users/count` | Active user count |

## Auth Behavior

- Browser requests without a session are redirected to `/auth/login`.
- HTMX or JSON-style requests without a session receive `401 Unauthorized`.
- Registered users get Argon2id password hashes.
- Seed/demo users with no password hash can still log in with any password. That is convenient for local demos and not something to keep if you turn this into a real app.

## CSRF

- `POST`, `PUT`, `PATCH`, and `DELETE` require a CSRF token.
- Tokens are issued in the `_csrf` cookie and exposed to templates/HTMX through response headers and hidden form fields.

## Health Endpoint

`GET /health` returns JSON like:

```json
{
  "status": "ok",
  "timestamp": "2026-03-07T15:04:05Z",
  "service": "go-web-server",
  "version": "1.0.0",
  "uptime": "12m3s",
  "checks": {
    "database": "ok",
    "database_connections": "ok"
  }
}
```

Status codes:

- `200 OK`: healthy
- `206 Partial Content`: warning or degraded
- `503 Service Unavailable`: unhealthy
