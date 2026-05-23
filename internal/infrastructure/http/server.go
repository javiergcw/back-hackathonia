package http

import (
	"fmt"
	"net/http"

	"github.com/javierg/hackathon-bqia/internal/infrastructure/config"
	"gorm.io/gorm"
)

type Server struct {
	httpServer *http.Server
}

func NewServer(cfg *config.Config, db *gorm.DB) *Server {
	router := NewRouter(cfg, db)

	return &Server{
		httpServer: &http.Server{
			Addr:    fmt.Sprintf(":%s", cfg.Port),
			Handler: router,
		},
	}
}

func (s *Server) ListenAndServe() error {
	return s.httpServer.ListenAndServe()
}
