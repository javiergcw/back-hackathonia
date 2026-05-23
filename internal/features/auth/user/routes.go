package user

import (
	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

func RegisterRoutes(r *mux.Router, db *gorm.DB) {
	h := NewHandler(db)

	r.HandleFunc("/users", h.List).Methods("GET")
	r.HandleFunc("/users", h.Create).Methods("POST")
	r.HandleFunc("/users/{id}", h.GetByID).Methods("GET")
}
