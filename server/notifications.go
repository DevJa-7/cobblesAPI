package server

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sns"
)

func (s *Server) PublishNotificationToUser(userID int64, message string) error {
	rows, err := s.ConnPool.Query(`
		select id, endpoint_arn from user_device_tokens
		where user_id = $1
	`, userID)
	if err != nil {
		return err
	}
	defer rows.Close()

	var endpointArns []string
	for rows.Next() {
		var result struct {
			id          int64
			endpointArn string
		}
		if err := rows.Scan(&result.id, &result.endpointArn); err != nil {
			return err
		}

		endpointArns = append(endpointArns, result.endpointArn)
	}

	if err := rows.Err(); err != nil {
		return err
	}
	log.Println("##########")
	for _, endpointArn := range endpointArns {
		_, err := s.SNS.Publish(&sns.PublishInput{
			Message:   aws.String(message),
			TargetArn: aws.String(endpointArn),
		})
		if err != nil {
			// some errors may be from disabled tokens, watch and deal with later
			log.Println("@@@@@@@error", err)
			// return err
			continue
		}

	}

	return nil
}
