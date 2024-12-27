package server

import (
	"errors"
	"github.com/Av1shay/di-demo/pkg/errs"
	"github.com/Av1shay/di-demo/pkg/log"
	"github.com/google/uuid"
	"net/http"
	"strings"
)

type respWriterWithStatus struct {
	http.ResponseWriter
	statusCode int
}

func (rw *respWriterWithStatus) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func TraceIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		traceID := uuid.NewString()
		ctx := log.ContextWithTraceID(r.Context(), traceID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func LogMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health-check" {
			next.ServeHTTP(w, r)
			return
		}
		rw := &respWriterWithStatus{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(rw, r)
		log.Infof(r.Context(), "Request: method=%s path=%s status=%d", r.Method, r.URL.Path, rw.statusCode)
	})
}

func (s *Server) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		token, err := bearerAuthHeader(r.Header.Get("Authorization"))
		if err != nil {
			errorResponse(ctx, w, errs.NewUnauthorizedErr(err, http.StatusText(http.StatusUnauthorized)))
			return
		}

		ctx, err = s.auth.ContextWithUserAuth(ctx, token)
		if err != nil {
			errorResponse(ctx, w, errs.NewUnauthorizedErr(err, http.StatusText(http.StatusUnauthorized)))
			return
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func bearerAuthHeader(header string) (string, error) {
	if header == "" {
		return "", errors.New("no auth header provided")
	}
	bearer := "BEARER"
	if len(header) > len(bearer) && strings.ToUpper(header[0:len(bearer)]) == bearer {
		return header[len(bearer)+1:], nil
	}
	return header, nil
}
