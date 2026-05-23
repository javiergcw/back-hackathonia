package http

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/javierg/hackathon-bqia/internal/features/auth/user"
	"github.com/javierg/hackathon-bqia/internal/shared/response"
	"gorm.io/gorm"
)

func NewRouter(db *gorm.DB) *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		response.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}).Methods(http.MethodGet)

	user.RegisterRoutes(r.PathPrefix("/api/v1").Subrouter(), db)

	return r
}
