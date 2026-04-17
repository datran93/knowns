package routes

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/howznguyen/knowns/internal/models"
	"github.com/howznguyen/knowns/internal/search"
	"github.com/howznguyen/knowns/internal/storage"
	"github.com/howznguyen/knowns/internal/workingmemory"
)

// MemoryRoutes handles persistent memories and session-scoped working memory endpoints.
type MemoryRoutes struct {
	store   *storage.Store
	mgr     *storage.Manager
	sse     Broadcaster
	working *workingmemory.Store
}

func (mr *MemoryRoutes) getStore() *storage.Store {
	if mr.mgr != nil {
		return mr.mgr.GetStore()
	}
	return mr.store
}

// Register wires the memory routes onto r.
func (mr *MemoryRoutes) Register(r chi.Router) {
	r.Get("/memories", mr.list)
	r.Post("/memories", mr.create)
	r.Get("/memories/{id}", mr.get)
	r.Put("/memories/{id}", mr.update)
	r.Delete("/memories/{id}", mr.delete)
	r.Post("/memories/{id}/promote", mr.promote)
	r.Post("/memories/{id}/demote", mr.demote)

	r.Get("/working-memories", mr.listWorking)
	r.Post("/working-memories", mr.createWorking)
	r.Get("/working-memories/{id}", mr.getWorking)
	r.Delete("/working-memories/{id}", mr.deleteWorking)
	r.Post("/working-memories/clean", mr.cleanWorking)
}

func (mr *MemoryRoutes) list(w http.ResponseWriter, r *http.Request) {
	layer := r.URL.Query().Get("layer")
	category := r.URL.Query().Get("category")
	tag := r.URL.Query().Get("tag")

	if layer != "" && !models.ValidPersistentMemoryLayer(layer) {
		respondError(w, http.StatusBadRequest, "invalid layer: must be project or global")
		return
	}

	entries, err := mr.getStore().Memory.ListPersistent(layer)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if category != "" {
		filtered := entries[:0]
		for _, e := range entries {
			if e.Category == category {
				filtered = append(filtered, e)
			}
		}
		entries = filtered
	}
	if tag != "" {
		filtered := entries[:0]
		for _, e := range entries {
			for _, t := range e.Tags {
				if t == tag {
					filtered = append(filtered, e)
					break
				}
			}
		}
		entries = filtered
	}

	if entries == nil {
		entries = []*models.MemoryEntry{}
	}

	respondJSON(w, http.StatusOK, entries)
}

func (mr *MemoryRoutes) get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	entry, err := mr.getStore().Memory.Get(id)
	if err != nil || !models.ValidPersistentMemoryLayer(entry.Layer) {
		respondError(w, http.StatusNotFound, "memory not found")
		return
	}
	respondJSON(w, http.StatusOK, entry)
}

type createMemoryRequest struct {
	Title    string            `json:"title"`
	Content  string            `json:"content"`
	Layer    string            `json:"layer"`
	Category string            `json:"category"`
	Tags     []string          `json:"tags"`
	Metadata map[string]string `json:"metadata"`
}

func (mr *MemoryRoutes) create(w http.ResponseWriter, r *http.Request) {
	var req createMemoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Layer == "" {
		req.Layer = models.MemoryLayerProject
	}
	if !models.ValidPersistentMemoryLayer(req.Layer) {
		respondError(w, http.StatusBadRequest, "invalid layer: must be project or global")
		return
	}

	now := time.Now().UTC()
	entry := &models.MemoryEntry{
		Title:     req.Title,
		Content:   req.Content,
		Layer:     req.Layer,
		Category:  req.Category,
		Tags:      req.Tags,
		Metadata:  req.Metadata,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if entry.Tags == nil {
		entry.Tags = []string{}
	}

	if err := mr.getStore().Memory.Create(entry); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	search.BestEffortIndexMemory(mr.getStore(), entry.ID)
	respondJSON(w, http.StatusCreated, entry)
}

type updateMemoryRequest struct {
	Title    *string  `json:"title"`
	Content  *string  `json:"content"`
	Category *string  `json:"category"`
	Tags     []string `json:"tags"`
}

func (mr *MemoryRoutes) update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	entry, err := mr.getStore().Memory.Get(id)
	if err != nil || !models.ValidPersistentMemoryLayer(entry.Layer) {
		respondError(w, http.StatusNotFound, "memory not found")
		return
	}

	var req updateMemoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Title != nil {
		entry.Title = *req.Title
	}
	if req.Content != nil {
		entry.Content = *req.Content
	}
	if req.Category != nil {
		entry.Category = *req.Category
	}
	if req.Tags != nil {
		entry.Tags = req.Tags
	}

	entry.UpdatedAt = time.Now().UTC()

	if err := mr.getStore().Memory.Update(entry); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	search.BestEffortIndexMemory(mr.getStore(), entry.ID)
	respondJSON(w, http.StatusOK, entry)
}

func (mr *MemoryRoutes) delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	entry, err := mr.getStore().Memory.Get(id)
	if err != nil || !models.ValidPersistentMemoryLayer(entry.Layer) {
		respondError(w, http.StatusNotFound, "memory not found")
		return
	}

	if err := mr.getStore().Memory.Delete(id); err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}

	search.BestEffortRemoveMemory(mr.getStore(), id)
	respondJSON(w, http.StatusOK, map[string]any{"deleted": true, "id": id})
}

func (mr *MemoryRoutes) promote(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	entry, err := mr.getStore().Memory.PromotePersistent(id)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	search.BestEffortIndexMemory(mr.getStore(), entry.ID)
	respondJSON(w, http.StatusOK, entry)
}

func (mr *MemoryRoutes) demote(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	entry, err := mr.getStore().Memory.DemotePersistent(id)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	search.BestEffortIndexMemory(mr.getStore(), entry.ID)
	respondJSON(w, http.StatusOK, entry)
}

func (mr *MemoryRoutes) listWorking(w http.ResponseWriter, r *http.Request) {
	entries := mr.working.List()
	if entries == nil {
		entries = []*models.MemoryEntry{}
	}
	respondJSON(w, http.StatusOK, entries)
}

func (mr *MemoryRoutes) createWorking(w http.ResponseWriter, r *http.Request) {
	var req createMemoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	entry := mr.working.Add(&models.MemoryEntry{
		Title:    req.Title,
		Content:  req.Content,
		Layer:    models.MemoryLayerWorking,
		Category: req.Category,
		Tags:     req.Tags,
		Metadata: req.Metadata,
	})
	respondJSON(w, http.StatusCreated, entry)
}

func (mr *MemoryRoutes) getWorking(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	entry, ok := mr.working.Get(id)
	if !ok {
		respondError(w, http.StatusNotFound, "working memory not found")
		return
	}
	respondJSON(w, http.StatusOK, entry)
}

func (mr *MemoryRoutes) deleteWorking(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if !mr.working.Delete(id) {
		respondError(w, http.StatusNotFound, "working memory not found")
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{"deleted": true, "id": id})
}

func (mr *MemoryRoutes) cleanWorking(w http.ResponseWriter, r *http.Request) {
	count := mr.working.Clear()
	respondJSON(w, http.StatusOK, map[string]any{"cleaned": count})
}
