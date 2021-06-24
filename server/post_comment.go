package server

import (
	"time"
)

// PostComment ...
type PostComment struct {
	// gorm.Model
	ID              int64
	Comment         string
	ParentCommentID int32 `gorm:"index"`
	UserID          int32 `gorm:"index"`
	PostID          int32 `gorm:"index"`

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

// CreatePostComment ...
func (s *Server) CreatePostComment(
	parentCommentID,
	userID,
	postID int32,
	comment string) (*PostComment, error) {
	db := s.DB
	pc := &PostComment{
		Comment:         comment,
		UserID:          userID,
		PostID:          postID,
		ParentCommentID: parentCommentID,
	}

	if err := db.Create(&pc).Error; err != nil {
		return nil, err
	}

	_, err := s.RecalculatePostCommentCount(postID)
	if err != nil {
		return nil, err
	}

	return pc, nil
}

// RemovePostComment ...
func (s *Server) RemovePostComment(commentID, postID int32) error {
	db := s.DB

	if err := db.Unscoped().Where("id = ?", commentID).Delete(&PostComment{}).Error; err != nil {
		return err
	}
	_, err := s.RecalculatePostCommentCount(postID)
	if err != nil {
		return err
	}

	return nil
}

// RecalculatePostCommentCount ...
// Todo
func (s *Server) RecalculatePostCommentCount(postID int32) (bool, error) {

	// Recalculate Likes count
	db := s.DB
	PostCommentArray := []PostComment{}

	if err := db.Find(&PostCommentArray, "post_id = ?", postID).Error; err != nil {
		return false, err
	}

	if err := db.Exec("update posts set comment_count = ? where id = ?", len(PostCommentArray), postID).Error; err != nil {
		return false, err
	}

	return true, nil
}
