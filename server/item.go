package server

import (
	"encoding/json"
	"github.com/Av1shay/di-demo/authentication"
	"github.com/Av1shay/di-demo/pkg/errs"
	"github.com/Av1shay/di-demo/pkg/log"
	"github.com/Av1shay/di-demo/pkg/types"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strconv"
)

func (s *Server) GetItemByNameHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, err := authentication.UserFromContext(ctx)
	if err != nil {
		errorResponse(ctx, w, errs.NewUnauthorizedErr(err, "Can't find authenticated user"))
		return
	}
	name := chi.URLParam(r, "name")
	item, err := s.uamAPI.GetItemByName(ctx, name, user.AccountID)
	if err != nil {
		errorResponse(ctx, w, err)
		return
	}
	successResponse(ctx, w, http.StatusOK, item)
}

func (s *Server) AddItemHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, err := authentication.UserFromContext(ctx)
	if err != nil {
		errorResponse(ctx, w, errs.NewUnauthorizedErr(err, "Can't find authenticated user"))
		return
	}

	var input types.ItemCreateInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		errorResponse(ctx, w, errs.NewBadRequestErr(err, "Failed to json-decode request"))
		return
	}

	log.Infof(ctx, "AddItem with input: %+v", input)

	if err := s.validator.Struct(&input); err != nil {
		errorResponse(ctx, w, errs.NewBadRequestErr(err, ""))
		return
	}
	item, err := s.uamAPI.CreateItem(ctx, input, user.AccountID)
	if err != nil {
		errorResponse(ctx, w, err)
		return
	}

	successResponse(ctx, w, http.StatusCreated, item)
}

func (s *Server) UpdateItemHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, err := authentication.UserFromContext(ctx)
	if err != nil {
		errorResponse(ctx, w, errs.NewUnauthorizedErr(err, "Can't find authenticated user"))
		return
	}

	id := chi.URLParam(r, "id")

	var input types.UpdateItemInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		errorResponse(ctx, w, errs.NewBadRequestErr(err, "Failed to json-decode request"))
		return
	}

	log.Infof(ctx, "UpdateItem with input: %+v", input)

	if err := s.validator.Struct(&input); err != nil {
		errorResponse(ctx, w, errs.NewBadRequestErr(err, ""))
		return
	}

	input.ID = id

	item, err := s.uamAPI.UpdateItem(ctx, input, user.AccountID)
	if err != nil {
		errorResponse(ctx, w, err)
		return
	}

	successResponse(ctx, w, http.StatusOK, item)
}

func (s *Server) ListItemsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, err := authentication.UserFromContext(ctx)
	if err != nil {
		errorResponse(ctx, w, errs.NewUnauthorizedErr(err, "Can't find authenticated user"))
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	input := types.ListItemsInput{
		Sort:    types.Sort(r.URL.Query().Get("sort")),
		OrderBy: types.OrderBy(r.URL.Query().Get("order_by")),
		Limit:   limit,
	}

	log.Infof(ctx, "ListItems with input: %+v", input)

	if err := s.validator.Struct(&input); err != nil {
		errorResponse(ctx, w, errs.NewBadRequestErr(err, ""))
		return
	}

	items, err := s.uamAPI.ListItems(ctx, input, user.AccountID)
	if err != nil {
		errorResponse(ctx, w, err)
		return
	}

	successResponse(ctx, w, http.StatusOK, items)
}

func (s *Server) DeleteItemHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, err := authentication.UserFromContext(ctx)
	if err != nil {
		errorResponse(ctx, w, errs.NewUnauthorizedErr(err, "Can't find authenticated user"))
		return
	}

	id := chi.URLParam(r, "id")

	log.Infof(ctx, "DeleteItem with id: %s", id)

	if err := s.uamAPI.DeleteItem(ctx, id, user.AccountID); err != nil {
		errorResponse(ctx, w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
