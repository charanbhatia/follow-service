package models

import "time"

type User struct {
	ID        int32
	Username  string
	Email     string
	CreatedAt time.Time
}

type Follow struct {
	FollowerID  int32
	FollowingID int32
	CreatedAt   time.Time
}
