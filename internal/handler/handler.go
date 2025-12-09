package handler

import (
	"context"
	"errors"

	pb "github.com/pratilipi/follow-service/proto/follow"
	"github.com/pratilipi/follow-service/internal/models"
	"github.com/pratilipi/follow-service/internal/repository"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type FollowServiceServer struct {
	pb.UnimplementedFollowServiceServer
	repo   *repository.Repository
	logger *zap.Logger
}

func NewFollowServiceServer(repo *repository.Repository, logger *zap.Logger) *FollowServiceServer {
	return &FollowServiceServer{
		repo:   repo,
		logger: logger,
	}
}

func (s *FollowServiceServer) Follow(ctx context.Context, req *pb.FollowRequest) (*pb.FollowResponse, error) {
	if req.FollowerId <= 0 || req.FollowingId <= 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid user IDs")
	}

	err := s.repo.Follow(ctx, req.FollowerId, req.FollowingId)
	if err != nil {
		return nil, mapError(err)
	}

	s.logger.Info("user followed successfully",
		zap.Int32("follower_id", req.FollowerId),
		zap.Int32("following_id", req.FollowingId),
	)

	return &pb.FollowResponse{
		Success: true,
		Message: "successfully followed user",
	}, nil
}

func (s *FollowServiceServer) Unfollow(ctx context.Context, req *pb.UnfollowRequest) (*pb.UnfollowResponse, error) {
	if req.FollowerId <= 0 || req.FollowingId <= 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid user IDs")
	}

	err := s.repo.Unfollow(ctx, req.FollowerId, req.FollowingId)
	if err != nil {
		return nil, mapError(err)
	}

	s.logger.Info("user unfollowed successfully",
		zap.Int32("follower_id", req.FollowerId),
		zap.Int32("following_id", req.FollowingId),
	)

	return &pb.UnfollowResponse{
		Success: true,
		Message: "successfully unfollowed user",
	}, nil
}

func (s *FollowServiceServer) GetFollowers(ctx context.Context, req *pb.GetFollowersRequest) (*pb.GetFollowersResponse, error) {
	if req.UserId <= 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid user ID")
	}

	limit := req.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	offset := req.Offset
	if offset < 0 {
		offset = 0
	}

	users, total, err := s.repo.GetFollowers(ctx, req.UserId, limit, offset)
	if err != nil {
		return nil, mapError(err)
	}

	pbUsers := make([]*pb.User, len(users))
	for i, user := range users {
		pbUsers[i] = userToProto(user)
	}

	return &pb.GetFollowersResponse{
		Followers: pbUsers,
		Total:     total,
	}, nil
}

func (s *FollowServiceServer) GetFollowing(ctx context.Context, req *pb.GetFollowingRequest) (*pb.GetFollowingResponse, error) {
	if req.UserId <= 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid user ID")
	}

	limit := req.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	offset := req.Offset
	if offset < 0 {
		offset = 0
	}

	users, total, err := s.repo.GetFollowing(ctx, req.UserId, limit, offset)
	if err != nil {
		return nil, mapError(err)
	}

	pbUsers := make([]*pb.User, len(users))
	for i, user := range users {
		pbUsers[i] = userToProto(user)
	}

	return &pb.GetFollowingResponse{
		Following: pbUsers,
		Total:     total,
	}, nil
}

func (s *FollowServiceServer) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	if req.UserId <= 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid user ID")
	}

	user, err := s.repo.GetUser(ctx, req.UserId)
	if err != nil {
		return nil, mapError(err)
	}

	return &pb.GetUserResponse{
		User: userToProto(user),
	}, nil
}

func (s *FollowServiceServer) ListUsers(ctx context.Context, req *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	limit := req.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	offset := req.Offset
	if offset < 0 {
		offset = 0
	}

	users, total, err := s.repo.ListUsers(ctx, limit, offset)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to list users")
	}

	pbUsers := make([]*pb.User, len(users))
	for i, user := range users {
		pbUsers[i] = userToProto(user)
	}

	return &pb.ListUsersResponse{
		Users: pbUsers,
		Total: total,
	}, nil
}

func mapError(err error) error {
	switch {
	case errors.Is(err, repository.ErrUserNotFound):
		return status.Error(codes.NotFound, "user not found")
	case errors.Is(err, repository.ErrAlreadyFollowing):
		return status.Error(codes.AlreadyExists, "already following this user")
	case errors.Is(err, repository.ErrNotFollowing):
		return status.Error(codes.NotFound, "not following this user")
	case errors.Is(err, repository.ErrSelfFollow):
		return status.Error(codes.InvalidArgument, "cannot follow yourself")
	default:
		return status.Error(codes.Internal, "internal server error")
	}
}

func userToProto(user *models.User) *pb.User {
	return &pb.User{
		Id:             user.ID,
		Username:       user.Username,
		Email:          user.Email,
		FollowersCount: user.FollowersCount,
		FollowingCount: user.FollowingCount,
	}
}
