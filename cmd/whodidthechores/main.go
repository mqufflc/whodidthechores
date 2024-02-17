package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/mqufflc/whodidthechores/internal/api"
	"github.com/mqufflc/whodidthechores/internal/repository"
)

func main() {
	service, err := repository.NewService("postgres://postgres:example@localhost:5432/whodidthechores?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	err = service.Migrate("file://migrations")
	if err != nil {
		log.Fatal(err)
	}
	defaultHandler := api.New(service)

	http := &http.Server{
		Addr:    ":3000",
		Handler: defaultHandler,
	}

	fmt.Println("Listening on :3000")
	http.ListenAndServe()
}
