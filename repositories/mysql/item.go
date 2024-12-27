package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/Av1shay/di-demo/pkg/errs"
	"github.com/Av1shay/di-demo/pkg/types"
	"github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"strings"
	"time"
)

const (
	itemsTbName           = "items"
	CreateItemsTableQuery = `
		CREATE TABLE IF NOT EXISTS items (
			id CHAR(36) NOT NULL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			account_id VARCHAR(255) NOT NULL,
			value VARCHAR(255) NULL,
			version INT NOT NULL DEFAULT 1,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, 
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		    UNIQUE (name, account_id)
		)`
)

var orderBys = map[types.OrderBy]string{
	types.OrderByName:      "name",
	types.OrderByCreatedAt: "created_at",
	types.OrderByUpdatedAt: "updated_at",
}

type Item struct {
	ID        string
	AccountID string
	Name      string
	Value     sql.NullString
	Version   int
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (r *Repository) GetItem(ctx context.Context, id, accountID string) (types.Item, error) {
	query := fmt.Sprintf("SELECT id, name, value, account_id, version, created_at, updated_at FROM %s WHERE id=? AND account_id=?",
		itemsTbName)
	var item Item
	err := r.db.QueryRowContext(ctx, query, id, accountID).
		Scan(&item.ID, &item.Name, &item.Value, &item.AccountID, &item.Version, &item.CreatedAt, &item.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return types.Item{}, &errs.AppError{
			Code: errs.ErrorCodeNotFound,
			Msg:  fmt.Sprintf("item with id %s not found", id),
			Err:  err,
		}
	}
	if err != nil {
		return types.Item{}, err
	}
	return parseItem(item), nil
}

func (r *Repository) GetItemByName(ctx context.Context, name, accountID string) (types.Item, error) {
	query := fmt.Sprintf("SELECT id, name, value, account_id, version, created_at, updated_at FROM %s WHERE name=? AND account_id=?",
		itemsTbName)
	var item Item
	err := r.db.QueryRowContext(ctx, query, name, accountID).
		Scan(&item.ID, &item.Name, &item.Value, &item.AccountID, &item.Version, &item.CreatedAt, &item.UpdatedAt)

	if errors.Is(err, sql.ErrNoRows) {
		return types.Item{}, &errs.AppError{
			Code: errs.ErrorCodeNotFound,
			Msg:  fmt.Sprintf("item '%s' not found", name),
			Err:  err,
		}
	}
	if err != nil {
		return types.Item{}, err
	}
	return parseItem(item), nil
}

func (r *Repository) SaveItem(ctx context.Context, input types.ItemCreateInput, accountID string) (types.Item, error) {
	id := uuid.NewString()
	query := fmt.Sprintf("INSERT INTO %s (id, name, value, account_id) VALUES (?,?,?,?)", itemsTbName)
	_, err := r.db.ExecContext(ctx, query, id, input.Name, input.Value, accountID)
	if err != nil {
		var msqlErr *mysql.MySQLError
		if errors.As(err, &msqlErr) && msqlErr.Number == 1062 {
			return types.Item{}, &errs.AppError{
				Code: errs.ErrorCodeDuplicate,
				Msg:  fmt.Sprintf("item '%s' already exist", input.Name),
				Err:  err,
			}
		}
		return types.Item{}, err
	}

	return r.GetItem(ctx, id, accountID)
}

func (r *Repository) ListItems(ctx context.Context, input types.ListItemsInput, accountID string) ([]types.Item, error) {
	query := fmt.Sprintf("SELECT id, name, account_id, value, version, created_at, updated_at FROM %s WHERE account_id=?", itemsTbName)

	qb := strings.Builder{}
	qb.WriteString(query)

	sort := "ASC"
	if input.Sort == types.DESC {
		sort = "DESC"
	}
	orderBy := "updated_at"
	if ob, ok := orderBys[input.OrderBy]; ok {
		orderBy = ob
	}
	qb.WriteString(fmt.Sprintf(" ORDER BY %s %s", orderBy, sort))

	if input.Limit > 0 {
		qb.WriteString(fmt.Sprintf(" LIMIT %d", input.Limit))
	}

	rows, err := r.db.QueryContext(ctx, qb.String(), accountID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	defer rows.Close()

	res := make([]types.Item, 0, 10)
	for rows.Next() {
		var item Item
		err := rows.Scan(&item.ID, &item.Name, &item.AccountID, &item.Value, &item.Version, &item.CreatedAt, &item.UpdatedAt)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, nil
			}
			return nil, err
		}
		res = append(res, parseItem(item))
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return res, nil
}

func (r *Repository) UpdateItem(ctx context.Context, input types.UpdateItemInput, accountID string) (types.Item, error) {
	query := fmt.Sprintf("UPDATE %s SET name=?,value=?,version=version+1 WHERE id=? AND account_id=?", itemsTbName)
	_, err := r.db.ExecContext(ctx, query, input.Name, input.Value, input.ID, accountID)
	if err != nil {
		return types.Item{}, err
	}
	return r.GetItem(ctx, input.ID, accountID)
}

func (r *Repository) DeleteItem(ctx context.Context, id, accountID string) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE id=? AND account_id=?", itemsTbName)
	_, err := r.db.ExecContext(ctx, query, id, accountID)
	return err
}

func parseItem(item Item) types.Item {
	val := ""
	if item.Value.Valid {
		val = item.Value.String
	}
	return types.Item{
		ID:        item.ID,
		Name:      item.Name,
		Value:     val,
		Version:   item.Version,
		AccountID: item.AccountID,
		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
	}
}
