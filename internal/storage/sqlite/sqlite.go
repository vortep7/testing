package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"url-shortener/internal/storage"

	"github.com/mattn/go-sqlite3"
)

type Storage struct {
	db *sql.DB
}

func (s *Storage) AliasChecker(alias string) (bool, error) {
	const op = "storage.sqlite.AliasChecker"
	var exists bool
	stmt, err := s.db.Prepare("SELECT EXISTS(SELECT 1 FROM url WHERE alias = ?)")
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}
	err = stmt.QueryRow(alias).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}
	return exists, nil
}

func New(storagePath string) (*Storage, error) {
	const op = "storage.sqlite.New"

	db, err := sql.Open("sqlite3", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	stmt, err := db.Prepare(
		`CREATE TABLE IF NOT EXISTS url(
	     id INTEGER PRIMARY KEY,
	     alias TEXT NOT NULL UNIQUE,
	     url TEXT NOT NULL);
	    CREATE INDEX IF NOT EXISTS idx_alias ON url(alias)`)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	_, err = stmt.Exec()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &Storage{db: db}, nil
}

func (s *Storage) SaveURL(urlName string, alias string) (int64, error) {
	const op = "storage.sqlite.SaveURL"

	stmt, err := s.db.Prepare("INSERT INTO url(url, alias) VALUES(?, ?)")
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	res, err := stmt.Exec(urlName, alias)
	///TODO: refactor this
	if err != nil {
		if sql3Err, ok := err.(sqlite3.Error); ok {
			if sql3Err.ExtendedCode == sqlite3.ErrConstraintUnique {
				return 0, fmt.Errorf("%s: %w", op, storage.ErrURLExists)
			}
			return 0, fmt.Errorf("%s: %w", op, err)
		}
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	return id, nil
}

func (s *Storage) GetURL(alias string) (string, error) {
	const op = "storage.sqlite.GetURL"

	stmt, err := s.db.Prepare("SELECT url FROM url WHERE alias = ?")
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}
	var resUrl string
	err = stmt.QueryRow(alias).Scan(&resUrl)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", storage.ErrURLNotFound
		}
		return "", fmt.Errorf("%s: %w", op, err)
	}
	return resUrl, nil
}

func (s *Storage) DeleteURL(alias string) error {
	const op = "storage.sqlite.DeleteURL"

	stmt, err := s.db.Prepare("DELETE FROM url WHERE alias = ?")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	res, err := stmt.Exec(alias)
	if err != nil {

		return fmt.Errorf("%s: %w", op, err)

	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if rowsAffected == 0 {
		return storage.ErrURLNotFound
	}

	return nil
}
