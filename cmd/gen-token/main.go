// quick app to generate jwt token for local testing ;)
package main

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/lambdacollective/cobbles-api/server"
)

func main() {
	secret := os.Getenv("SERVER_SECRET")
	if secret == "" {
		panic("SERVER_SECRET missing")
	}

	s := server.Server{ServerSecret: secret}
	var userID int64
	var err error
	if os.Getenv("USER_ID") != "" {
		userID, err = strconv.ParseInt(os.Getenv("USER_ID"), 10, 64)
		if err != nil {
			panic(err)
		}
	} else {
		userID = rand.Int63n(100000000)
	}

	token, err := s.GenerateAuthJWT(userID)
	if err != nil {
		panic(err)
	}

	fmt.Printf("UserID: %d\n", userID)
	fmt.Printf("Token: %s\n", token.Token)
	fmt.Printf("ExpiresAt: %s", token.ExpiresAt.Format(time.RFC3339))
}
