package projects

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/loomlabs-ai/loom/internal/server/models"
	"github.com/loomlabs-ai/loom/internal/server/store"
	"modernc.org/sqlite"
)

// CreateProject registers a folder on disk as a project so chats can be
// created on top of it.
func CreateProject(
	w http.ResponseWriter,
	r *http.Request,
) {
	w.Header().Set("Content-Type", "application/json")

	var req models.CreateProjectReq

	err := json.NewDecoder(r.Body).Decode(&req)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(map[string]any{
			"error":   "invalid request",
			"message": err.Error(),
		})

		return
	}

	req.Path = strings.TrimSpace(req.Path)

	if req.Path == "" {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "path is required",
		})

		return
	}

	path, err := filepath.Abs(req.Path)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(map[string]any{
			"error":   "invalid path",
			"message": err.Error(),
		})

		return
	}

	// Resolve symlinks so the same folder is not registered twice
	// under different paths.
	resolved, err := filepath.EvalSymlinks(path)

	if err == nil {
		path = resolved
	}

	info, err := os.Stat(path)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(map[string]any{
			"error":   "folder not found",
			"message": err.Error(),
		})

		return
	}

	if !info.IsDir() {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "path is not a folder",
		})

		return
	}

	name := strings.TrimSpace(req.Name)

	if name == "" {
		name = filepath.Base(path)
	}

	result, err := store.DB.Exec(
		`
		INSERT INTO projects (name, path)
		VALUES (?,?)
		`,
		name,
		path,
	)

	if err != nil {
		var sqliteErr *sqlite.Error

		if errors.As(err, &sqliteErr) {
			if sqliteErr.Code() == 2067 {
				w.WriteHeader(http.StatusConflict)

				json.NewEncoder(w).Encode(map[string]any{
					"error": "The folder is already added as a project.",
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

	id, err := result.LastInsertId()

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		json.NewEncoder(w).Encode(map[string]any{
			"error":   "database error",
			"message": err.Error(),
		})

		return
	}

	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(map[string]any{
		"message": "project created",
		"project": map[string]any{
			"id":   id,
			"name": name,
			"path": path,
		},
	})
}

// ListProjects returns every registered project, newest first.
func ListProjects(
	w http.ResponseWriter,
	r *http.Request,
) {
	w.Header().Set("Content-Type", "application/json")

	rows, err := store.DB.Query(
		`
		SELECT id, name, path, created_at
		FROM projects
		ORDER BY created_at DESC, id DESC
		`,
	)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		json.NewEncoder(w).Encode(map[string]any{
			"error":   "database error",
			"message": err.Error(),
		})

		return
	}

	defer rows.Close()

	projects := []map[string]any{}

	for rows.Next() {
		var (
			id        int64
			name      string
			path      string
			createdAt string
		)

		err := rows.Scan(&id, &name, &path, &createdAt)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)

			json.NewEncoder(w).Encode(map[string]any{
				"error":   "database error",
				"message": err.Error(),
			})

			return
		}

		projects = append(projects, map[string]any{
			"id":         id,
			"name":       name,
			"path":       path,
			"created_at": createdAt,
		})
	}

	if err := rows.Err(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		json.NewEncoder(w).Encode(map[string]any{
			"error":   "database error",
			"message": err.Error(),
		})

		return
	}

	json.NewEncoder(w).Encode(map[string]any{
		"projects": projects,
	})
}

// idParam reads the {id} path parameter and writes an error response when it
// is not a valid project id. It reports whether the request may continue.
func idParam(
	w http.ResponseWriter,
	r *http.Request,
) (int64, bool) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)

	if err != nil || id <= 0 {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "invalid project id",
		})

		return 0, false
	}

	return id, true
}

func GetProject(
	w http.ResponseWriter,
	r *http.Request,
) {
	w.Header().Set("Content-Type", "application/json")

	id, ok := idParam(w, r)

	if !ok {
		return
	}

	var (
		name      string
		path      string
		createdAt string
	)

	err := store.DB.QueryRow(
		`
		SELECT name, path, created_at
		FROM projects
		WHERE id = ?
		`,
		id,
	).Scan(&name, &path, &createdAt)

	if errors.Is(err, sql.ErrNoRows) {
		w.WriteHeader(http.StatusNotFound)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "project not found",
		})

		return
	}

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		json.NewEncoder(w).Encode(map[string]any{
			"error":   "database error",
			"message": err.Error(),
		})

		return
	}

	json.NewEncoder(w).Encode(map[string]any{
		"project": map[string]any{
			"id":         id,
			"name":       name,
			"path":       path,
			"created_at": createdAt,
		},
	})
}

// DeleteProject removes the project registration. The folder itself and the
// files inside it are left untouched on disk.
func DeleteProject(
	w http.ResponseWriter,
	r *http.Request,
) {
	w.Header().Set("Content-Type", "application/json")

	id, ok := idParam(w, r)

	if !ok {
		return
	}

	result, err := store.DB.Exec(
		`
		DELETE FROM projects
		WHERE id = ?
		`,
		id,
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
			"error": "project not found",
		})

		return
	}

	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(map[string]any{
		"message": "project deleted",
	})
}
