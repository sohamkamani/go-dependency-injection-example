package database

import (
	"database/sql"
)

type Store interface {
	Get(ID int) (int, error)
}

func NewStore(db *sql.DB) Store {
	return &store{db}
}

// The actual store would contain some state. In this case it's the sql.db instance, that holds the connection to our database
type store struct {
	db *sql.DB
}

func (d *store) Get(ID int) (int, error) {
	//we would perform some external database operation with d.db
	return 0, nil
}
