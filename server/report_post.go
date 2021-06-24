package server

import (
	"database/sql"

	"github.com/jinzhu/gorm"
)

// ReportedPost Model
type ReportedPost struct {
	gorm.Model

	PostID              int32
	PersonReportingID   int32
	ReportingReasonText string
	ActionTaken         sql.NullInt64
}

// CreateReportedPost - to create reporting post
func (s *Server) CreateReportedPost(
	postID int32,
	personReportingID int32,
	reportingReasonText string,
) (*ReportedPost, error) {

	db := s.DB
	reportedPost := &ReportedPost{
		PostID:              postID,
		PersonReportingID:   personReportingID,
		ReportingReasonText: reportingReasonText,
	}

	if err := db.Create(&reportedPost).Error; err != nil {
		return nil, err
	}

	return reportedPost, nil
}

// UpdateReportedPost - to update action taken
func (s *Server) UpdateReportedPost(
	id int64,
	actionTaken sql.NullInt64,
) (*ReportedPost, error) {
	if &id == nil {
		return nil, nil
	}

	db := s.DB
	reportedPost := ReportedPost{}
	if err := db.First(&reportedPost, id).Error; err != nil {
		return nil, err
	}

	reportedPost.ActionTaken = actionTaken
	db.Save(&reportedPost)

	return &reportedPost, nil

}
