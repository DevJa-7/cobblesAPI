package server

import "time"

type PostKind string

type MediaMetadata struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

type Post struct {
	ID int64

	UserID int64

	NeighborhoodID int64

	Kind PostKind

	Title       string
	Description *string

	Poster *string

	UploadedMediaURL *string
	UploadedMedia    *PostMedia

	Preview *PostMedia
	Media   *PostMedia

	Tags []string

	Processing bool

	ViewTimes	 *int64 `gorm:"default:0"`
	Likes        *int32 `gorm:"default:0"`
	CommentCount *int32 `gorm:"default:0"`

	CreatedAt time.Time
	UpdatedAt time.Time
}
