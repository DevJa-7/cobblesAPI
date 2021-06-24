package server

import (
	"github.com/jinzhu/gorm"
)

// Like ...
type Like struct {
	gorm.Model
	UserID int32 `gorm:"index"`
	PostID int32 `gorm:"index"`
}

// Like a post
func (s *Server) Like(userID, postID int32) (*Like, error) {
	db := s.DB
	like := &Like{UserID: userID, PostID: postID}

	if err := db.Create(&like).Error; err != nil {
		return nil, err
	}

	_, err := s.RecalculateLikeCount(postID)
	if err != nil {
		return nil, err
	}

	return like, nil
}

// Unlike a post
func (s *Server) Unlike(userID, postID int32) (bool, error) {
	db := s.DB
	like := &Like{UserID: userID, PostID: postID}

	if err := db.Unscoped().Delete(&like).Error; err != nil {
		return false, err
	}

	_, err := s.RecalculateLikeCount(postID)
	if err != nil {
		return false, err
	}

	return true, nil
}

// RecalculateLikeCount ...
func (s *Server) RecalculateLikeCount(postID int32) (bool, error) {
	// Recalculate Likes count
	db := s.DB
	LikesArray := []Like{}

	if err := db.Find(&LikesArray, "post_id = ?", postID).Error; err != nil {
		return false, err
	}

	if err := db.Exec("update posts set likes = ? where id = ?", len(LikesArray), postID).Error; err != nil {
		return false, err
	}

	return true, nil
}

// HasUserLikedPost ...
func (s *Server) HasUserLikedPost(userID, postID int32) (bool, error) {

	db := s.DB
	like := Like{}
	if err := db.Where("user_id = ? AND post_id = ?", userID, postID).Find(&like).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// Get whether user like that post or not
func (s *Server) GetLike(userID int32, postID int32) (bool, error) {
	var _id int64
	err := s.ConnPool.QueryRow(`
		select
			id
		from likes
		where user_id = $1 and post_id = $2
	`, userID, postID).Scan(
		&_id,
	)

	if err != nil {
		return false, err
	}

	return true, nil

}
