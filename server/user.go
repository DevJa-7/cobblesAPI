package server

import (
	"time"

	"github.com/jackc/pgx/pgtype"
)

type User struct {
	ID          int64
	Name        *string
	PhoneNumber *string
	ZIPCode     *string
	PhotoURL    *string
	Bio         *string
	FCMToken    *string
	Followers   *int32
	Following   *int32
	PostCount   int32

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (s *Server) UserByID(userID int64) (*User, error) {
	// Recalculate follower & following value
	_, err := s.RecalculateFollowerAndFollowing(int32(userID))
	if err != nil {
		return nil, err
	}

	var u User
	var result struct {
		createdAt pgtype.Timestamptz
		updatedAt pgtype.Timestamptz
	}
	err = s.ConnPool.QueryRow(`
		select
			id,
			name,
			bio,
			phone_number,
			zip_code,
			photo_url,
			followers,
			following,
			post_count,
			fcm_token,
			created_at,
			updated_at
		from users
		where id = $1
	`, userID).Scan(
		&u.ID,
		&u.Name,
		&u.Bio,
		&u.PhoneNumber,
		&u.ZIPCode,
		&u.PhotoURL,
		&u.Followers,
		&u.Following,
		&u.PostCount,
		&u.FCMToken,
		&result.createdAt,
		&result.updatedAt,
	)
	switch {
	// case err == pgx.ErrNoRows:
	// 	return nil, errors.New("request is not authenticated to a user")
	case err != nil:
		return nil, err
	}

	u.CreatedAt = result.createdAt.Time
	u.UpdatedAt = result.updatedAt.Time

	return &u, nil
}

// RecalculatePostCount : calculate post count according to user id.
func (s *Server) RecalculatePostCount(userID int64) (bool, error) {
	// Recalculate Post count
	db := s.DB

	var cnt int64
	err := s.ConnPool.QueryRow(`
		select count(*) from posts where user_id = $1 and removed = false
	`, userID).Scan(
		&cnt,
	)

	if err != nil {
		return false, err
	}

	if err := db.Exec("update users set post_count = ? where id = ?", cnt, userID).Error; err != nil {
		return false, err
	}

	return true, nil
}
