package mysqlstore

import (
	"log"
	"sync"

	// load mySQL driver
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

// MySQLStore ...
type MySQLStore struct {
	mu sync.Mutex
	db *sqlx.DB
}

// New creates new Store backed by SQLite3
func New(dbFile string) (*MySQLStore, error) {

	log.Print("Connecting to " + "{username}:{password}@({server}:{port})/{database}?parseTime=true")
	d, err := sqlx.Connect("mysql", "{username}:{password}@({server}:{port})/{database}?parseTime=true")
	if err != nil {
		log.Fatal(err)
	}

	err = d.Ping()
	if err != nil {
		log.Fatal(err)
	}
	createSchema(d)

	return &MySQLStore{db: d}, nil
}

// Close ...
func (s *MySQLStore) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.db.Close()
}
