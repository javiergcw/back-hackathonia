package user

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	apperrors "github.com/javierg/hackathon-bqia/internal/shared/errors"
	"github.com/javierg/hackathon-bqia/internal/shared/response"
	"github.com/javierg/hackathon-bqia/internal/shared/validation"
	"gorm.io/gorm"
)

type Handler struct {
	repo *Repository
}

func NewHandler(db *gorm.DB) *Handler {
	return &Handler{repo: NewRepository(db)}
}

type createUserRequest struct {
	Email string `json:"email"`
	Name  string `json:"name"`
	Role  string `json:"role"`
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req createUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if err := validation.Required(req.Email, "email"); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := validation.Required(req.Name, "name"); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	u := &User{
		Email: req.Email,
		Name:  req.Name,
		Role:  req.Role,
	}
	if u.Role == "" {
		u.Role = "user"
	}

	if err := h.repo.Create(r.Context(), u); err != nil {
		response.Error(w, http.StatusInternalServerError, "could not create user")
		return
	}

	response.JSON(w, http.StatusCreated, u)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	users, err := h.repo.FindAll(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "could not list users")
		return
	}

	if users == nil {
		users = []User{}
	}

	response.JSON(w, http.StatusOK, users)
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid user id")
		return
	}

	u, err := h.repo.FindByID(r.Context(), uint(id))
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			response.Error(w, http.StatusNotFound, "user not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "could not get user")
		return
	}

	response.JSON(w, http.StatusOK, u)
}
