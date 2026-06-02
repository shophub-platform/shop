package handler

import (
	"net/http"
	"time"

	"github.com/shophub/shop/pkg/response"
)

// HealthResponse je struktura odgovora za health check endpoint.
type HealthResponse struct {
	Status    string `json:"status"`
	Service   string `json:"service"`
	Timestamp string `json:"timestamp"`
}

// HealthHandler drži zavisnosti za health handler.
// U Spring-u bi ovo bio @RestController sa @Autowired zavisnostima.
type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// Health je handler za GET /health.
// U Spring-u bi ovo bio @GetMapping("/health") metod.
func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	response.Ok(w, HealthResponse{
		Status:    "ok",
		Service:   "shop",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	})
}
