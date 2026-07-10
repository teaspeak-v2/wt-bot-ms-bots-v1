package handlers

import "net/http"

type HealthHandler struct {
	dbPing    func() error
	redisPing func() error
}

func NewHealthHandler(dbPing func() error, redisPing func() error) *HealthHandler {
	return &HealthHandler{dbPing: dbPing, redisPing: redisPing}
}

func (h *HealthHandler) Root(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// Healthz godoc
// @Summary Liveness probe
// @Tags health
// @Produce json
// @Success 200 {object} map[string]string
// @Router /healthz [get]
func (h *HealthHandler) Healthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// Readyz godoc
// @Summary Readiness probe
// @Tags health
// @Produce json
// @Success 200 {object} map[string]string
// @Router /readyz [get]
func (h *HealthHandler) Readyz(w http.ResponseWriter, r *http.Request) {
	if h.dbPing != nil {
		if err := h.dbPing(); err != nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "db down"})
			return
		}
	}
	if h.redisPing != nil {
		if err := h.redisPing(); err != nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "redis down"})
			return
		}
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}
