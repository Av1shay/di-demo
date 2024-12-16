package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/Av1shay/di-demo/pkg/types"
	_ "github.com/go-sql-driver/mysql"
)

const itemsTbName = "items"

type Repository struct {
	db *sql.DB
}

func NewRepository(conn string) (*Repository, error) {
	db, err := sql.Open("mysql", conn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return &Repository{db}, nil
}

func (r *Repository) GetItemByName(ctx context.Context, name string) (types.Item, error) {
	query := fmt.Sprintf("SELECT id, name, value, created_at, updated_at FROM %s WHERE name=?", itemsTbName)
	var item Item
	err := r.db.QueryRowContext(ctx, query, name).
		Scan(&item.ID, &item.Name, &item.Value, &item.CreatedAt, &item.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return types.Item{}, &types.APIError{
			Code: types.ErrorCodeNotFound,
			Msg:  fmt.Sprintf("item with name %s not found", name),
		}
	}
	if err != nil {
		return types.Item{}, err
	}
	return parseItem(item), nil
}

func (r *Repository) SaveItem(ctx context.Context, input types.ItemCreateInput) (types.Item, error) {
	query := fmt.Sprintf("INSERT INTO %s (name, value) VALUES (?,?)", itemsTbName)
	_, err := r.db.ExecContext(ctx, query, input.Name, input.Value)
	if err != nil {
		return types.Item{}, err
	}
	return r.GetItemByName(ctx, input.Name)
}

func (r *Repository) ListItems(ctx context.Context) ([]types.Item, error) {
	query := fmt.Sprintf("SELECT id, name, value, createdAt, updatedAt FROM %s", itemsTbName)
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []types.Item
	for rows.Next() {
		var item Item
		err := rows.Scan(&item.ID, &item.Name, &item.Value, &item.CreatedAt, &item.UpdatedAt)
		if err != nil {
			return nil, err
		}
		res = append(res, parseItem(item))
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return res, nil
}

func (r *Repository) DeleteItem(ctx context.Context, id string) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE id=?", itemsTbName)
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func parseItem(item Item) types.Item {
	val := ""
	if item.Value.Valid {
		val = item.Value.String
	}
	return types.Item{
		ID:        fmt.Sprintf("%d", item.ID),
		Name:      item.Name,
		Value:     val,
		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
	}
}
