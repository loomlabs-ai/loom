package providers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/loomlabs-ai/loom/internal/config"
	"github.com/loomlabs-ai/loom/internal/server/models"
	"github.com/loomlabs-ai/loom/internal/server/store"
	"modernc.org/sqlite"
)

func ConnectProvider(
	w http.ResponseWriter,
	r *http.Request,
) {
	w.Header().Set("Content-Type", "application/json")

	var req models.ConnectProviderReq

	err := json.NewDecoder(r.Body).Decode(&req)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(map[string]any{
			"error":   "invalid request",
			"message": err.Error(),
		})

		return
	}

	if req.ProviderName == "" {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "name is required",
		})

		return
	}

	if req.APIKey == "" {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "api_key is required",
		})

		return
	}

	switch strings.ToLower(req.ProviderName) {
	case "groq":
		req.BaseURL = config.GroqBaseURL
	case "openrouter":
		req.BaseURL = config.OpenRouterBaseURL
	}

	_, err = store.DB.Exec(
		`
		INSERT INTO providers (name, base_url, api_key)
		VALUES (?,?,?)
		`,
		strings.ToLower(req.ProviderName),
		req.BaseURL,
		req.APIKey,
	)

	if err != nil {
		var sqliteErr *sqlite.Error

		if errors.As(err, &sqliteErr) {
			if sqliteErr.Code() == 2067 {
				w.WriteHeader(409)

				json.NewEncoder(w).Encode(map[string]any{
					"error": "The provider is already connected.",
				})

				return
			}
		}

		w.WriteHeader(http.StatusInternalServerError)

		json.NewEncoder(w).Encode(map[string]any{
			"error":   "database error",
			"message": err.Error(),
		})

		return
	}

	w.WriteHeader(201)

	json.NewEncoder(w).Encode(map[string]any{
		"message": "provider connected",
	})
}

func DeleteProvider(
	w http.ResponseWriter,
	r *http.Request,
) {
	w.Header().Set("Content-Type", "application/json")

	name := strings.ToLower(chi.URLParam(r, "name"))

	if name == "" {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "name is required",
		})

		return
	}

	result, err := store.DB.Exec(
		`
		DELETE FROM providers
		WHERE name = ?
		`,
		name,
	)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		json.NewEncoder(w).Encode(map[string]any{
			"error":   "database error",
			"message": err.Error(),
		})

		return
	}

	deleted, err := result.RowsAffected()

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		json.NewEncoder(w).Encode(map[string]any{
			"error":   "database error",
			"message": err.Error(),
		})

		return
	}

	if deleted == 0 {
		w.WriteHeader(http.StatusNotFound)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "provider not found",
		})

		return
	}

	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(map[string]any{
		"message": "provider deleted",
	})
}
