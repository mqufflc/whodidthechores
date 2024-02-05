package main

import (
	"log"

	"github.com/mqufflc/whodidthechores/internal/repository"
)

func main() {
	service, err := repository.NewService("postgres://postgres:example@localhost:5432/whodidthechores?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	err = service.Migrate()
	if err != nil {
		log.Fatal(err)
	}
}
