package resolvers

import (
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Requires AWS_* environment variables
func TestRequestMediaUpload(t *testing.T) {
	t.Parallel()

	request := func(h *Harness, mediaType MediaType) (string, string) {
		t := h.t

		var res map[string]interface{}
		h.MustExec(ExecInput{
			Query: `
			mutation RequestMediaUpload() {
				requestMediaUpload(input: {
					mediaType: IMAGE
				}) {
					putURL
					getURL
				}
			}`,
		}, &res)

		res = res["requestMediaUpload"].(map[string]interface{})
		putURL := res["putURL"].(string)
		require.NotEmpty(t, putURL)

		getURL := res["getURL"].(string)
		require.NotEmpty(t, getURL)

		return putURL, getURL
	}

	testUpload := func(h *Harness, putURL, getURL string) {
		t := h.t
		const data = "dummy data"

		reader := strings.NewReader(data)
		uploadReq, err := http.NewRequest("PUT", putURL, reader)
		require.NoError(t, err)

		uploadResp, err := http.DefaultClient.Do(uploadReq)
		require.NoError(t, err)

		uploadBody, err := ioutil.ReadAll(uploadResp.Body)
		require.NoError(t, err)
		require.Equal(t, 200, uploadResp.StatusCode, string(uploadBody))

		getResp, err := http.Get(getURL)
		require.NoError(t, err)
		require.Equal(t, 200, getResp.StatusCode)

		b, err := ioutil.ReadAll(getResp.Body)
		require.NoError(t, err)

		assert.Equal(t, data, string(b))
	}

	for _, mediaType := range []MediaType{MediaTypeImage, MediaTypeVideo} {
		testName := strings.ToLower(string(mediaType))
		t.Run(testName, func(t *testing.T) {
			harness := NewTestHarness(t)
			putURL, getURL := request(harness, MediaTypeImage)
			testUpload(harness, putURL, getURL)
		})
	}
}
