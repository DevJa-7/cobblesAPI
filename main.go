package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/graph-gophers/graphql-go/relay"
	"github.com/lambdacollective/cobbles-api/gqlschema"
	"github.com/lambdacollective/cobbles-api/resolvers"
	"github.com/lambdacollective/cobbles-api/server"
)

var (
	listenAddr = ":8080"
)

func main() {
	// connPool := mustSetupPostgres()
	server := server.NewServer()

	resolver := resolvers.NewResolver(server)

	schema := gqlschema.MustParseSchema(resolver)
	graphqlHandler := &relay.Handler{Schema: schema}

	http.Handle("/api-doc", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(page)
	}))
	http.Handle("/graphql", authMiddleware(server, graphqlHandler))

	envPort := os.Getenv("PORT")
	if envPort != "" {
		listenAddr = ":" + envPort
	}

	log.Println("listening on", listenAddr)
	log.Fatal(http.ListenAndServe(listenAddr, nil))
}

func authMiddleware(s *server.Server, handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("authorization")

		if authHeader != "" {
			//Ash Commented Start
			parts := strings.Split(authHeader, " ")
			// log.Print(len(parts))
			if !(len(parts) >= 2) {
				http.Error(w, "authorization header invalid", http.StatusBadRequest)
				return
			}
			token, err := s.ValidateAuthJWT(parts[1])
			// End

			// parts := strings.Split(authHeader, ".")
			// log.Print(parts[1])
			// log.Print(parts[0])
			// if (len(parts[1]) != 3) {
			// 	http.Error(w, "authorization header invalid 1", http.StatusBadRequest)
			// 	return
			// }
			// log.Print(parts[1])
			// //token, err := s.ValidateAuthJWT(authHeader)
			// token, err := s.ValidateAuthJWT(parts[1])
			if err != nil {
				log.Print(err)
				http.Error(w, "authorization header invalid", http.StatusBadRequest)
				return
			}

			claims, ok := token.Claims.(*server.AuthJWTClaims)
			if !ok {
				log.Print(err)
				return
			}
			authKey := "user_id"
			r = r.WithContext(context.WithValue(r.Context(), authKey, claims.UserID))
		}

		handler.ServeHTTP(w, r)
	})
}

var page = []byte(`
<!DOCTYPE html>
<html>
	<head>
		<link href="https://cdnjs.cloudflare.com/ajax/libs/graphiql/0.11.11/graphiql.min.css" rel="stylesheet" />
		<script src="https://cdnjs.cloudflare.com/ajax/libs/es6-promise/4.1.1/es6-promise.auto.min.js"></script>
		<script src="https://cdnjs.cloudflare.com/ajax/libs/fetch/2.0.3/fetch.min.js"></script>
		<script src="https://cdnjs.cloudflare.com/ajax/libs/react/16.2.0/umd/react.production.min.js"></script>
		<script src="https://cdnjs.cloudflare.com/ajax/libs/react-dom/16.2.0/umd/react-dom.production.min.js"></script>
		<script src="https://cdnjs.cloudflare.com/ajax/libs/graphiql/0.11.11/graphiql.min.js"></script>
	</head>
	<body style="width: 100%; height: 100%; margin: 0; overflow: hidden;">
		<div id="graphiql" style="height: 100vh;">Loading...</div>
		<script>
			function graphQLFetcher(graphQLParams) {
				return fetch("/graphql", {
					method: "post",
					body: JSON.stringify(graphQLParams),
					credentials: "include",
				}).then(function (response) {
					return response.text();
				}).then(function (responseBody) {
					try {
						return JSON.parse(responseBody);
					} catch (error) {
						return responseBody;
					}
				});
			}
			ReactDOM.render(
				React.createElement(GraphiQL, {fetcher: graphQLFetcher}),
				document.getElementById("graphiql")
			);
		</script>
	</body>
</html>
`)
