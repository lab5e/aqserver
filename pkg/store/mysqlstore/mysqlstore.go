package mysqlstore

import (
	"log"

	// load mySQL driver
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

// MySQLStore ...
type MySQLStore struct {
	connectString string
	db            *sqlx.DB
}

// New creates new Store backed by SQLite3
func New(connectString string) (*MySQLStore, error) {
	d, err := sqlx.Connect("mysql", connectString)
	if err != nil {
		log.Fatal(err)
	}

	err = d.Ping()
	if err != nil {
		log.Fatal(err)
	}
	createSchema(d)

	return &MySQLStore{
		db:            d,
		connectString: connectString,
	}, nil
}

// Close ...
func (s *MySQLStore) Close() error {
	return s.db.Close()
}
