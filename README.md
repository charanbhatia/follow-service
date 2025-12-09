# Follow Service

A production-ready gRPC microservice for managing user follow/unfollow relationships.

## Stack

- Go 1.22
- PostgreSQL 16
- gRPC with Protocol Buffers
- Docker

## Features

- Follow/Unfollow operations
- Paginated follower and following lists
- User management
- Health check endpoints
- Structured logging
- Error recovery middleware
- Database migrations

## Technology Decisions

**Go**: High performance, strong typing, excellent gRPC support, efficient concurrency model.

**PostgreSQL**: ACID compliance, optimized indexing for social graph queries, production-proven reliability.

**gRPC**: Binary protocol for performance, type-safe contracts, efficient service-to-service communication.

## Database Schema

```sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE follows (
    follower_id INTEGER NOT NULL,
    following_id INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (follower_id, following_id),
    FOREIGN KEY (follower_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (following_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT no_self_follow CHECK (follower_id != following_id)
);

CREATE INDEX idx_follows_follower ON follows(follower_id);
CREATE INDEX idx_follows_following ON follows(following_id);
```

**Design Decisions:**
- Composite primary key prevents duplicate follows
- Two indexes optimize both "followers" and "following" queries
- CHECK constraint prevents self-following at database level
- CASCADE delete maintains referential integrity

## Quick Start

```bash
# Clone and start
git clone https://github.com/charanbhatia/follow-service.git
cd follow-service
docker-compose up -d

# Health check
curl http://localhost:8080/health/ready
```

**Endpoints:**
- gRPC: `localhost:50051`
- Health: `localhost:8080/health/ready`

**Environment:**
```
DATABASE_URL=postgres://user:pass@host:5432/dbname?sslmode=disable
GRPC_PORT=50051
HEALTH_PORT=8080
```

## API Reference

### gRPC Methods

```protobuf
service FollowService {
  rpc Follow(FollowRequest) returns (FollowResponse);
  rpc Unfollow(UnfollowRequest) returns (UnfollowResponse);
  rpc GetFollowers(GetFollowersRequest) returns (GetFollowersResponse);
  rpc GetFollowing(GetFollowingRequest) returns (GetFollowingResponse);
  rpc GetUser(GetUserRequest) returns (GetUserResponse);
  rpc ListUsers(ListUsersRequest) returns (ListUsersResponse);
}
```

### Testing with grpcurl

```bash
# List all services
grpcurl -plaintext localhost:50051 list

# List users
grpcurl -plaintext -d '{"limit": 10}' localhost:50051 follow.FollowService/ListUsers

# Follow a user
grpcurl -plaintext -d '{"follower_id": 1, "following_id": 2}' \
  localhost:50051 follow.FollowService/Follow

# Get followers
grpcurl -plaintext -d '{"user_id": 2, "limit": 10}' \
  localhost:50051 follow.FollowService/GetFollowers

# Unfollow
grpcurl -plaintext -d '{"follower_id": 1, "following_id": 2}' \
  localhost:50051 follow.FollowService/Unfollow
```

## Error Handling

The service handles the following error cases:

- **Invalid User IDs**: Returns `NOT_FOUND` if user doesn't exist
- **Duplicate Follows**: Returns `ALREADY_EXISTS` if already following
- **Self-Following**: Returns `INVALID_ARGUMENT` if trying to follow yourself
- **Database Errors**: Returns `INTERNAL` with logged details

## Production Capabilities

**Health Checks**: Liveness and readiness probes for orchestration.

**Logging**: Structured JSON logs with request tracing.

**Middleware**: Panic recovery and request logging interceptors.

**Security**: Input validation, prepared statements, database-level constraints.

## Deployment

**Docker Image:**
```bash
docker pull charanbhatia/follow-service:latest
```

## Project Structure

```
follow-service/
├── cmd/server/          # Application entry point
├── internal/
│   ├── database/       # Database connection & migrations
│   ├── handler/        # gRPC service handlers
│   ├── health/         # Health check endpoints
│   ├── middleware/     # gRPC interceptors
│   ├── models/         # Data models
│   └── repository/     # Data access layer
├── migrations/         # SQL migration files
├── proto/follow/       # Protocol Buffer definitions
├── Dockerfile          # Multi-stage Docker build
└── docker-compose.yml  # Local development setup
```

## Development

**Build:**
```bash
go build -o follow-service ./cmd/server
```

**Regenerate proto:**
```bash
protoc --go_out=. --go-grpc_out=. proto/follow/follow.proto
```

## Scalability

- Indexed database queries for performance
- Pagination for large result sets
- Connection pooling
- Stateless design for horizontal scaling
- Health checks for orchestration

## Author

Charan Bhatia - [GitHub](https://github.com/charanbhatia)
