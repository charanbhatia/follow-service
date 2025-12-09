package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/pratilipi/follow-service/internal/models"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrAlreadyFollowing  = errors.New("already following this user")
	ErrNotFollowing      = errors.New("not following this user")
	ErrSelfFollow        = errors.New("cannot follow yourself")
	ErrDuplicateUsername = errors.New("username already exists")
)

type Repository struct {
	db *sql.DB
}

func New(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetUser(ctx context.Context, userID int32) (*models.User, error) {
	query := `SELECT id, username, email, followers_count, following_count, created_at FROM users WHERE id = $1`
	
	var user models.User
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.FollowersCount,
		&user.FollowingCount,
		&user.CreatedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	
	return &user, nil
}

func (r *Repository) ListUsers(ctx context.Context, limit, offset int32) ([]*models.User, int32, error) {
	countQuery := `SELECT COUNT(*) FROM users`
	var total int32
	if err := r.db.QueryRowContext(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	query := `SELECT id, username, email, followers_count, following_count, created_at FROM users ORDER BY id LIMIT $1 OFFSET $2`
	
	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()
	
	users := make([]*models.User, 0)
	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.FollowersCount, &user.FollowingCount, &user.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, &user)
	}
	
	return users, total, nil
}

func (r *Repository) Follow(ctx context.Context, followerID, followingID int32) error {
	if followerID == followingID {
		return ErrSelfFollow
	}

	if _, err := r.GetUser(ctx, followerID); err != nil {
		return err
	}
	if _, err := r.GetUser(ctx, followingID); err != nil {
		return err
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	query := `INSERT INTO follows (follower_id, following_id) VALUES ($1, $2)`
	_, err = tx.ExecContext(ctx, query, followerID, followingID)
	
	if err != nil {
		if isUniqueViolation(err) {
			return ErrAlreadyFollowing
		}
		return fmt.Errorf("failed to create follow: %w", err)
	}

	_, err = tx.ExecContext(ctx, `UPDATE users SET following_count = following_count + 1 WHERE id = $1`, followerID)
	if err != nil {
		return fmt.Errorf("failed to update following count: %w", err)
	}

	_, err = tx.ExecContext(ctx, `UPDATE users SET followers_count = followers_count + 1 WHERE id = $1`, followingID)
	if err != nil {
		return fmt.Errorf("failed to update followers count: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	
	return nil
}

func (r *Repository) Unfollow(ctx context.Context, followerID, followingID int32) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	query := `DELETE FROM follows WHERE follower_id = $1 AND following_id = $2`
	
	result, err := tx.ExecContext(ctx, query, followerID, followingID)
	if err != nil {
		return fmt.Errorf("failed to delete follow: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return ErrNotFollowing
	}

	_, err = tx.ExecContext(ctx, `UPDATE users SET following_count = following_count - 1 WHERE id = $1`, followerID)
	if err != nil {
		return fmt.Errorf("failed to update following count: %w", err)
	}

	_, err = tx.ExecContext(ctx, `UPDATE users SET followers_count = followers_count - 1 WHERE id = $1`, followingID)
	if err != nil {
		return fmt.Errorf("failed to update followers count: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	
	return nil
}

func (r *Repository) GetFollowers(ctx context.Context, userID, limit, offset int32) ([]*models.User, int32, error) {
	countQuery := `SELECT COUNT(*) FROM follows WHERE following_id = $1`
	var total int32
	if err := r.db.QueryRowContext(ctx, countQuery, userID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count followers: %w", err)
	}

	query := `
		SELECT u.id, u.username, u.email, u.followers_count, u.following_count, u.created_at 
		FROM users u
		INNER JOIN follows f ON u.id = f.follower_id
		WHERE f.following_id = $1
		ORDER BY f.created_at DESC
		LIMIT $2 OFFSET $3
	`
	
	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get followers: %w", err)
	}
	defer rows.Close()
	
	users := make([]*models.User, 0)
	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.FollowersCount, &user.FollowingCount, &user.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("failed to scan follower: %w", err)
		}
		users = append(users, &user)
	}
	
	return users, total, nil
}

func (r *Repository) GetFollowing(ctx context.Context, userID, limit, offset int32) ([]*models.User, int32, error) {
	countQuery := `SELECT COUNT(*) FROM follows WHERE follower_id = $1`
	var total int32
	if err := r.db.QueryRowContext(ctx, countQuery, userID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count following: %w", err)
	}

	query := `
		SELECT u.id, u.username, u.email, u.followers_count, u.following_count, u.created_at 
		FROM users u
		INNER JOIN follows f ON u.id = f.following_id
		WHERE f.follower_id = $1
		ORDER BY f.created_at DESC
		LIMIT $2 OFFSET $3
	`
	
	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get following: %w", err)
	}
	defer rows.Close()
	
	users := make([]*models.User, 0)
	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.FollowersCount, &user.FollowingCount, &user.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("failed to scan following: %w", err)
		}
		users = append(users, &user)
	}
	
	return users, total, nil
}

func isUniqueViolation(err error) bool {
	return err != nil && (
		err.Error() == `pq: duplicate key value violates unique constraint "follows_pkey"` ||
		err.Error() == `duplicate key value violates unique constraint "follows_pkey"`)
}
