package server

import (
	"bytes"
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"github.com/Av1shay/di-demo/authentication"
	"github.com/Av1shay/di-demo/cache/memory"
	"github.com/Av1shay/di-demo/pkg/errs"
	"github.com/Av1shay/di-demo/pkg/types"
	"github.com/Av1shay/di-demo/repositories/mock"
	"github.com/Av1shay/di-demo/uam"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"
)

type mockAuthenticator struct {
	user    types.User
	tokenIn string
}

func (m *mockAuthenticator) ContextWithUserAuth(ctx context.Context, token string) (context.Context, error) {
	m.tokenIn = token
	ctx = authentication.ContextWithUser(ctx, &m.user)
	return ctx, nil
}

func TestServer_GetItem(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	item := buildItem()
	mockRepo := &mock.Repository{
		SaveItemRes:      item,
		GetItemByNameRes: item,
	}

	uamAPI, err := uam.NewAPI(uam.Config{}, mockRepo, memory.NewCache())
	require.NoError(t, err)

	user := buildUser(item.AccountID)
	mockAuth := &mockAuthenticator{user: user}
	v := validator.New(validator.WithRequiredStructEnabled())
	serv, err := New(ctx, v, mockAuth, uamAPI)
	require.NoError(t, err)
	serv.MountHandlers()
	ts := httptest.NewServer(serv.Router())
	t.Cleanup(ts.Close)

	t.Run("test_not_exist", func(t *testing.T) {
		mockRepo.ReturnErr = errs.NewNotFoundErr(errors.New("not found"), "")
		t.Cleanup(func() { mockRepo.ReturnErr = nil })
		req, err := http.NewRequestWithContext(ctx, "GET", ts.URL+"/item/not-exist", nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "BEARER "+user.Token)
		resp, err := ts.Client().Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusNotFound, resp.StatusCode)
		b, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Contains(t, string(b), "not found")
	})

	t.Run("test_ok", func(t *testing.T) {
		createdItem, err := uamAPI.CreateItem(ctx, types.ItemCreateInput{}, user.AccountID)
		require.NoError(t, err)

		req, err := http.NewRequestWithContext(ctx, "GET", ts.URL+"/item/xxx", nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "BEARER "+user.Token)

		resp, err := ts.Client().Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, 200, resp.StatusCode)
		b, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		var gotItem types.Item
		require.NoError(t, json.Unmarshal(b, &gotItem))
		require.Equal(t, createdItem.ID, gotItem.ID)
		require.Equal(t, createdItem.Name, gotItem.Name)
		require.Equal(t, createdItem.Value, gotItem.Value)
		require.Equal(t, createdItem.AccountID, gotItem.AccountID)
	})
}

func TestServer_AddItem(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	item := buildItem()
	mockRepo := mock.Repository{
		SaveItemRes:      item,
		GetItemByNameRes: item,
	}

	uamAPI, err := uam.NewAPI(uam.Config{}, &mockRepo, memory.NewCache())
	require.NoError(t, err)

	user := buildUser(item.AccountID)
	mockAuth := &mockAuthenticator{user: user}

	v := validator.New(validator.WithRequiredStructEnabled())
	serv, err := New(ctx, v, mockAuth, uamAPI)
	require.NoError(t, err)
	serv.MountHandlers()
	ts := httptest.NewServer(serv.Router())
	t.Cleanup(ts.Close)

	t.Run("test_validation_err", func(t *testing.T) {
		badItem := types.ItemCreateInput{Value: "some val"}
		b, err := json.Marshal(badItem)
		require.NoError(t, err)
		req, err := http.NewRequestWithContext(ctx, "POST", ts.URL+"/item", bytes.NewReader(b))
		require.NoError(t, err)
		req.Header.Set("Authorization", "BEARER "+user.Token)
		resp, err := ts.Client().Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
		var resErr struct {
			Error string `json:"error"`
		}
		err = json.NewDecoder(resp.Body).Decode(&resErr)
		require.NoError(t, err)
		require.NotEmpty(t, resErr.Error)
		require.Contains(t, strings.ToLower(resErr.Error), "field validation for 'name' failed")
	})

	t.Run("test_ok", func(t *testing.T) {
		name, val := "test-item", "test-val"
		itemCreateIn := types.ItemCreateInput{Name: name, Value: val}
		b, err := json.Marshal(itemCreateIn)
		require.NoError(t, err)
		req, err := http.NewRequestWithContext(ctx, "POST", ts.URL+"/item", bytes.NewReader(b))
		require.NoError(t, err)
		req.Header.Set("Authorization", "BEARER "+user.Token)
		resp, err := ts.Client().Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusCreated, resp.StatusCode)
		var gotItem types.Item
		err = json.NewDecoder(resp.Body).Decode(&gotItem)
		require.NoError(t, err)

		require.NotEmpty(t, gotItem.ID)
		require.Equal(t, mockRepo.SaveItemIn.Name, name)
		require.Equal(t, mockRepo.SaveItemIn.Value, val)

		require.Equal(t, item.Name, gotItem.Name)
		require.Equal(t, item.Value, gotItem.Value)
		require.Equal(t, user.AccountID, gotItem.AccountID)
	})
}

func TestServer_UpdateItem(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	item := buildItem()
	mockRepo := mock.Repository{
		UpdateItemRes:    item,
		GetItemByNameRes: item,
	}

	uamAPI, err := uam.NewAPI(uam.Config{}, &mockRepo, memory.NewCache())
	require.NoError(t, err)

	user := buildUser(item.AccountID)
	mockAuth := &mockAuthenticator{user: user}

	v := validator.New(validator.WithRequiredStructEnabled())
	serv, err := New(ctx, v, mockAuth, uamAPI)
	require.NoError(t, err)
	serv.MountHandlers()
	ts := httptest.NewServer(serv.Router())
	t.Cleanup(ts.Close)

	t.Run("test_not_exist", func(t *testing.T) {
		badItem := types.UpdateItemInput{Name: "not found", Value: "some val"}
		b, err := json.Marshal(badItem)
		require.NoError(t, err)
		mockRepo.ReturnErr = errs.NewNotFoundErr(errors.New("not found"), "")
		t.Cleanup(func() { mockRepo.ReturnErr = nil })
		req, err := http.NewRequestWithContext(ctx, "PUT", ts.URL+"/item/-1", bytes.NewReader(b))
		require.NoError(t, err)
		req.Header.Set("Authorization", "BEARER "+user.Token)
		resp, err := ts.Client().Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusNotFound, resp.StatusCode)
		var resErr struct {
			Error string `json:"error"`
		}
		err = json.NewDecoder(resp.Body).Decode(&resErr)
		require.NoError(t, err)
		require.NotEmpty(t, resErr.Error)
		require.Equal(t, strings.ToLower(resErr.Error), "not found")
	})

	t.Run("test_ok", func(t *testing.T) {
		input := types.UpdateItemInput{ID: "-2", Name: "My nice item", Value: "some val"}
		b, err := json.Marshal(input)
		require.NoError(t, err)
		req, err := http.NewRequestWithContext(ctx, "PUT", ts.URL+"/item/"+item.ID, bytes.NewReader(b))
		require.NoError(t, err)
		req.Header.Set("Authorization", "BEARER "+user.Token)
		resp, err := ts.Client().Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)
		var gotItem types.Item
		err = json.NewDecoder(resp.Body).Decode(&gotItem)
		require.NoError(t, err)
		require.Equal(t, item.ID, gotItem.ID)
		require.Equal(t, item.Name, gotItem.Name)
		require.Equal(t, item.Version, gotItem.Version)
		require.Equal(t, item.AccountID, gotItem.AccountID)
	})
}

func TestServer_ListItems(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	items := []types.Item{buildItem(), buildItem()}
	mockRepo := mock.Repository{ListItemsRes: items}

	uamAPI, err := uam.NewAPI(uam.Config{}, &mockRepo, memory.NewCache())
	require.NoError(t, err)

	user := buildUser(items[0].AccountID)
	mockAuth := &mockAuthenticator{user: user}

	v := validator.New(validator.WithRequiredStructEnabled())
	serv, err := New(ctx, v, mockAuth, uamAPI)
	require.NoError(t, err)
	serv.MountHandlers()
	ts := httptest.NewServer(serv.Router())
	t.Cleanup(ts.Close)

	t.Run("test_validation_err", func(t *testing.T) {
		v := url.Values{}
		v.Set("sort", "not-valid")
		req, err := http.NewRequestWithContext(ctx, "GET", ts.URL+"/items?"+v.Encode(), nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "BEARER "+user.Token)
		resp, err := ts.Client().Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
		var resErr struct {
			Error string `json:"error"`
		}
		err = json.NewDecoder(resp.Body).Decode(&resErr)
		require.NoError(t, err)
		require.NotEmpty(t, resErr.Error)
		require.Contains(t, strings.ToLower(resErr.Error), "field validation for 'sort' failed on the 'is_valid_sort'")
	})

	t.Run("test_ok", func(t *testing.T) {
		sort, orderBy, limit := types.ASC, types.OrderByCreatedAt, 10
		v := url.Values{}
		v.Set("sort", string(sort))
		v.Set("order_by", string(orderBy))
		v.Set("limit", strconv.Itoa(limit))
		req, err := http.NewRequestWithContext(ctx, "GET", ts.URL+"/items?"+v.Encode(), nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "BEARER "+user.Token)
		resp, err := ts.Client().Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)

		var gotItems []types.Item
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&gotItems))

		testListsEqual(t, items, gotItems)

		require.Equal(t, sort, mockRepo.ListItemsIn.Sort)
		require.Equal(t, orderBy, mockRepo.ListItemsIn.OrderBy)
		require.Equal(t, limit, mockRepo.ListItemsIn.Limit)
	})
}

func TestServer_DeleteItem(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	item := buildItem()
	mockRepo := mock.Repository{
		UpdateItemRes:    item,
		GetItemByNameRes: item,
	}

	uamAPI, err := uam.NewAPI(uam.Config{}, &mockRepo, memory.NewCache())
	require.NoError(t, err)

	user := buildUser(item.AccountID)
	mockAuth := &mockAuthenticator{user: user}

	v := validator.New(validator.WithRequiredStructEnabled())
	serv, err := New(ctx, v, mockAuth, uamAPI)
	require.NoError(t, err)
	serv.MountHandlers()
	ts := httptest.NewServer(serv.Router())
	t.Cleanup(ts.Close)

	t.Run("test_not_exist", func(t *testing.T) {
		badItem := types.UpdateItemInput{Name: "not exist", Value: "some val"}
		b, err := json.Marshal(badItem)
		require.NoError(t, err)
		mockRepo.ReturnErr = errs.NewNotFoundErr(errors.New("not found"), "")
		t.Cleanup(func() { mockRepo.ReturnErr = nil })
		req, err := http.NewRequestWithContext(ctx, "PUT", ts.URL+"/item/-1", bytes.NewReader(b))
		require.NoError(t, err)
		req.Header.Set("Authorization", "BEARER "+user.Token)
		resp, err := ts.Client().Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusNotFound, resp.StatusCode)
		var resErr struct {
			Error string `json:"error"`
		}
		err = json.NewDecoder(resp.Body).Decode(&resErr)
		require.NoError(t, err)
		require.NotEmpty(t, resErr.Error)
		require.Equal(t, strings.ToLower(resErr.Error), "not found")
	})

	t.Run("test_ok", func(t *testing.T) {
		input := types.UpdateItemInput{ID: "-2", Name: "My nice item", Value: "some val"}
		b, err := json.Marshal(input)
		require.NoError(t, err)
		req, err := http.NewRequestWithContext(ctx, "PUT", ts.URL+"/item/"+item.ID, bytes.NewReader(b))
		require.NoError(t, err)
		req.Header.Set("Authorization", "BEARER "+user.Token)
		resp, err := ts.Client().Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)
		var gotItem types.Item
		err = json.NewDecoder(resp.Body).Decode(&gotItem)
		require.NoError(t, err)
		require.Equal(t, item.ID, gotItem.ID)
		require.Equal(t, item.Name, gotItem.Name)
		require.Equal(t, item.Version, gotItem.Version)
		require.Equal(t, item.AccountID, gotItem.AccountID)
	})
}

func TestServer_Authentication(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	item := buildItem()

	mockRepo := &mock.Repository{}

	uamAPI, err := uam.NewAPI(uam.Config{}, mockRepo, memory.NewCache())
	require.NoError(t, err)

	user := buildUser(item.AccountID)
	mockAuth := &mockAuthenticator{user: user}

	v := validator.New(validator.WithRequiredStructEnabled())
	serv, err := New(ctx, v, mockAuth, uamAPI)
	require.NoError(t, err)
	serv.MountHandlers()
	ts := httptest.NewServer(serv.Router())
	t.Cleanup(ts.Close)

	req, err := http.NewRequestWithContext(ctx, "DELETE", ts.URL+"/item/"+item.ID, nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "BEARER "+user.Token)
	resp, err := ts.Client().Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, resp.StatusCode)
	require.Equal(t, item.ID, mockRepo.DeleteItemIn)
}

func buildItem() types.Item {
	now := time.Now()
	return types.Item{
		ID:        gofakeit.UUID(),
		Name:      "test-item-" + gofakeit.LetterN(6),
		Value:     "test-value-" + gofakeit.UUID(),
		AccountID: gofakeit.UUID(),
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func buildUser(accountID string) types.User {
	return types.User{
		ID:        gofakeit.UUID(),
		Name:      gofakeit.Name(),
		Email:     gofakeit.Email(),
		AccountID: accountID,
		Token:     gofakeit.UUID(),
	}
}

func testListsEqual(t *testing.T, want, got []types.Item) {
	t.Helper()

	require.Equal(t, len(want), len(got))

	slices.SortFunc(want, func(a, b types.Item) int {
		return cmp.Compare(a.ID, b.ID)
	})
	slices.SortFunc(got, func(a, b types.Item) int {
		return cmp.Compare(a.ID, b.ID)
	})

	for i := range want {
		require.Equal(t, want[i].ID, got[i].ID)
		require.Equal(t, want[i].Name, got[i].Name)
		require.Equal(t, want[i].Value, got[i].Value)
		require.Equal(t, want[i].AccountID, got[i].AccountID)
	}
}

func TestBearerAuthHeader(t *testing.T) {
	t.Parallel()

	for i, tt := range []struct {
		header         string
		wantErrContain string
		wantRes        string
	}{
		{
			header:         "",
			wantErrContain: "no auth header",
		},
		{
			header:  "abc",
			wantRes: "abc",
		},
		{
			header:  "BEARER abc123",
			wantRes: "abc123",
		},
		{
			header:  "bearer abc123",
			wantRes: "abc123",
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			got, err := bearerAuthHeader(tt.header)
			if tt.wantErrContain != "" {
				require.NotNil(t, err)
				require.Contains(t, err.Error(), tt.wantErrContain)
				return
			}
			require.Equal(t, tt.wantRes, got)
		})
	}
}
