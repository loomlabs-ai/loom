package server

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/loomlabs-ai/loom/internal/server/handlers/projects"
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

		// Create new project handler
		r.Post(
			"/projects", projects.CreateProject,
		)

		// List projects handler
		r.Get(
			"/projects", projects.ListProjects,
		)

		// Get project handler
		r.Get(
			"/projects/{id}", projects.GetProject,
		)

		// Delete project handler
		r.Delete(
			"/projects/{id}", projects.DeleteProject,
		)
	})

	http.ListenAndServe(port, router)
}
