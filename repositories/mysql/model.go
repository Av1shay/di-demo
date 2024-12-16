package mysql

import (
	"database/sql"
	"time"
)

type Item struct {
	ID        int
	Name      string
	Value     sql.NullString
	CreatedAt time.Time
	UpdatedAt time.Time
}
