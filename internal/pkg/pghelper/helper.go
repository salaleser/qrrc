package pghelper

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/pkg/errors"
)

type PGHelper struct {
	connStr string
	db      *sql.DB
}

type Connection struct {
	User     string
	Password string
	Host     string
	Port     string
	Database string
}

func New(c Connection) (*PGHelper, error) {
	h := &PGHelper{
		connStr: fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
			c.User, c.Password, c.Host, c.Port, c.Database),
	}

	var err error
	if h.db, err = sql.Open("postgres", h.connStr); err != nil {
		return nil, errors.Wrap(err, "open db")
	}

	return h, nil
}

func (h *PGHelper) Close() error {
	if err := h.db.Close(); err != nil {
		fmt.Printf("close db: %v", err)
		return errors.Wrap(err, "close db")
	}
	return nil
}

func (h *PGHelper) Query(query string) ([]string, error) {
	rows, err := h.db.Query(query)
	if err != nil {
		return nil, errors.Wrap(err, "query")
	}
	defer rows.Close()

	var result []string
	for rows.Next() {
		var value string
		if err := rows.Scan(&value); err != nil {
			return nil, errors.Wrap(err, "next row")
		}
		result = append(result, value)
	}

	return result, nil
}

func (h *PGHelper) Route(command string) ([]string, error) {
	rows, err := h.db.Query("SELECT side.route(1, '1', -1, '11', '" + command + "');")
	if err != nil {
		return nil, errors.Wrap(err, "query")
	}
	defer rows.Close()

	var result []string
	for rows.Next() {
		var value string
		if err := rows.Scan(&value); err != nil {
			return nil, errors.Wrap(err, "next row")
		}
		result = append(result, value)
	}

	return result, nil
}
