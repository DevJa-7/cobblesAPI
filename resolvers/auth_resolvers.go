package resolvers

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"

	"github.com/jackc/pgx"
	"github.com/lambdacollective/cobbles-api/server"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sns"
)

type LoginUserResultResolver struct {
	server *server.Server

	token *server.AuthToken
}

func (r *LoginUserResultResolver) Token() string {
	return r.token.Token
}

func (r *LoginUserResultResolver) ExpiresAt() Timestamp {
	return Timestamp{r.token.ExpiresAt}
}

// LoginUser uses a login code + phone number to produce an authentication JWT
// of the authenticated user. A new user is created if none is found for the
// given phone number.
func (r *Resolver) LoginUser(ctx context.Context, args struct {
	Input struct {
		PhoneNumber string
		LoginCode   string
		FCMToken    *string
	}
}) (*LoginUserResultResolver, error) {
	inputPhoneNumber := args.Input.PhoneNumber
	inputLoginCode := args.Input.LoginCode
	inputFCMToken := args.Input.FCMToken
	fmt.Println("===inputFCMToken (not pointer) ===", inputFCMToken)
	// fmt.Println("===inputFCMToken (pointer) ===", *inputFCMToken)
	phoneNumber, err := parsePhoneNumber(inputPhoneNumber)
	if err != nil {
		return nil, err
	}

	var loginCodeResult struct {
		userID *int64
	}
	err = r.server.ConnPool.QueryRow(`
		select
			users.id as user_id
		from login_codes
		left join users on users.phone_number = login_codes.phone_number
		where login_codes.phone_number = $1
			and login_codes.login_code = $2
			and login_codes.expires_at >= now()
			and enabled is true
		order by login_codes.created_at desc
		limit 1
	`, phoneNumber, inputLoginCode).Scan(&loginCodeResult.userID)
	switch {
	case err == pgx.ErrNoRows:
		return nil, errors.New("invalid login")
	case err != nil:
		return nil, err
	}

	// TODO: wrap this and previous query in tx
	_, err = r.server.ConnPool.Exec(`
		update login_codes 
		set enabled = false
		where phone_number = $1
	`, phoneNumber)
	if err != nil {
		return nil, err
	}

	if loginCodeResult.userID != nil {
		// TODO: update the fcm_token in user table.
		_, err = r.server.ConnPool.Exec(`
			update users 
			set fcm_token = $1
			where id = $2
		`, inputFCMToken, *loginCodeResult.userID)
		if err != nil {
			return nil, err
		}

		authToken, err := r.server.GenerateAuthJWT(*loginCodeResult.userID)
		if err != nil {
			return nil, err
		}

		return &LoginUserResultResolver{
			token: authToken,
		}, nil
	}

	var newUserResult struct {
		userID int64
	}
	err = r.server.ConnPool.QueryRow(`
		insert into users (phone_number, fcm_token, created_at, updated_at)
		values ($1, $2, now(), now())
		on conflict (phone_number) do nothing
		returning users.id
	`, phoneNumber, inputFCMToken).Scan(&newUserResult.userID)
	if err != nil {
		return nil, err
	}

	authToken, err := r.server.GenerateAuthJWT(newUserResult.userID)
	if err != nil {
		return nil, err
	}
	fmt.Println("=============authToken user success==========", authToken)
	return &LoginUserResultResolver{
		token: authToken,
	}, nil
}

// RequestLoginCode produces a 6-digit login code that may be used in
// conjunction with a phone number in LoginUser to login a user.
//
// Enabled login codes are naively generated given there there has been no new
// code in the checked duration serving as a rate limit, although LoginUser
// only checks the most recently generated one.
func (r *Resolver) RequestLoginCode(ctx context.Context, args struct {
	Input struct {
		PhoneNumber string
	}
}) (*bool, error) {
	inputPhoneNumber := args.Input.PhoneNumber

	phoneNumber, err := parsePhoneNumber(inputPhoneNumber)
	if err != nil {
		return nil, err
	}

	var _id int64
	err = r.server.ConnPool.QueryRow(`
		select id from login_codes
		where phone_number = $1
			and created_at >= now() - interval '15' second
			and enabled is true
	`, phoneNumber).Scan(&_id)
	switch {
	case err == pgx.ErrNoRows:
		break
	case err != nil:
		return nil, err
	default:
		return nil, errors.New("you're doing that too quick! please try again later")
	}

	randInt, err := rand.Int(rand.Reader, big.NewInt(999999))
	if err != nil {
		return nil, err
	}
	loginCode := fmt.Sprintf("%06d", randInt.Int64())

	_, err = r.server.ConnPool.Exec(`
		insert into login_codes (phone_number, login_code, expires_at, enabled, created_at)
		values ($1, $2, now() + interval '15' minute, true, now())
	`, phoneNumber, loginCode)
	if err != nil {
		return nil, err
	}

	message := fmt.Sprintf("Your login code is %s", loginCode)

	_, err = r.server.SNS.Publish(&sns.PublishInput{
		Message:     aws.String(message),
		PhoneNumber: aws.String(phoneNumber),
	})
	if err != nil {
		return nil, err
	}

	return nil, nil
}
