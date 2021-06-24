package main

import (
	"log"

	"github.com/lambdacollective/cobbles-api/server"
)

func main() {
	s := server.NewServer()

	log.Fatalln(s.ProcessMediaQueue())
}
