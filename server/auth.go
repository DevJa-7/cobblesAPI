package server

import (
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

type AuthToken struct {
	Token     string
	ExpiresAt time.Time
}

type AuthJWTClaims struct {
	UserID int64
	jwt.StandardClaims
}

func (s *Server) GenerateAuthJWT(userID int64) (*AuthToken, error) {
	expiresAt := time.Now().Add(time.Hour * 24 * 365).UTC()

	claims := &AuthJWTClaims{
		UserID: userID,
		StandardClaims: jwt.StandardClaims{
			NotBefore: time.Now().Unix(),
			IssuedAt:  time.Now().Unix(),
			ExpiresAt: expiresAt.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign and get the complete encoded token as a string using the secret
	tokenStr, err := token.SignedString([]byte(s.ServerSecret))
	if err != nil {
		return nil, err
	}

	return &AuthToken{
		Token:     tokenStr,
		ExpiresAt: expiresAt,
	}, nil
}

func (s *Server) ValidateAuthJWT(tokenStr string) (*jwt.Token, error) {
	return jwt.ParseWithClaims(tokenStr, &AuthJWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.ServerSecret), nil
	})
}
