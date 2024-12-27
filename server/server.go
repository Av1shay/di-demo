package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Av1shay/di-demo/pkg/errs"
	"github.com/Av1shay/di-demo/pkg/log"
	"github.com/Av1shay/di-demo/pkg/types"
	"github.com/Av1shay/di-demo/uam"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"net/http"
	"sync/atomic"
	"time"
)

var errCodeToHttpCode = map[errs.ErrorCode]int{
	errs.ErrorCodeNotFound:     http.StatusNotFound,
	errs.ErrorCodeDuplicate:    http.StatusConflict,
	errs.ErrorCodeInternal:     http.StatusInternalServerError,
	errs.ErrorCodeValidation:   http.StatusBadRequest,
	errs.ErrorCodeUnauthorized: http.StatusUnauthorized,
	errs.ErrorCodeBadRequest:   http.StatusBadRequest,
}

type Authenticator interface {
	ContextWithUserAuth(ctx context.Context, token string) (context.Context, error)
}

type Server struct {
	router    *chi.Mux
	validator *validator.Validate
	auth      Authenticator
	uamAPI    *uam.API
	hcStatus  atomic.Bool
}

func New(ctx context.Context, validator *validator.Validate, auth Authenticator, uamAPI *uam.API) (*Server, error) {
	s := &Server{
		validator: validator,
		auth:      auth,
		uamAPI:    uamAPI,
	}
	if err := s.registerValidators(); err != nil {
		return nil, fmt.Errorf("failed to register validators: %w", err)
	}
	go s.startHealthCheck(ctx, time.Second)
	return s, nil
}

func (s *Server) MountHandlers() {
	s.router = chi.NewRouter()
	s.router.Use(TraceIDMiddleware)
	s.router.Use(LogMiddleware)

	s.router.Get("/health-check", s.HealthCheckHandler)

	s.router.Group(func(r chi.Router) {
		r.Use(s.AuthMiddleware)
		r.Get("/item/{name}", s.GetItemByNameHandler)
		r.Post("/item", s.AddItemHandler)
		r.Put("/item/{id}", s.UpdateItemHandler)
		r.Get("/items", s.ListItemsHandler)
		r.Delete("/item/{id}", s.DeleteItemHandler)
	})
}

func (s *Server) Router() *chi.Mux {
	return s.router
}

func (s *Server) registerValidators() error {
	if err := s.validator.RegisterValidation("is_valid_sort", types.IsValidSort); err != nil {
		return err
	}
	if err := s.validator.RegisterValidation("is_valid_orderby", types.IsValidOrderBy); err != nil {
		return err
	}
	return nil
}

func (s *Server) startHealthCheck(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			func() {
				ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
				defer cancel()
				if err := s.uamAPI.HealthCheck(ctx); err != nil {
					log.Errorf(ctx, "Health check failed: %v", err)
					s.hcStatus.Store(false)
					return
				}
				s.hcStatus.Store(true)
			}()
		case <-ctx.Done():
			return
		}
	}
}

func (s *Server) HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	if s.hcStatus.Load() {
		w.WriteHeader(http.StatusOK)
		return
	}
	w.WriteHeader(http.StatusServiceUnavailable)
}

func successResponse(ctx context.Context, w http.ResponseWriter, code int, val any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(val); err != nil {
		log.Errorf(ctx, "Failed to encode value: %v", err)
		return
	}
}

func errorResponse(ctx context.Context, w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")

	code := http.StatusInternalServerError
	msg := "Something went wrong"

	var apiErr *errs.AppError
	if errors.As(err, &apiErr) {
		if c, ok := errCodeToHttpCode[apiErr.Code]; ok {
			code = c
		}
		msg = apiErr.Msg
		err = apiErr.Err
	}

	if code == http.StatusInternalServerError {
		log.Errorf(ctx, "Server error: %v", err)
	}

	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
