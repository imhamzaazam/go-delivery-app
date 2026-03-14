package presentation

import "github.com/horiondreher/go-web-api-boilerplate/internal/core"

type Handler struct {
	shared *core.Shared
}

func New(shared *core.Shared) *Handler {
	return &Handler{shared: shared}
}
