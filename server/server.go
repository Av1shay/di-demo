package server

import (
	"encoding/json"
	"errors"
	"github.com/Av1shay/di-demo/pkg/types"
	"github.com/Av1shay/di-demo/uam"
	"github.com/go-chi/chi/v5"
	"net/http"
)

var errCodeToHttpCode = map[types.ErrorCode]int{
	types.ErrorCodeNotFound:   404,
	types.ErrorCodeDuplicate:  409,
	types.ErrorCodeInternal:   500,
	types.ErrorCodeValidation: 400,
}

type Server struct {
	router *chi.Mux
	uamAPI *uam.API
}

func New(uamAPI *uam.API) *Server {
	return &Server{
		uamAPI: uamAPI,
	}
}

func (s *Server) MountHandlers() {
	s.router = chi.NewRouter()
	s.router.Get("/item/{name}", s.GetUserByName)
	s.router.Post("/item", s.AddItem)
}

func (s *Server) Router() *chi.Mux {
	return s.router
}

func (s *Server) GetUserByName(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	name := chi.URLParam(r, "name")
	item, err := s.uamAPI.GetItem(ctx, name)
	if err != nil {
		errorResponse(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(item); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Server) AddItem(w http.ResponseWriter, r *http.Request) {
	// todo
}

func errorResponse(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")

	code := http.StatusInternalServerError
	msg := err.Error()

	var apiErr *types.APIError
	if errors.As(err, &apiErr) {
		if c, ok := errCodeToHttpCode[apiErr.Code]; ok {
			code = c
		}
		msg = apiErr.Msg
	}
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
