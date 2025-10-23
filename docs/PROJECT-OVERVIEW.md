# 📖 NAT Management System - Project Overview

## Table of Contents

- [System Architecture](#system-architecture)
- [Feature Details](#feature-details)
- [Database Schema](#database-schema)
- [User Roles & Permissions](#user-roles--permissions)
- [Project Structure](#project-structure)
- [Component Descriptions](#component-descriptions)
- [Data Flow](#data-flow)
- [Security Model](#security-model)

---

## System Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         Client Browser                           │
│                    (HTML/CSS/JavaScript)                         │
└────────────────────────┬────────────────────────────────────────┘
                         │ HTTP/HTTPS
                         │ (Session + JWT)
┌────────────────────────▼────────────────────────────────────────┐
│                     Gin Web Server                               │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  Authentication Middleware                                │  │
│  │  - Session Validation                                     │  │
│  │  - JWT Verification                                       │  │
│  │  - Rate Limiting                                          │  │
│  │  - CORS Protection                                        │  │
│  └────────────────────┬─────────────────────────────────────┘  │
│                       │                                          │
│  ┌────────────────────▼─────────────────────────────────────┐  │
│  │              API Handlers Layer                          │  │
│  │  - AuthHandler     - RouterHandler                       │  │
│  │  - NATHandler      - UserHandler                         │  │
│  │  - ActivityLogHandler                                    │  │
│  └────────────────────┬─────────────────────────────────────┘  │
│                       │                                          │
│  ┌────────────────────▼─────────────────────────────────────┐  │
│  │            Business Logic Services                       │  │
│  │  - AuthService     - RouterService                       │  │
│  │  - NATService      - UserService                         │  │
│  │  - ActivityLogService                                    │  │
│  └────────────────────┬─────────────────────────────────────┘  │
└───────────────────────┼──────────────────────────────────────┘
                        │
        ┌───────────────┴───────────────┐
        │                               │
┌───────▼────────┐            ┌─────────▼──────────┐
│   PostgreSQL   │            │  MikroTik Routers  │
│   (Neon.tech)  │            │   (RouterOS API)   │
│                │            │                    │
│  - users       │            │  - NAT configs     │
│  - routers     │            │  - PPPoE sessions  │
│  - logs        │            │  - System info     │
└────────────────┘            └────────────────────┘
```

### Architecture Layers

1. **Presentation Layer**
   - HTML templates (Gin templating)
   - Static assets (CSS, JS, images)
   - Mobile-responsive UI

2. **API Layer**
   - RESTful endpoints
   - JWT authentication
   - Request validation
   - Response formatting

3. **Business Logic Layer**
   - Service interfaces
   - Business rules
   - Data transformation
   - Error handling

4. **Data Access Layer**
   - PostgreSQL database
   - RouterOS API client
   - Connection pooling
   - Transaction management

---

## Feature Details

### 1. NAT Configuration Management

**Purpose**: Update NAT rules di MikroTik untuk ONT remote access

**Flow:**
```
User Input (PPPoE Username)
    ↓
Search NAT rule dengan comment "REMOTE ONT PELANGGAN"
    ↓
Get current configuration
    ↓
Update dst-address dengan IP:port baru
    ↓
Verify changes
    ↓
Log activity
```

**MikroTik Commands:**
```routeros
# Find NAT rule
/ip firewall nat print where comment~"REMOTE ONT PELANGGAN"

# Update NAT rule
/ip firewall nat set <id> dst-address=<new-ip>:<port>

# Verify
/ip firewall nat print where comment~"REMOTE ONT PELANGGAN"
```

**UI Features:**
- Router selection dropdown
- PPPoE username search
- Current config display
- New IP/port input
- Update confirmation
- Success/error notification

---

### 2. Online Clients Monitoring

**Purpose**: Monitor active NAT connections real-time

**Flow:**
```
Select Router
    ↓
Get NAT clients from /ip firewall nat
    ↓
Parse connection data
    ↓
Display in table with filters
    ↓
Auto-refresh option
```

**Data Collected:**
- Source IP & Port
- Destination IP & Port
- Protocol (TCP/UDP)
- Connection state
- Bytes transferred (if available)

---

### 3. PPPoE Status Checker

**Purpose**: Check if PPPoE user is online dan get connection details

**Flow:**
```
Input: Username (partial match supported)
Select: Router(s) to search
    ↓
Concurrent search di semua selected routers
    ↓
/ppp active print where name~"username"
    ↓
Aggregate results
    ↓
Display: IP, uptime, interface, router
```

**Search Options:**
- Exact match: `user123`
- Partial match: `user` (finds user123, user456, etc.)
- Multi-router search
- Fuzzy search support

---

### 4. Router Management

**Purpose**: Manage multiple MikroTik routers dari satu interface

**Router Properties:**
```go
type Router struct {
    ID             string    // UUID
    Name           string    // Unique identifier
    Host           string    // IP or hostname
    Port           int       // API port (8728/8729)
    Username       string    // MikroTik user
    Password       string    // Encrypted password
    TunnelEndpoint string    // Internal IP:port
    PublicONTURL   string    // Public URL
    Description    string    // Optional notes
    Enabled        bool      // Active status
    CreatedAt      time.Time
    UpdatedAt      time.Time
}
```

**Operations:**
- Create: Add new router with validation
- Read: List all routers (filtered by user permissions)
- Update: Modify router configuration
- Delete: Remove router (soft delete)
- Test: Verify connection health

**Connection Test:**
1. TCP connection check (15s timeout)
2. RouterOS API auth
3. Get system identity
4. Get system resources
5. Return detailed report

---

### 5. User Management

**Purpose**: Role-based access control untuk multi-user

**User Properties:**
```go
type User struct {
    ID        string    // UUID
    Username  string    // Unique
    Password  string    // Bcrypt hashed
    FullName  string
    Email     string
    Role      string    // See roles below
    Active    bool
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

**User-Router Relationship:**
```go
type UserRouterAccess struct {
    UserID     string
    RouterName string
    CreatedAt  time.Time
}
```

Each user can access specific routers only (except Administrator).

---

### 6. Activity Logging

**Purpose**: Comprehensive audit trail untuk compliance & debugging

**Log Properties:**
```go
type ActivityLog struct {
    ID          string    // UUID
    UserID      string    // Who performed action
    Username    string    // Denormalized for reporting
    Action      string    // CREATE/UPDATE/DELETE/LOGIN/etc
    Resource    string    // ROUTER/NAT/USER/etc
    RouterName  string    // Affected router (if applicable)
    Details     string    // JSON details
    IPAddress   string    // Client IP
    UserAgent   string    // Browser/client info
    Status      string    // SUCCESS/FAILED
    ErrorMsg    string    // Error details (if failed)
    CreatedAt   time.Time
}
```

**Logged Actions:**
- User login/logout
- Router CRUD operations
- NAT rule updates
- PPPoE searches
- Configuration changes
- Failed attempts

---

## Database Schema

### Tables Overview

```sql
-- Users table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(50) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,  -- bcrypt hash
    full_name VARCHAR(100) NOT NULL,
    email VARCHAR(100),
    role VARCHAR(50) NOT NULL,       -- Administrator, Head Branch 1/2/3
    active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Routers table
CREATE TABLE routers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) UNIQUE NOT NULL,
    host VARCHAR(255) NOT NULL,
    port INTEGER NOT NULL DEFAULT 8728,
    username VARCHAR(100) NOT NULL,
    password VARCHAR(255) NOT NULL,
    tunnel_endpoint VARCHAR(100),
    public_ont_url VARCHAR(255),
    description TEXT,
    enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- User-Router access mapping
CREATE TABLE user_router_access (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    router_name VARCHAR(100) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(user_id, router_name)
);

-- Activity logs
CREATE TABLE activity_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    username VARCHAR(50),
    action VARCHAR(50) NOT NULL,
    resource VARCHAR(50) NOT NULL,
    router_name VARCHAR(100),
    details TEXT,
    ip_address VARCHAR(50),
    user_agent TEXT,
    status VARCHAR(20) NOT NULL,
    error_message TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Sessions table (optional - currently using in-memory)
CREATE TABLE sessions (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_routers_name ON routers(name);
CREATE INDEX idx_routers_enabled ON routers(enabled);
CREATE INDEX idx_user_router_access_user ON user_router_access(user_id);
CREATE INDEX idx_user_router_access_router ON user_router_access(router_name);
CREATE INDEX idx_activity_logs_user ON activity_logs(user_id);
CREATE INDEX idx_activity_logs_action ON activity_logs(action);
CREATE INDEX idx_activity_logs_created ON activity_logs(created_at DESC);
CREATE INDEX idx_sessions_user ON sessions(user_id);
CREATE INDEX idx_sessions_expires ON sessions(expires_at);
```

### Default Data

```sql
-- Default admin user (password: admin123)
INSERT INTO users (username, password, full_name, role, active)
VALUES (
    'admin',
    '$2a$10$...',  -- bcrypt hash of "admin123"
    'System Administrator',
    'Administrator',
    true
);

-- Default branch users (password: head123)
INSERT INTO users (username, password, full_name, role, active)
VALUES
    ('head1', '$2a$10$...', 'Head Branch 1', 'Head Branch 1', true),
    ('head2', '$2a$10$...', 'Head Branch 2', 'Head Branch 2', true),
    ('head3', '$2a$10$...', 'Head Branch 3', 'Head Branch 3', true);
```

---

## User Roles & Permissions

### Role Hierarchy

```
┌─────────────────────────────────────────────────┐
│         Administrator (admin)                    │
│  - Full access to all routers                   │
│  - User management                              │
│  - Router management                            │
│  - View all activity logs                       │
│  - System configuration                         │
└─────────────────────────────────────────────────┘
                       │
        ┌──────────────┼──────────────┐
        │              │              │
┌───────▼───────┐ ┌───▼───────┐ ┌───▼───────┐
│ Head Branch 1 │ │Head Branch2│ │Head Branch3│
│               │ │            │ │            │
│ Limited access│ │Limited     │ │Limited     │
│ to assigned   │ │access      │ │access      │
│ routers only  │ │            │ │            │
└───────────────┘ └────────────┘ └────────────┘
```

### Permission Matrix

| Feature | Administrator | Head Branch 1/2/3 |
|---------|--------------|-------------------|
| **Router Management** |
| View routers | ✅ All | ✅ Assigned only |
| Add router | ✅ | ❌ |
| Edit router | ✅ | ❌ |
| Delete router | ✅ | ❌ |
| Test connection | ✅ | ✅ |
| **NAT Management** |
| View NAT configs | ✅ All routers | ✅ Assigned routers |
| Update NAT rules | ✅ All routers | ✅ Assigned routers |
| View online clients | ✅ All routers | ✅ Assigned routers |
| **PPPoE Checker** |
| Check PPPoE status | ✅ All routers | ✅ Assigned routers |
| Fuzzy search | ✅ | ✅ |
| **User Management** |
| View users | ✅ | ❌ |
| Create user | ✅ | ❌ |
| Edit user | ✅ | ❌ |
| Delete user | ✅ | ❌ |
| Change password | ✅ (all) | ✅ (own only) |
| **Activity Logs** |
| View all logs | ✅ | ❌ |
| View own logs | ✅ | ✅ |
| Export logs | ✅ | ❌ |
| Delete old logs | ✅ | ❌ |

---

## Project Structure

### Directory Layout

```
nat-management-app/
│
├── cmd/                        # Application entry points
│   └── main.go                 # Main application
│
├── config/                     # Configuration management
│   └── config.go               # Config loader (from .env)
│
├── internal/                   # Private application code
│   │
│   ├── api/                    # HTTP handlers
│   │   ├── auth_handler.go     # Authentication endpoints
│   │   ├── nat_handler.go      # NAT management endpoints
│   │   ├── router_handler.go   # Router CRUD endpoints
│   │   ├── user_handler.go     # User management endpoints
│   │   └── activity_log_handler.go  # Logging endpoints
│   │
│   ├── database/               # Database layer
│   │   └── db.go               # PostgreSQL connection & pooling
│   │
│   ├── middleware/             # HTTP middleware
│   │   ├── auth_middleware.go  # Authentication/authorization
│   │   └── rate_limiter.go     # Rate limiting (in auth_middleware)
│   │
│   ├── models/                 # Data models & DTOs
│   │   ├── user.go             # User model
│   │   ├── router.go           # Router model
│   │   ├── nat.go              # NAT config model
│   │   ├── pppoe.go            # PPPoE session model
│   │   └── activity_log.go     # Activity log model
│   │
│   ├── services/               # Business logic
│   │   ├── auth_service.go     # Authentication service
│   │   ├── auth_service_db.go  # Auth with PostgreSQL
│   │   ├── nat_service.go      # NAT management service
│   │   ├── router_service.go   # Router service interface
│   │   ├── router_service_db.go # Router service with PostgreSQL
│   │   ├── user_service.go     # User management service
│   │   └── activity_log_service.go # Activity logging service
│   │
│   └── utils/                  # Utility functions
│       ├── jwt.go              # JWT token management
│       ├── password.go         # Password hashing (bcrypt)
│       └── validator.go        # Input validation
│
├── web/                        # Frontend assets
│   │
│   ├── static/                 # Static files
│   │   ├── css/                # Stylesheets
│   │   │   ├── base.css        # Base styles
│   │   │   ├── login.css       # Login page styles
│   │   │   └── dashboard.css   # Dashboard styles
│   │   │
│   │   ├── js/                 # JavaScript files
│   │   │   ├── auth.js         # Authentication logic
│   │   │   ├── nat.js          # NAT management logic
│   │   │   ├── pppoe.js        # PPPoE checker logic
│   │   │   ├── routers.js      # Router management logic
│   │   │   └── users.js        # User management logic
│   │   │
│   │   └── images/             # Images & icons
│   │
│   └── templates/              # HTML templates (Gin)
│       ├── base.html           # Base layout
│       ├── login.html          # Login page
│       ├── nat_management.html # NAT management page
│       ├── pppoe_checker.html  # PPPoE status checker
│       ├── router_management.html # Router management
│       ├── user_management.html   # User management
│       └── activity_logs.html     # Activity logs viewer
│
├── tools/                      # Diagnostic & setup tools
│   ├── router-diagnostic.go    # Connection diagnostic tool
│   └── router-setup-wizard.go  # Interactive setup wizard
│
├── docs/                       # Documentation
│   ├── TROUBLESHOOTING.md
│   ├── ROUTER-SETUP.md
│   ├── CONNECTION-FIX-SUMMARY.md
│   ├── PROJECT-OVERVIEW.md     # This file
│   ├── API-REFERENCE.md
│   ├── DEVELOPMENT-GUIDE.md
│   ├── DEPLOYMENT.md
│   └── CHANGELOG.md
│
├── migrations/                 # Database migrations
│   └── init.sql                # Initial schema
│
├── .env.example                # Environment template
├── .gitignore
├── go.mod                      # Go module definition
├── go.sum                      # Dependency checksums
├── build-tools.bat             # Build script for tools
└── README.md                   # Main documentation
```

---

## Component Descriptions

### Backend Components

#### 1. **Handlers (API Layer)**

**Purpose**: Handle HTTP requests, validate input, call services

**Example** (router_handler.go):
```go
type RouterHandler struct {
    routerService  services.RouterService
    natService     services.NATService
    activityLog    services.ActivityLogService
    logger         *logrus.Logger
}

func (h *RouterHandler) GetRouters(c *gin.Context) {
    // 1. Authenticate & authorize
    user := c.MustGet("user").(models.User)

    // 2. Get routers (filtered by user permissions)
    routers, err := h.routerService.GetRoutersForUser(user)

    // 3. Log activity
    h.activityLog.Log(user, "LIST", "ROUTER", ...)

    // 4. Return response
    c.JSON(http.StatusOK, gin.H{"routers": routers})
}
```

#### 2. **Services (Business Logic)**

**Purpose**: Implement business rules, coordinate data access

**Example** (nat_service.go):
```go
type NATService struct {
    routers map[string]RouterConfig
    logger  *logrus.Logger
    mu      sync.RWMutex
}

func (ns *NATService) UpdateNATRule(routerName, username, newIP string, newPort int) error {
    // 1. Connect to router (with retry logic)
    client, err := ns.ConnectRouter(routerName)

    // 2. Find NAT rule by comment
    natRule := ns.FindNATRuleByUsername(client, username)

    // 3. Update dst-address
    err = ns.UpdateDestination(client, natRule.ID, newIP, newPort)

    // 4. Verify changes
    return ns.VerifyUpdate(client, natRule.ID)
}
```

#### 3. **Middleware**

**Purpose**: Cross-cutting concerns (auth, logging, rate limiting)

**Example** (auth_middleware.go):
```go
func (am *AuthMiddleware) RequireJWTAuth() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 1. Extract token from header
        token := c.GetHeader("Authorization")

        // 2. Validate & parse JWT
        claims, err := jwt.ValidateToken(token)

        // 3. Load user from claims
        user, err := am.authService.GetUserByID(claims.UserID)

        // 4. Check rate limit
        if !am.rateLimiter.Allow(user.ID) {
            c.AbortWithStatusJSON(429, gin.H{"error": "rate limit exceeded"})
            return
        }

        // 5. Set user in context
        c.Set("user", user)
        c.Next()
    }
}
```

### Frontend Components

#### 1. **Templates (HTML)**

**Purpose**: Server-side rendered pages with Gin templating

**Features:**
- Base layout dengan navigation sidebar
- Responsive design (mobile-first)
- Version badge (v4.1)
- Role-based UI rendering

#### 2. **JavaScript Modules**

**Purpose**: Client-side logic & API communication

**Common Pattern:**
```javascript
// auth.js - JWT management
class AuthManager {
    constructor() {
        this.token = localStorage.getItem('jwt_token');
    }

    async login(username, password) {
        const response = await fetch('/api/auth/login', {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify({username, password})
        });

        const data = await response.json();
        this.token = data.token;
        localStorage.setItem('jwt_token', data.token);
        return data;
    }

    getAuthHeader() {
        return {'Authorization': `Bearer ${this.token}`};
    }
}
```

---

## Data Flow

### Example: Update NAT Rule

```
┌─────────┐     1. Submit Form      ┌──────────────┐
│  User   │────────────────────────→│   Browser    │
│  (Web)  │                          │  (nat.js)    │
└─────────┘                          └──────┬───────┘
                                            │
                          2. POST /api/nat/update
                          (JWT in Authorization header)
                                            │
                                    ┌───────▼────────┐
                                    │  Gin Router    │
                                    │  (middleware)  │
                                    └───────┬────────┘
                                            │
                          3. Validate JWT & permissions
                                            │
                                    ┌───────▼────────┐
                                    │  NATHandler    │
                                    │  (api layer)   │
                                    └───────┬────────┘
                                            │
                          4. Call UpdateNATRule()
                                            │
                                    ┌───────▼────────┐
                                    │  NATService    │
                                    │  (business)    │
                                    └───────┬────────┘
                                            │
                          5. Connect to router (retry logic)
                                            │
                              ┌─────────────┴──────────────┐
                              │                            │
                      ┌───────▼────────┐       ┌──────────▼─────────┐
                      │  RouterService │       │  MikroTik Router   │
                      │  (get config)  │       │  (RouterOS API)    │
                      └───────┬────────┘       └──────────┬─────────┘
                              │                            │
                              │  6. Execute RouterOS       │
                              │     commands               │
                              └────────────────────────────┘
                                            │
                          7. Verify update & log activity
                                            │
                                    ┌───────▼────────┐
                                    │ ActivityLog    │
                                    │ Service        │
                                    └───────┬────────┘
                                            │
                          8. Insert log to PostgreSQL
                                            │
                                    ┌───────▼────────┐
                                    │  PostgreSQL    │
                                    │  (Neon)        │
                                    └────────────────┘
```

---

## Security Model

### Authentication Flow

```
1. User Login
   ↓
2. Validate credentials (bcrypt compare)
   ↓
3. Generate JWT token (24h expiry)
   ↓
4. Generate refresh token (7d expiry)
   ↓
5. Return both tokens to client
   ↓
6. Client stores tokens (localStorage)
   ↓
7. Client sends JWT in Authorization header for API calls
   ↓
8. Server validates JWT signature & expiry
   ↓
9. Token expired? → Use refresh token
   ↓
10. Refresh successful → New JWT issued
```

### Security Features

1. **Password Security**
   - Bcrypt hashing (cost 10)
   - Minimum 6 characters
   - No password in logs

2. **JWT Security**
   - HMAC-SHA256 signing
   - 24h access token expiry
   - 7d refresh token expiry
   - Secure secret (min 32 chars)

3. **API Security**
   - CORS whitelist
   - Rate limiting (100 req/min per user)
   - Request validation
   - SQL injection prevention (parameterized queries)

4. **Router Credentials**
   - Stored encrypted in database
   - Never exposed in API responses
   - Only accessible by authorized users

5. **Session Security**
   - Session timeout (24h default)
   - Secure cookies (if HTTPS)
   - CSRF protection (implicit via JWT)

---

## Performance Considerations

### Connection Management

**Problem:** Setiap NAT operation membuka koneksi baru ke router (slow)

**Solution (v4.1):**
- Retry logic dengan exponential backoff
- Timeout increased (15-45s vs 5s)
- Detailed error logging

**Future Enhancement (Planned):**
- Connection pooling
- Background health monitoring
- In-memory cache (60s TTL)

### Database Performance

**Current:**
- Connection pooling (pgx)
- Indexes on frequently queried columns
- Prepared statements

**Optimizations:**
- Denormalized fields (username in logs)
- Paginated queries
- Efficient JOIN strategies

---

**Version:** 4.1
**Last Updated:** 2025-10-16
**Maintained by:** NAT Management Team
