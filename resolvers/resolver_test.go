package resolvers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"testing"
	"text/template"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/davecgh/go-spew/spew"
	graphql "github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/errors"
	"github.com/jackc/pgx"
	"github.com/lambdacollective/cobbles-api/gqlschema"
	"github.com/lambdacollective/cobbles-api/server"
	snakecase "github.com/segmentio/go-snakecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var connPool *pgx.ConnPool

func init() {
	var connConfig pgx.ConnConfig
	var err error
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		connConfig = pgx.ConnConfig{
			User:     "postgres",
			Database: "test",
		}
	} else {
		connConfig, err = pgx.ParseConnectionString(databaseURL)
		if err != nil {
			panic(err)
		}
	}

	connPool, err = pgx.NewConnPool(pgx.ConnPoolConfig{
		ConnConfig:     connConfig,
		MaxConnections: 10,
	})
	if err != nil {
		panic(err)
	}
}

type Harness struct {
	t      *testing.T
	schema *graphql.Schema
	mutex  *sync.Mutex
}

func NewTestHarness(t *testing.T) *Harness {
	resolver := NewResolver(&server.Server{
		ConnPool:            connPool,
		S3UserMediaBucket:   "llc-cobbles-dev-user-media",
		S3ImageProxyBaseURL: "https://llc-cobbles-dev-user-images.imgix.net",
		S3:                  s3.New(session.New()),
		// ImgixProcessedMediaEndpoint: "https://processed-user-media.imgix.net",
		ImgixProcessedMediaEndpoint: "https://llc-cobbles-dev-processed-user-images.imgix.net",
		ImgixUserMediaMediaEndpoint: "https://llc-cobbles-dev-user-images.imgix.net",
	})

	return &Harness{
		t:      t,
		schema: gqlschema.MustParseSchema(resolver),
		mutex:  &sync.Mutex{},
	}
}

type GQLAssertInput struct {
	ExecInput ExecInput

	ExpectedResult interface{}
	ExpectedErrors []*errors.QueryError
}

func (h *Harness) GQLAssert(name string, in GQLAssertInput) {
	var actualResult map[string]interface{}
	errors := h.Exec(in.ExecInput, &actualResult)

	if in.ExpectedErrors != nil {
		for i, err := range errors {
			assert.Equal(h.t, in.ExpectedErrors[i].Message, err.Message)
		}
	} else {
		require.EqualValues(h.t, in.ExpectedErrors, errors, fmt.Sprintf("%s: errors", name))
	}

	switch specific := in.ExpectedResult.(type) {
	// assume json string
	case string:
		var expectedUnmarshaled map[string]interface{}
		err := json.Unmarshal([]byte(specific), &expectedUnmarshaled)
		require.NoError(h.t, err, "invalid expected json")

		if !assert.Equal(h.t, expectedUnmarshaled, actualResult) {
			fmt.Println(h.t.Name() + " ACTUAL")
			bytes, err := json.Marshal(actualResult)
			require.NoError(h.t, err)
			fmt.Println(string(bytes))
			spew.Dump(actualResult)

			fmt.Println()

			fmt.Println(h.t.Name() + " EXPECTED")
			bytes, err = json.Marshal(expectedUnmarshaled)
			require.NoError(h.t, err)
			fmt.Println(string(bytes))
			spew.Dump(expectedUnmarshaled)
		}
	default:
		untypedExpectedResult := h.JSONify(in.ExpectedResult)
		assert.Equal(h.t, untypedExpectedResult, actualResult, fmt.Sprintf("%s: result", name))
	}
}

func (h *Harness) ResetDB() {
	rows, err := connPool.Query(`
SELECT table_name FROM information_schema.tables 
WHERE table_schema = 'public'
`)
	require.NoError(h.t, err)
	defer rows.Close()

	for rows.Next() {
		var table string
		err = rows.Scan(&table)
		require.NoError(h.t, err)

		// neighborhoods migration has an INSERT
		if table == "neighborhoods" {
			continue
		}

		_, err := connPool.Exec(fmt.Sprintf("TRUNCATE %s CASCADE", table))
		require.NoError(h.t, err)
	}
}

type ExecInput struct {
	UserID    int64
	Query     string
	Variables map[string]interface{}
}

func (h *Harness) Exec(in ExecInput, to interface{}) []*errors.QueryError {
	ctx := context.WithValue(context.Background(), "user_id", in.UserID)
	resp := h.schema.Exec(ctx, in.Query, "", in.Variables)

	if len(resp.Errors) > 0 {
		return resp.Errors
	}

	err := json.Unmarshal(resp.Data, &to)
	require.NoError(h.t, err, "Query response failed to json.Ummarshal: %s", in.Query)
	return nil
}

func (h *Harness) MustExec(in ExecInput, to interface{}) {
	errors := h.Exec(in, to)
	require.Empty(h.t, errors, "Query returned errors: %s", in.Query)
}

// make whaetver lose all its types so you can assert
// it against real JSON
func (h *Harness) JSONify(v interface{}) (res map[string]interface{}) {
	b, err := json.Marshal(v)
	require.NoError(h.t, err)

	err = json.Unmarshal(b, &res)
	require.NoError(h.t, err)
	return
}

func (h *Harness) MustCreateUser(userID int64) {
	// HACK, should probably use some gql method... shrug
	_, err := connPool.Exec(`
INSERT INTO users (id, created_at, updated_at)
	VALUES ($1, now(), now())
ON CONFLICT (id)
	DO NOTHING
`, userID)
	require.NoError(h.t, err)
}

// TODO
func (h *Harness) NewUser() string {
	return "TODO: user ID"
}

func (h *Harness) TemplateString(templ string, args interface{}) string {
	tname := snakecase.Snakecase(h.t.Name())
	t := template.New(tname)
	parsed := template.Must(t.Delims("[[", "]]").Parse(templ))

	var b bytes.Buffer
	err := parsed.Execute(&b, args)
	require.NoError(h.t, err)

	return b.String()
}
