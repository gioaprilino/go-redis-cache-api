package handler

import (
	"encoding/json"
	"net/http"

	"go-redis-cache-api/internal/model"
	"go-redis-cache-api/internal/service"

	"github.com/go-chi/chi/v5"
)

type ProductHandler struct {
	svc *service.ProductService
}

func NewProductHandler(svc *service.ProductService) *ProductHandler {
	return &ProductHandler{svc: svc}
}

func respond(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *ProductHandler) GetAllProducts(w http.ResponseWriter, r *http.Request) {
	resp, err := h.svc.GetAllProducts(r.Context())
	if err != nil {
		respond(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	respond(w, http.StatusOK, resp)
}

func (h *ProductHandler) GetProductByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	resp, err := h.svc.GetProductByID(r.Context(), id)
	if err != nil {
		respond(w, http.StatusNotFound, map[string]string{"error": "Product not found"})
		return
	}
	respond(w, http.StatusOK, resp)
}

func (h *ProductHandler) GetProductNoCache(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	resp, err := h.svc.GetProductNoCache(r.Context(), id)
	if err != nil {
		respond(w, http.StatusNotFound, map[string]string{"error": "Product not found"})
		return
	}
	respond(w, http.StatusOK, resp)
}

func (h *ProductHandler) CreateProductWriteThrough(w http.ResponseWriter, r *http.Request) {
	var req model.CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}
	resp, err := h.svc.CreateProductWriteThrough(r.Context(), req)
	if err != nil {
		respond(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	respond(w, http.StatusCreated, resp)
}

func (h *ProductHandler) CreateProductWriteBack(w http.ResponseWriter, r *http.Request) {
	var req model.CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}
	resp, err := h.svc.CreateProductWriteBack(r.Context(), req)
	if err != nil {
		respond(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	respond(w, http.StatusAccepted, resp)
}

func (h *ProductHandler) UpdateProductWriteAround(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req model.UpdateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}
	resp, err := h.svc.UpdateProductWriteAround(r.Context(), id, req)
	if err != nil {
		respond(w, http.StatusNotFound, map[string]string{"error": "Product not found"})
		return
	}
	respond(w, http.StatusOK, resp)
}

func (h *ProductHandler) Benchmark(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	results, err := h.svc.Benchmark(r.Context(), id)
	if err != nil {
		respond(w, http.StatusNotFound, map[string]string{"error": "Product not found"})
		return
	}
	respond(w, http.StatusOK, map[string]interface{}{
		"data":     results,
		"message":  "Latency comparison: Redis (RAM ~100ns) vs Database (disk ~10ms)",
	})
}

func (h *ProductHandler) GetSession(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "sessionId")
	resp, err := h.svc.GetSession(r.Context(), sessionID)
	if err != nil {
		respond(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	respond(w, http.StatusOK, resp)
}

func (h *ProductHandler) GetCacheStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.svc.GetCacheStats(r.Context())
	if err != nil {
		respond(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	respond(w, http.StatusOK, stats)
}
