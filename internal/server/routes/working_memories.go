package routes

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/datran93/knowns/internal/models"
	"github.com/datran93/knowns/internal/storage"
	"github.com/go-chi/chi/v5"
)

// WorkingMemoryRoutes handles working (session) memory endpoints.
type WorkingMemoryRoutes struct {
	store *storage.Store
	mgr   *storage.Manager
}

func (mr *WorkingMemoryRoutes) getStore() *storage.Store {
	if mr.mgr != nil {
		return mr.mgr.GetStore()
	}
	return mr.store
}

// Register wires the working memory routes onto r.
func (mr *WorkingMemoryRoutes) Register(r chi.Router) {
	r.Get("/working-memories", mr.list)
	r.Post("/working-memories", mr.create)
	r.Get("/working-memories/{id}", mr.get)
	r.Delete("/working-memories/{id}", mr.delete)
	r.Post("/working-memories/clean", mr.clean)
}

func (mr *WorkingMemoryRoutes) list(w http.ResponseWriter, r *http.Request) {
	entries, err := mr.getStore().Memory.ListWorking()
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, entries)
}

func (mr *WorkingMemoryRoutes) get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	entry, err := mr.getStore().Memory.GetWorking(id)
	if err != nil {
		respondError(w, http.StatusNotFound, "working memory not found")
		return
	}
	respondJSON(w, http.StatusOK, entry)
}

type createWorkingMemoryRequest struct {
	Title    string            `json:"title"`
	Content  string            `json:"content"`
	Category string            `json:"category"`
	Tags     []string          `json:"tags"`
	Metadata map[string]string `json:"metadata"`
}

func (mr *WorkingMemoryRoutes) create(w http.ResponseWriter, r *http.Request) {
	var req createWorkingMemoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	now := time.Now().UTC()
	entry := &models.MemoryEntry{
		Title:     req.Title,
		Content:   req.Content,
		Layer:     models.MemoryLayerWorking,
		Category:  req.Category,
		Tags:      req.Tags,
		Metadata:  req.Metadata,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if entry.Tags == nil {
		entry.Tags = []string{}
	}

	if err := mr.getStore().Memory.CreateWorking(entry); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, entry)
}

func (mr *WorkingMemoryRoutes) delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := mr.getStore().Memory.DeleteWorking(id); err != nil {
		respondError(w, http.StatusNotFound, "working memory not found")
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{"deleted": true, "id": id})
}

func (mr *WorkingMemoryRoutes) clean(w http.ResponseWriter, r *http.Request) {
	count, err := mr.getStore().Memory.CleanWorking()
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{"cleaned": count})
}
