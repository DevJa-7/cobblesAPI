package resolvers

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	uuid "github.com/satori/go.uuid"
)

type MediaType string

const (
	MediaTypeImage  MediaType = "IMAGE"
	MediaTypeVideo  MediaType = "VIDEO"
	MediaTypeAvatar MediaType = "AVATAR"
)

type RequestMediaUploadResult struct {
	putURL string
	getURL string
}

func (r *Resolver) RequestMediaUpload(ctx context.Context, args struct {
	Input struct {
		MediaType MediaType
	}
}) (*RequestMediaUploadResult, error) {
	var prefix string
	switch args.Input.MediaType {
	case MediaTypeImage:
		prefix = "images"
	case MediaTypeVideo:
		prefix = "videos"
	case MediaTypeAvatar:
		prefix = "avatars"
	default:
		return nil, errors.New("invalid MediaType")
	}

	uniqueID := uuid.Must(uuid.NewV4()).String()
	key := fmt.Sprintf("%s/%s", prefix, uniqueID)

	req, _ := r.server.S3.PutObjectRequest(&s3.PutObjectInput{
		Bucket: aws.String(r.server.S3UserMediaBucket),
		Key:    aws.String(key),
	})
	putURL, err := req.Presign(15 * time.Minute)
	if err != nil {
		return nil, err
	}

	getURL := fmt.Sprintf("https://%s/%s", r.s3Hostname(), key)
	return &RequestMediaUploadResult{putURL: putURL, getURL: getURL}, nil
}

func (r *RequestMediaUploadResult) PutURL() string {
	return r.putURL
}

func (r *RequestMediaUploadResult) GetURL() string {
	return r.getURL
}

func (r *Resolver) s3Hostname() string {
	return fmt.Sprintf("%s.s3-external-1.amazonaws.com", r.server.S3UserMediaBucket)
}
