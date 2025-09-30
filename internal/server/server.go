package server

import (
	"net/http"

	"github.com/ryanchen01/claude-code-router-go/internal/api"
	"github.com/ryanchen01/claude-code-router-go/internal/handler"
)

type handlers struct {
	handler.MessagesHandler
}

func newHandler() *handlers {
	return &handlers{
		MessagesHandler: *handler.NewMessagesHandler(),
	}
}

func Start() {
	apiHandler := api.Handler(newHandler())
	http.ListenAndServe(":8000", apiHandler)
}
