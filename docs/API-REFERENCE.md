# 🔌 API Reference - NAT Management System

## Table of Contents

- [Authentication](#authentication)
- [Error Responses](#error-responses)
- [Rate Limiting](#rate-limiting)
- [API Endpoints](#api-endpoints)
  - [Auth Endpoints](#auth-endpoints)
  - [Router Endpoints](#router-endpoints)
  - [NAT Endpoints](#nat-endpoints)
  - [PPPoE Endpoints](#pppoe-endpoints)
  - [User Endpoints](#user-endpoints)
  - [Activity Log Endpoints](#activity-log-endpoints)

---

## Authentication

All API endpoints (except `/api/auth/login`) require authentication via JWT token.

### Getting Authentication Token

```http
POST /api/auth/login
Content-Type: application/json

{
  "username": "admin",
  "password": "admin123"
}
```

**Response:**
```json
{
  "success": true,
  "message": "Login successful",
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_at": "2025-10-17T10:30:00Z",
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "username": "admin",
    "full_name": "System Administrator",
    "role": "Administrator",
    "email": "admin@example.com"
  }
}
```

### Using Authentication Token

Include JWT token in Authorization header for all authenticated requests:

```http
GET /api/routers
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

### Token Refresh

When access token expires (24h), use refresh token to get new access token:

```http
POST /api/auth/refresh
Content-Type: application/json

{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Response:**
```json
{
  "success": true,
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_at": "2025-10-17T10:30:00Z"
}
```

---

## Error Responses

### Standard Error Format

```json
{
  "error": "Error message here",
  "details": "Optional detailed error message"
}
```

### HTTP Status Codes

| Code | Meaning | Description |
|------|---------|-------------|
| 200 | OK | Request successful |
| 201 | Created | Resource created successfully |
| 400 | Bad Request | Invalid request parameters |
| 401 | Unauthorized | Authentication required or failed |
| 403 | Forbidden | Insufficient permissions |
| 404 | Not Found | Resource not found |
| 409 | Conflict | Resource already exists |
| 422 | Unprocessable Entity | Validation failed |
| 429 | Too Many Requests | Rate limit exceeded |
| 500 | Internal Server Error | Server error occurred |

### Error Examples

**Unauthorized:**
```json
{
  "error": "Unauthorized"
}
```

**Validation Error:**
```json
{
  "error": "Validation failed",
  "details": "Port must be between 1 and 65535"
}
```

**Rate Limit:**
```json
{
  "error": "Rate limit exceeded. Please try again later."
}
```

---

## Rate Limiting

- **Default Limit:** 100 requests per 60 seconds per user
- **Applies to:** All authenticated API endpoints
- **Headers Returned:**
  - `X-RateLimit-Limit`: Maximum requests allowed
  - `X-RateLimit-Remaining`: Requests remaining
  - `X-RateLimit-Reset`: Time when limit resets (Unix timestamp)

**Example:**
```http
HTTP/1.1 200 OK
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1697530800
```

When limit exceeded:
```http
HTTP/1.1 429 Too Many Requests
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 0
X-RateLimit-Reset: 1697530800

{
  "error": "Rate limit exceeded. Please try again later."
}
```

---

## API Endpoints

## Auth Endpoints

### POST /api/auth/login

Authenticate user and get JWT tokens.

**Request:**
```http
POST /api/auth/login
Content-Type: application/json

{
  "username": "admin",
  "password": "admin123"
}
```

**Response (200 OK):**
```json
{
  "success": true,
  "message": "Login successful",
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_at": "2025-10-17T10:30:00Z",
  "user": {
    "id": "uuid",
    "username": "admin",
    "full_name": "System Administrator",
    "role": "Administrator",
    "email": "admin@example.com"
  }
}
```

**Error Responses:**
- `400`: Missing username or password
- `401`: Invalid credentials
- `403`: User account is inactive

---

### POST /api/auth/logout

Logout user and invalidate session.

**Request:**
```http
POST /api/auth/logout
Authorization: Bearer <token>
```

**Response (200 OK):**
```json
{
  "success": true,
  "message": "Logout successful"
}
```

---

### POST /api/auth/refresh

Refresh access token using refresh token.

**Request:**
```http
POST /api/auth/refresh
Content-Type: application/json

{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Response (200 OK):**
```json
{
  "success": true,
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_at": "2025-10-17T10:30:00Z"
}
```

**Error Responses:**
- `400`: Missing refresh token
- `401`: Invalid or expired refresh token

---

### GET /api/auth/check

Check if current session/token is valid.

**Request:**
```http
GET /api/auth/check
Authorization: Bearer <token>
```

**Response (200 OK):**
```json
{
  "authenticated": true,
  "user": {
    "id": "uuid",
    "username": "admin",
    "role": "Administrator"
  }
}
```

---

### GET /api/auth/me

Get current authenticated user info.

**Request:**
```http
GET /api/auth/me
Authorization: Bearer <token>
```

**Response (200 OK):**
```json
{
  "id": "uuid",
  "username": "admin",
  "full_name": "System Administrator",
  "email": "admin@example.com",
  "role": "Administrator",
  "active": true,
  "created_at": "2025-01-01T00:00:00Z",
  "routers": ["JAKARTA-01", "BANDUNG-01"]
}
```

---

## Router Endpoints

### GET /api/routers

Get list of all routers (filtered by user permissions).

**Request:**
```http
GET /api/routers
Authorization: Bearer <token>
```

**Response (200 OK):**
```json
{
  "routers": [
    {
      "id": "uuid",
      "name": "JAKARTA-01",
      "host": "192.168.1.1",
      "port": 8728,
      "username": "admin",
      "tunnel_endpoint": "172.22.28.5:80",
      "public_ont_url": "http://tunnel3.ebilling.id:19701",
      "description": "Router Cabang Jakarta Pusat",
      "enabled": true,
      "created_at": "2025-01-01T00:00:00Z",
      "updated_at": "2025-01-01T00:00:00Z"
    }
  ]
}
```

**Notes:**
- Password is never returned in API responses
- Administrators see all routers
- Branch users only see assigned routers

---

### POST /api/routers

Create new router (Administrator only).

**Request:**
```http
POST /api/routers
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "JAKARTA-01",
  "host": "192.168.1.1",
  "port": 8728,
  "username": "admin",
  "password": "password123",
  "tunnel_endpoint": "172.22.28.5:80",
  "public_ont_url": "http://tunnel3.ebilling.id:19701",
  "description": "Router Cabang Jakarta Pusat",
  "enabled": true
}
```

**Response (201 Created):**
```json
{
  "success": true,
  "message": "Router created successfully",
  "router": {
    "id": "uuid",
    "name": "JAKARTA-01",
    "host": "192.168.1.1",
    "port": 8728,
    "username": "admin",
    "tunnel_endpoint": "172.22.28.5:80",
    "public_ont_url": "http://tunnel3.ebilling.id:19701",
    "description": "Router Cabang Jakarta Pusat",
    "enabled": true,
    "created_at": "2025-10-16T10:30:00Z",
    "updated_at": "2025-10-16T10:30:00Z"
  }
}
```

**Validation Rules:**
- `name`: Required, unique, max 100 chars
- `host`: Required, valid IP or hostname
- `port`: Required, 1-65535
- `username`: Required
- `password`: Required
- `tunnel_endpoint`: Optional, format IP:PORT
- `public_ont_url`: Optional, valid URL

**Error Responses:**
- `400`: Validation failed
- `403`: Insufficient permissions (not Administrator)
- `409`: Router name already exists

---

### GET /api/routers/:id

Get router details by ID.

**Request:**
```http
GET /api/routers/550e8400-e29b-41d4-a716-446655440000
Authorization: Bearer <token>
```

**Response (200 OK):**
```json
{
  "id": "uuid",
  "name": "JAKARTA-01",
  "host": "192.168.1.1",
  "port": 8728,
  "username": "admin",
  "tunnel_endpoint": "172.22.28.5:80",
  "public_ont_url": "http://tunnel3.ebilling.id:19701",
  "description": "Router Cabang Jakarta Pusat",
  "enabled": true,
  "created_at": "2025-01-01T00:00:00Z",
  "updated_at": "2025-01-01T00:00:00Z"
}
```

**Error Responses:**
- `403`: User doesn't have access to this router
- `404`: Router not found

---

### PUT /api/routers/:id

Update router configuration (Administrator only).

**Request:**
```http
PUT /api/routers/550e8400-e29b-41d4-a716-446655440000
Authorization: Bearer <token>
Content-Type: application/json

{
  "host": "192.168.1.2",
  "port": 8728,
  "description": "Updated description",
  "enabled": true
}
```

**Response (200 OK):**
```json
{
  "success": true,
  "message": "Router updated successfully",
  "router": {
    "id": "uuid",
    "name": "JAKARTA-01",
    "host": "192.168.1.2",
    ...
  }
}
```

**Error Responses:**
- `400`: Validation failed
- `403`: Insufficient permissions
- `404`: Router not found

---

### DELETE /api/routers/:id

Delete router (Administrator only).

**Request:**
```http
DELETE /api/routers/550e8400-e29b-41d4-a716-446655440000
Authorization: Bearer <token>
```

**Response (200 OK):**
```json
{
  "success": true,
  "message": "Router deleted successfully"
}
```

**Error Responses:**
- `403`: Insufficient permissions
- `404`: Router not found

---

### POST /api/routers/:id/test

Test router connection.

**Request:**
```http
POST /api/routers/550e8400-e29b-41d4-a716-446655440000/test
Authorization: Bearer <token>
```

**Response (200 OK):**
```json
{
  "status": "connected",
  "message": "Connection successful",
  "router_identity": "RB750-JAKARTA",
  "system_info": {
    "version": "6.49.10",
    "board_name": "RB750Gr3",
    "platform": "MikroTik",
    "architecture": "arm",
    "cpu": "ARMv7",
    "cpu_count": 4,
    "uptime": "2d3h15m"
  },
  "response_time_ms": 245,
  "timestamp": "2025-10-16T10:30:00Z"
}
```

**Error Response (when failed):**
```json
{
  "status": "disconnected",
  "message": "Connection failed: dial tcp 192.168.1.1:8728: i/o timeout",
  "timestamp": "2025-10-16T10:30:00Z"
}
```

---

### GET /api/routers/stats

Get router statistics (Administrator only).

**Request:**
```http
GET /api/routers/stats
Authorization: Bearer <token>
```

**Response (200 OK):**
```json
{
  "total_routers": 10,
  "enabled_routers": 8,
  "disabled_routers": 2,
  "online_routers": 7,
  "offline_routers": 3
}
```

---

## NAT Endpoints

### GET /api/nat/configs

Get NAT configurations from router.

**Request:**
```http
GET /api/nat/configs?router=JAKARTA-01
Authorization: Bearer <token>
```

**Query Parameters:**
- `router` (required): Router name

**Response (200 OK):**
```json
{
  "router": "JAKARTA-01",
  "configs": [
    {
      "id": "*1",
      "chain": "dstnat",
      "protocol": "tcp",
      "dst_port": "19701",
      "to_addresses": "172.22.28.5",
      "to_ports": "80",
      "comment": "REMOTE ONT PELANGGAN user123",
      "disabled": false
    }
  ]
}
```

---

### GET /api/nat/clients

Get online NAT clients (active connections).

**Request:**
```http
GET /api/nat/clients?router=JAKARTA-01
Authorization: Bearer <token>
```

**Query Parameters:**
- `router` (required): Router name

**Response (200 OK):**
```json
{
  "router": "JAKARTA-01",
  "clients": [
    {
      "protocol": "tcp",
      "src_address": "10.10.10.100:54321",
      "dst_address": "172.22.28.5:80",
      "reply_src_address": "172.22.28.5:80",
      "reply_dst_address": "10.10.10.100:54321",
      "connection_state": "established",
      "timeout": "23h59m59s"
    }
  ]
}
```

---

### POST /api/nat/update

Update NAT rule destination.

**Request:**
```http
POST /api/nat/update
Authorization: Bearer <token>
Content-Type: application/json

{
  "router": "JAKARTA-01",
  "pppoe_username": "user123",
  "destination_ip": "10.10.10.100",
  "destination_port": 80
}
```

**Response (200 OK):**
```json
{
  "success": true,
  "message": "NAT rule updated successfully",
  "old_destination": "172.22.28.5:80",
  "new_destination": "10.10.10.100:80",
  "nat_rule_id": "*1",
  "router": "JAKARTA-01"
}
```

**Error Responses:**
- `400`: Missing required fields
- `403`: No access to router
- `404`: NAT rule not found for username
- `500`: Update failed

---

### GET /api/nat/status

Get NAT service status.

**Request:**
```http
GET /api/nat/status?router=JAKARTA-01
Authorization: Bearer <token>
```

**Response (200 OK):**
```json
{
  "router": "JAKARTA-01",
  "status": "online",
  "nat_rules_count": 15,
  "active_connections": 42
}
```

---

## PPPoE Endpoints

### POST /api/pppoe/check

Check PPPoE user status (single router).

**Request:**
```http
POST /api/pppoe/check
Authorization: Bearer <token>
Content-Type: application/json

{
  "router": "JAKARTA-01",
  "username": "user123"
}
```

**Response (200 OK):**
```json
{
  "found": true,
  "router": "JAKARTA-01",
  "session": {
    "name": "user123",
    "service": "pppoe-out1",
    "address": "10.10.10.100",
    "uptime": "2d3h15m",
    "encoding": "MPPE128",
    "session_id": "0x80700001"
  }
}
```

**Response (404 Not Found):**
```json
{
  "found": false,
  "message": "PPPoE user not found",
  "router": "JAKARTA-01"
}
```

---

### POST /api/pppoe/fuzzy-search

Fuzzy search PPPoE users across multiple routers.

**Request:**
```http
POST /api/pppoe/fuzzy-search
Authorization: Bearer <token>
Content-Type: application/json

{
  "username": "user",
  "routers": ["JAKARTA-01", "BANDUNG-01", "SURABAYA-01"]
}
```

**Response (200 OK):**
```json
{
  "query": "user",
  "total_found": 3,
  "results": [
    {
      "router": "JAKARTA-01",
      "sessions": [
        {
          "name": "user123",
          "address": "10.10.10.100",
          "uptime": "2d3h15m"
        },
        {
          "name": "user456",
          "address": "10.10.10.101",
          "uptime": "1d5h30m"
        }
      ]
    },
    {
      "router": "BANDUNG-01",
      "sessions": [
        {
          "name": "user789",
          "address": "10.20.20.100",
          "uptime": "3h45m"
        }
      ]
    },
    {
      "router": "SURABAYA-01",
      "sessions": []
    }
  ]
}
```

---

### GET /api/pppoe/routers

Get available routers for PPPoE checking (based on user permissions).

**Request:**
```http
GET /api/pppoe/routers
Authorization: Bearer <token>
```

**Response (200 OK):**
```json
{
  "routers": [
    {
      "name": "JAKARTA-01",
      "description": "Router Cabang Jakarta Pusat"
    },
    {
      "name": "BANDUNG-01",
      "description": "Router Cabang Bandung"
    }
  ]
}
```

---

## User Endpoints

### GET /api/users

Get list of all users (Administrator only).

**Request:**
```http
GET /api/users
Authorization: Bearer <token>
```

**Response (200 OK):**
```json
{
  "users": [
    {
      "id": "uuid",
      "username": "admin",
      "full_name": "System Administrator",
      "email": "admin@example.com",
      "role": "Administrator",
      "active": true,
      "created_at": "2025-01-01T00:00:00Z",
      "routers": ["*"]
    },
    {
      "id": "uuid",
      "username": "head1",
      "full_name": "Head Branch 1",
      "email": "head1@example.com",
      "role": "Head Branch 1",
      "active": true,
      "created_at": "2025-01-01T00:00:00Z",
      "routers": ["JAKARTA-01", "BANDUNG-01"]
    }
  ]
}
```

---

### POST /api/users

Create new user (Administrator only).

**Request:**
```http
POST /api/users
Authorization: Bearer <token>
Content-Type: application/json

{
  "username": "newuser",
  "password": "password123",
  "full_name": "New User",
  "email": "newuser@example.com",
  "role": "Head Branch 1",
  "routers": ["JAKARTA-01", "BANDUNG-01"],
  "active": true
}
```

**Response (201 Created):**
```json
{
  "success": true,
  "message": "User created successfully",
  "user": {
    "id": "uuid",
    "username": "newuser",
    "full_name": "New User",
    "email": "newuser@example.com",
    "role": "Head Branch 1",
    "active": true,
    "created_at": "2025-10-16T10:30:00Z"
  }
}
```

**Validation Rules:**
- `username`: Required, unique, 3-50 chars, alphanumeric
- `password`: Required, min 6 chars
- `full_name`: Required
- `email`: Optional, valid email format
- `role`: Required, one of: Administrator, Head Branch 1, Head Branch 2, Head Branch 3
- `routers`: Array of router names (not required for Administrator)

**Error Responses:**
- `400`: Validation failed
- `403`: Insufficient permissions
- `409`: Username already exists

---

### GET /api/users/:id

Get user details (Administrator or self).

**Request:**
```http
GET /api/users/550e8400-e29b-41d4-a716-446655440000
Authorization: Bearer <token>
```

**Response (200 OK):**
```json
{
  "id": "uuid",
  "username": "head1",
  "full_name": "Head Branch 1",
  "email": "head1@example.com",
  "role": "Head Branch 1",
  "active": true,
  "created_at": "2025-01-01T00:00:00Z",
  "updated_at": "2025-01-01T00:00:00Z",
  "routers": ["JAKARTA-01", "BANDUNG-01"]
}
```

---

### PUT /api/users/:id

Update user (Administrator or self for limited fields).

**Request:**
```http
PUT /api/users/550e8400-e29b-41d4-a716-446655440000
Authorization: Bearer <token>
Content-Type: application/json

{
  "full_name": "Updated Name",
  "email": "updated@example.com",
  "routers": ["JAKARTA-01", "BANDUNG-01", "SURABAYA-01"]
}
```

**Response (200 OK):**
```json
{
  "success": true,
  "message": "User updated successfully",
  "user": {
    "id": "uuid",
    "username": "head1",
    "full_name": "Updated Name",
    "email": "updated@example.com",
    ...
  }
}
```

**Notes:**
- Non-admins can only update their own profile (full_name, email)
- Admins can update all fields except username

---

### DELETE /api/users/:id

Delete user (Administrator only).

**Request:**
```http
DELETE /api/users/550e8400-e29b-41d4-a716-446655440000
Authorization: Bearer <token>
```

**Response (200 OK):**
```json
{
  "success": true,
  "message": "User deleted successfully"
}
```

**Error Responses:**
- `403`: Insufficient permissions or trying to delete self
- `404`: User not found

---

### PATCH /api/users/:id/password

Change user password (Administrator or self).

**Request:**
```http
PATCH /api/users/550e8400-e29b-41d4-a716-446655440000/password
Authorization: Bearer <token>
Content-Type: application/json

{
  "current_password": "oldpassword",
  "new_password": "newpassword123"
}
```

**Response (200 OK):**
```json
{
  "success": true,
  "message": "Password changed successfully"
}
```

**Notes:**
- `current_password` required when changing own password
- Admins can change any user's password without current password

---

### GET /api/users/:id/stats

Get user activity statistics.

**Request:**
```http
GET /api/users/550e8400-e29b-41d4-a716-446655440000/stats
Authorization: Bearer <token>
```

**Response (200 OK):**
```json
{
  "user_id": "uuid",
  "username": "head1",
  "total_actions": 150,
  "actions_by_type": {
    "LOGIN": 50,
    "NAT_UPDATE": 75,
    "PPPOE_CHECK": 25
  },
  "last_login": "2025-10-16T08:00:00Z",
  "last_action": "2025-10-16T10:15:00Z"
}
```

---

## Activity Log Endpoints

### GET /api/logs

Get activity logs (Administrator only, others see own logs).

**Request:**
```http
GET /api/logs?page=1&limit=50&action=NAT_UPDATE&user=head1&router=JAKARTA-01&start_date=2025-10-01&end_date=2025-10-16
Authorization: Bearer <token>
```

**Query Parameters:**
- `page` (optional): Page number, default 1
- `limit` (optional): Items per page, default 50, max 100
- `action` (optional): Filter by action type
- `user` (optional): Filter by username
- `router` (optional): Filter by router name
- `start_date` (optional): Start date (YYYY-MM-DD)
- `end_date` (optional): End date (YYYY-MM-DD)
- `status` (optional): SUCCESS or FAILED

**Response (200 OK):**
```json
{
  "logs": [
    {
      "id": "uuid",
      "user_id": "uuid",
      "username": "head1",
      "action": "NAT_UPDATE",
      "resource": "NAT",
      "router_name": "JAKARTA-01",
      "details": "{\"username\":\"user123\",\"old_ip\":\"172.22.28.5\",\"new_ip\":\"10.10.10.100\"}",
      "ip_address": "192.168.1.100",
      "user_agent": "Mozilla/5.0...",
      "status": "SUCCESS",
      "error_message": null,
      "created_at": "2025-10-16T10:30:00Z"
    }
  ],
  "pagination": {
    "current_page": 1,
    "per_page": 50,
    "total_pages": 3,
    "total_items": 150
  }
}
```

---

### GET /api/logs/:id

Get single log entry details.

**Request:**
```http
GET /api/logs/550e8400-e29b-41d4-a716-446655440000
Authorization: Bearer <token>
```

**Response (200 OK):**
```json
{
  "id": "uuid",
  "user_id": "uuid",
  "username": "head1",
  "action": "NAT_UPDATE",
  "resource": "NAT",
  "router_name": "JAKARTA-01",
  "details": {
    "username": "user123",
    "old_destination": "172.22.28.5:80",
    "new_destination": "10.10.10.100:80",
    "nat_rule_id": "*1"
  },
  "ip_address": "192.168.1.100",
  "user_agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64)...",
  "status": "SUCCESS",
  "error_message": null,
  "created_at": "2025-10-16T10:30:00Z"
}
```

---

### GET /api/logs/stats

Get activity log statistics (Administrator only).

**Request:**
```http
GET /api/logs/stats
Authorization: Bearer <token>
```

**Response (200 OK):**
```json
{
  "total_logs": 5000,
  "logs_by_action": {
    "LOGIN": 500,
    "NAT_UPDATE": 2500,
    "PPPOE_CHECK": 1500,
    "ROUTER_CREATE": 50,
    "USER_CREATE": 20
  },
  "logs_by_status": {
    "SUCCESS": 4800,
    "FAILED": 200
  },
  "logs_by_user": {
    "admin": 1000,
    "head1": 1500,
    "head2": 1200,
    "head3": 1300
  },
  "recent_failures": [
    {
      "action": "NAT_UPDATE",
      "username": "head1",
      "error": "Connection timeout",
      "created_at": "2025-10-16T10:15:00Z"
    }
  ]
}
```

---

### POST /api/logs/cleanup

Delete old logs (Administrator only).

**Request:**
```http
POST /api/logs/cleanup
Authorization: Bearer <token>
Content-Type: application/json

{
  "older_than_days": 90
}
```

**Response (200 OK):**
```json
{
  "success": true,
  "message": "Old logs deleted successfully",
  "deleted_count": 1500
}
```

---

## Action Types Reference

### User Actions
- `LOGIN` - User logged in
- `LOGOUT` - User logged out
- `PASSWORD_CHANGE` - Password changed
- `TOKEN_REFRESH` - JWT token refreshed

### Router Actions
- `ROUTER_CREATE` - Router created
- `ROUTER_UPDATE` - Router updated
- `ROUTER_DELETE` - Router deleted
- `ROUTER_TEST` - Router connection tested

### NAT Actions
- `NAT_UPDATE` - NAT rule updated
- `NAT_VIEW` - NAT configs viewed
- `NAT_CLIENT_VIEW` - NAT clients viewed

### PPPoE Actions
- `PPPOE_CHECK` - PPPoE status checked
- `PPPOE_SEARCH` - Fuzzy search performed

### User Management Actions
- `USER_CREATE` - User created
- `USER_UPDATE` - User updated
- `USER_DELETE` - User deleted
- `USER_ACTIVATE` - User activated
- `USER_DEACTIVATE` - User deactivated

---

## Rate Limit Headers

All authenticated endpoints include rate limit headers:

```http
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1697530800
```

---

**Version:** 4.1
**Last Updated:** 2025-10-16
**Base URL:** `http://localhost:8080`
