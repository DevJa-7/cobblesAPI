package server

import (
	"github.com/jinzhu/gorm"
)

// Follower ..
type Follower struct {
	gorm.Model
	UserID         int32 `gorm:"index"`
	FollowerUserID int32 `gorm:"index"`
}

// CreateFollower ...
func (s *Server) CreateFollower(
	userID int32,
	followerUserID int32,
) (*Follower, error) {
	db := s.DB
	follower := &Follower{UserID: userID, FollowerUserID: followerUserID}

	if err := db.Create(&follower).Error; err != nil {
		return nil, err
	}

	return follower, nil

}

// GetFollowingByUserID  - Who am I following
func (s *Server) GetFollowingByUserID(userID int32) (users []*User, err error) {
	if &userID == nil {
		return nil, nil
	}

	db := s.DB
	followingArray := []Follower{}

	if err := db.Find(&followingArray, "user_id = ?", userID).Error; err != nil {
		return nil, err
	}

	for _, follower := range followingArray {
		user, err := s.UserByID(int64(follower.FollowerUserID))
		if err != nil {
			return nil, err
		}

		duplicate := false
		for _, u := range users {
			if u.ID == user.ID {
				duplicate = true
				break
			}
		}
		if !duplicate {
			users = append(users, user)
		}

	}

	// Update Following count
	followPtr := new(int32)
	*followPtr = int32(len(followingArray))
	user := new(User)
	db.First(&user, userID)
	user.Following = followPtr
	db.Save(&user)

	return users, nil

}

// GetFollowersByUserID  - Who are my followers
func (s *Server) GetFollowersByUserID(userID int32) (users []*User, err error) {
	if &userID == nil {
		return nil, nil
	}

	db := s.DB
	followerArray := []Follower{}

	if err := db.Find(&followerArray, "follower_user_id = ?", userID).Error; err != nil {
		return nil, err
	}

	for _, follower := range followerArray {
		user, err := s.UserByID(int64(follower.UserID))
		if err != nil {
			return nil, err
		}

		duplicate := false
		for _, u := range users {
			if u.ID == user.ID {
				duplicate = true
				break
			}
		}
		if !duplicate {
			users = append(users, user)
		}

	}

	// Update Follower count
	followPtr := new(int32)
	*followPtr = int32(len(followerArray))
	user := new(User)
	db.First(&user, userID)
	user.Followers = followPtr
	db.Save(&user)

	return users, nil

}

// Unfollow ...
func (s *Server) Unfollow(
	userID int32,
	followerUserID int32, // User to unfollow
) (bool, error) {
	db := s.DB
	follower := &Follower{}
	db.Where("user_id = ? and follower_user_id = ?", userID, followerUserID).Delete(&follower)

	if err := db.Create(&follower).Error; err != nil {
		return false, err
	}

	return true, nil

}

// RecalculateFollowerAndFollowing -
func (s *Server) RecalculateFollowerAndFollowing(userID int32) (bool, error) {
	if &userID == nil {
		return false, nil
	}
	// Recalculate following count
	db := s.DB
	followingArray := []Follower{}

	if err := db.Find(&followingArray, "user_id = ?", userID).Error; err != nil {
		return false, err
	}

	followerArray := []Follower{}

	if err := db.Find(&followerArray, "follower_user_id = ?", userID).Error; err != nil {
		return false, err
	}

	// Update Following & followers count
	followingPtr := new(int32)
	followersPtr := new(int32)
	*followingPtr = int32(len(followingArray))
	*followersPtr = int32(len(followerArray))

	user := new(User)
	db.First(&user, userID)
	user.Following = followingPtr
	user.Followers = followersPtr

	if err := db.Save(&user).Error; err != nil {
		return false, err
	}

	return true, nil

}
