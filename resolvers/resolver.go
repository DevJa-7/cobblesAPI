package resolvers

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/lambdacollective/cobbles-api/server"
	validator "gopkg.in/go-playground/validator.v9"
)

var validate = validator.New()

type Resolver struct {
	server *server.Server
}

func NewResolver(s *server.Server) *Resolver {
	return &Resolver{
		server: s,
	}
}

func DecodeAfterIDCursor(pageToken *string) (int64, error) {
	if pageToken == nil {
		return -1, nil
	}

	// Indicate to client that value is opaque
	parts := strings.SplitN(*pageToken, "=", 2)
	key := parts[0]
	if key != "after" {
		return -1, fmt.Errorf("got '%s', expected 'after'", key)
	}

	value := parts[1]
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return -1, fmt.Errorf("%s is not an integer", value)
	}

	return id, nil
}

func EncodeAfterIDCursor(afterID int64) *string {
	if afterID == 0 {
		return nil
	}

	token := fmt.Sprintf("after=%d", afterID)
	return &token
}
