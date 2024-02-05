package api

import "github.com/mqufflc/whodidthechores/internal/repository"

type Server struct {
	HTTPAddress string
	Repository  *repository.Service
}
