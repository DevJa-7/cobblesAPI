package resolvers

import (
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/lambdacollective/cobbles-api/server"
)

type PostMediaResolver struct {
	server *server.Server

	postMedia *server.PostMedia
	post      *server.Post
	isImage   bool
}

func (r *PostMediaResolver) URL() *string {
	u, err := url.Parse(r.postMedia.URL)
	if err != nil {
		log.Println(err)
		return nil
	}

	var bucket string
	hostParts := strings.Split(u.Host, ".")
	if len(hostParts) > 0 {
		bucket = hostParts[0]
	}

	// TODO: wrangle messy media asset kinds
	mediaURL := r.postMedia.URL
	switch {
	case bucket == "llc-cobbles-dev-user-media":
		mediaURL = fmt.Sprintf("%s%s", r.server.ImgixUserMediaMediaEndpoint, u.Path)
	case bucket == "llc-cobbles-dev-processed-user-media" && r.isImage:
		mediaURL = fmt.Sprintf("%s%s", r.server.ImgixProcessedMediaEndpoint, u.Path)
	default:
	}

	return &mediaURL
}

func (r *PostMediaResolver) Width() *int32 {
	return &r.postMedia.Width

}

func (r *PostMediaResolver) Height() *int32 {
	return &r.postMedia.Height
}
