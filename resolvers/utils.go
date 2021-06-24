package resolvers

import (
	"context"
	"errors"

	"github.com/ttacon/libphonenumber"
	sq "gopkg.in/Masterminds/squirrel.v1"
)

type scannable interface {
	Scan(...interface{}) error
}

func newSelectBuilder(columns ...string) sq.SelectBuilder {
	return sq.Select(columns...).PlaceholderFormat(sq.Dollar)
}

func newInsertBuilder(table string) sq.InsertBuilder {
	return sq.Insert(table).PlaceholderFormat(sq.Dollar)
}

func parsePhoneNumber(phoneNumber string) (string, error) {
	number, err := libphonenumber.Parse(phoneNumber, "US")
	if err != nil {
		return "", err
	}

	return libphonenumber.Format(number, libphonenumber.E164), nil
}

func ctxUserID(ctx context.Context) (int64, error) {
	// use typed key
	// for test
	// isTest := true
	// if !isTest {
	if userID, ok := ctx.Value("user_id").(int64); ok {
		return userID, nil
	}

	return 0, errors.New("request context has no userID")
	// } else {
	// 	return 744, nil
	// }

}
