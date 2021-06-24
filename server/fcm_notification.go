package server

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/maddevsio/fcm"
)

// SendNotification : send notification to firebase.
func (s *Server) SendNotification(fcmToken string, msgType string, msgID int64, msgBody string, senderName string, convID int64, userID int64) error {
	fcmAPIKey := os.Getenv("FCM_API_KEY")

	if fcmAPIKey == "" {
		fcmAPIKey = "AAAAxZpGrIU:APA91bFjZWce9Eq2hh9fXsi4sxZIfZ5k2Q90QxnIoe5qjdsrULPBMU2DS7izDqDo8NdJLVoEFTE8iSzrAwhBW9K-qjJGin7vMlmqksAqppmJJxlmydvvgQQxyrbIQqRTF3xiBijkwmXx"
	}
	data := map[string]string{
		"messageID": strconv.FormatInt(msgID, 10),
		"senderID":  strconv.FormatInt(userID, 10),
		"convID":    strconv.FormatInt(convID, 10),
		"time":      strconv.FormatInt(time.Now().Unix(), 10),
		"message":   msgBody,
		"type":      msgType,
	}

	log.Println("##############fcmAPIKey ===", fcmAPIKey)
	c := fcm.NewFCM(fcmAPIKey)
	token := fcmToken

	response, err := c.Send(fcm.Message{
		Data:             data,
		RegistrationIDs:  []string{token},
		ContentAvailable: true,
		Priority:         fcm.PriorityHigh,
		Notification: fcm.Notification{
			Title: senderName,
			Body:  msgBody,
		},
	})
	if err != nil {
		log.Fatal(err)
		return err
	}
	fmt.Println("Status Code   :", response.StatusCode)
	
	return nil
}
