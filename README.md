# Follow Service

A production-ready gRPC microservice for managing user follow/unfollow relationships. Part of a distributed microservices architecture deployed on AWS.

## Live Deployment

**Production Environment:**
- GraphQL API: http://13.232.23.156:8080
- Follow Service: Private subnet (10.0.1.124:50051)
- Database: AWS RDS PostgreSQL 16.6
- Region: Asia Pacific (Mumbai) - ap-south-1

## Stack

- Go 1.22
- PostgreSQL 16.6 (AWS RDS)
- gRPC with Protocol Buffers
- Docker (deployed on Amazon Linux 2023)
- AWS EC2 (t3.micro)

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
CREATE TABLE follows (
    follower_id VARCHAR(255) NOT NULL,
    following_id VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (follower_id, following_id)
);

CREATE INDEX idx_follows_following_id ON follows(following_id);
CREATE INDEX idx_follows_follower_id ON follows(follower_id);
```

**Design Decisions:**
- Composite primary key prevents duplicate follows (O(1) lookup)
- Two B-tree indexes optimize both "followers" and "following" queries (O(log n))
- Self-follow prevention enforced at application layer
- VARCHAR IDs support flexible user identification schemes
- No foreign key constraints for decoupled microservices architecture

## Quick Start

### Local Development

```bash
# Clone repository
git clone https://github.com/charanbhatia/follow-service.git
cd follow-service

# Set environment variables
export DATABASE_URL="postgres://user:pass@localhost:5432/followdb?sslmode=disable"

# Run locally
go run cmd/server/main.go

# Or with Docker
docker build -t follow-service .
docker run -p 50051:50051 -p 8080:8080 \
  -e DATABASE_URL="postgres://user:pass@host:5432/followdb?sslmode=require" \
  follow-service
```

### Production Deployment

```bash
# Build and push to AWS ECR
docker build -t follow-service:latest .
aws ecr get-login-password --region ap-south-1 | docker login --username AWS --password-stdin 723402273260.dkr.ecr.ap-south-1.amazonaws.com
docker tag follow-service:latest 723402273260.dkr.ecr.ap-south-1.amazonaws.com/follow-service:latest
docker push 723402273260.dkr.ecr.ap-south-1.amazonaws.com/follow-service:latest
```

**Endpoints:**
- gRPC: Port 50051
- Health: Port 8080

**Environment Variables:**
```bash
DATABASE_URL=postgres://user:pass@host:5432/dbname?sslmode=require
```

## API Reference

### gRPC Methods

```protobuf
service FollowService {
  rpc Follow(FollowRequest) returns (FollowResponse);
  rpc Unfollow(UnfollowRequest) returns (UnfollowResponse);
  rpc IsFollowing(IsFollowingRequest) returns (IsFollowingResponse);
  rpc GetFollowers(GetFollowersRequest) returns (GetFollowersResponse);
  rpc GetFollowing(GetFollowingRequest) returns (GetFollowingResponse);
  rpc GetFollowerCount(GetFollowerCountRequest) returns (GetFollowerCountResponse);
  rpc GetFollowingCount(GetFollowingCountRequest) returns (GetFollowingCountResponse);
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
