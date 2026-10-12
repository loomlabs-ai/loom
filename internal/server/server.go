package server

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/loomlabs-ai/loom/internal/server/handlers/providers"
)

func StartServer(port string) {
	router := chi.NewRouter()

	fmt.Printf("server starting on %s\n", port)

	router.Route("/api/v1", func(r chi.Router) {
		// Connect new provider handler
		r.Post(
			"/provider", providers.ConnectProvider,
		)

		// Delete provider handler
		r.Delete(
			"/provider/{name}", providers.DeleteProvider,
		)
	})

	http.ListenAndServe(port, router)
}
