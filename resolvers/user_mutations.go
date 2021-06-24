package resolvers

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sns"
)

type UserAssignDeviceTokenInput struct {
	DeviceToken string
}

func (r *Resolver) UserAssignDeviceToken(ctx context.Context, args struct {
	Input UserAssignDeviceTokenInput
}) (*bool, error) {
	currentUserID, err := ctxUserID(ctx)
	if err != nil {
		return nil, err
	}

	inputDeviceToken := args.Input.DeviceToken

	output, err := r.server.SNS.CreatePlatformEndpoint(&sns.CreatePlatformEndpointInput{
		Token:                  aws.String(inputDeviceToken),
		PlatformApplicationArn: aws.String(r.server.SNSPlatformApplicationArn),
	})
	if err != nil {
		return nil, err
	}

	_, err = r.server.ConnPool.Exec(`
		insert into user_device_tokens (user_id, device_token, endpoint_arn, enabled, created_at, updated_at)
		values ($1, $2, $3, true, now(), now())
	`, currentUserID, inputDeviceToken, output.EndpointArn)
	if err != nil {
		return nil, err
	}

	return nil, nil
}
