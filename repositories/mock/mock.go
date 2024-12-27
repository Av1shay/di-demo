package mock

import (
	"context"
	"github.com/Av1shay/di-demo/pkg/types"
)

type Repository struct {
	ReturnErr        error
	GetItemByNameIn  string
	GetItemByNameRes types.Item
	SaveItemIn       types.ItemCreateInput
	UpdateItemIn     types.UpdateItemInput
	SaveItemRes      types.Item
	UpdateItemRes    types.Item
	ListItemsIn      types.ListItemsInput
	ListItemsRes     []types.Item
	DeleteItemRes    error
	DeleteItemIn     string
	HcRes            error
}

func (m *Repository) GetItemByName(ctx context.Context, name, accountID string) (types.Item, error) {
	if m.ReturnErr != nil {
		return types.Item{}, m.ReturnErr
	}
	m.GetItemByNameIn = name
	return m.GetItemByNameRes, nil
}

func (m *Repository) SaveItem(ctx context.Context, input types.ItemCreateInput, accountID string) (types.Item, error) {
	if m.ReturnErr != nil {
		return types.Item{}, m.ReturnErr
	}
	m.SaveItemIn = input
	return m.SaveItemRes, nil
}

func (m *Repository) ListItems(ctx context.Context, input types.ListItemsInput, accountID string) ([]types.Item, error) {
	if m.ReturnErr != nil {
		return nil, m.ReturnErr
	}
	m.ListItemsIn = input
	return m.ListItemsRes, nil
}

func (m *Repository) UpdateItem(ctx context.Context, input types.UpdateItemInput, accountID string) (types.Item, error) {
	if m.ReturnErr != nil {
		return types.Item{}, m.ReturnErr
	}
	m.UpdateItemIn = input
	return m.UpdateItemRes, nil
}

func (m *Repository) DeleteItem(ctx context.Context, id, accountID string) error {
	if m.ReturnErr != nil {
		return m.ReturnErr
	}
	m.DeleteItemIn = id
	return m.DeleteItemRes
}

func (m *Repository) Ping(ctx context.Context) error {
	return m.HcRes
}
