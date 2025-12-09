package models

import "time"

type User struct {
	ID             int32
	Username       string
	Email          string
	FollowersCount int32
	FollowingCount int32
	CreatedAt      time.Time
}

type Follow struct {
	FollowerID  int32
	FollowingID int32
	CreatedAt   time.Time
}
