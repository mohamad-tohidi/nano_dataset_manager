package main

import (
	"database/sql"
	"time"

	_ "modernc.org/sqlite"
)

type Store struct {
	db *sql.DB
}

func NewStore(dbPath string) (*Store, error) {
	db, err := sql.Open("sqlite", dbPath+"?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)")
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)

	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS datasets (
			id          TEXT PRIMARY KEY,
			name        TEXT NOT NULL,
			source_type TEXT NOT NULL,
			source_ref  TEXT NOT NULL,
			local_path  TEXT NOT NULL,
			size_bytes  INTEGER NOT NULL DEFAULT 0,
			status      TEXT NOT NULL DEFAULT 'pending',
			created_at  TEXT NOT NULL,
			updated_at  TEXT NOT NULL,
			error_msg   TEXT
		)
	`); err != nil {
		db.Close()
		return nil, err
	}

	return &Store{db: db}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) Insert(d *Dataset) error {
	_, err := s.db.Exec(
		`INSERT INTO datasets (id, name, source_type, source_ref, local_path, size_bytes, status, created_at, updated_at, error_msg)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		d.ID, d.Name, d.SourceType, d.SourceRef, d.LocalPath, d.SizeBytes,
		d.Status, d.CreatedAt.Format(time.RFC3339), d.UpdatedAt.Format(time.RFC3339), d.ErrorMsg,
	)
	return err
}

func (s *Store) GetByID(id string) (*Dataset, error) {
	row := s.db.QueryRow(
		`SELECT id, name, source_type, source_ref, local_path, size_bytes, status, created_at, updated_at, COALESCE(error_msg, '')
		 FROM datasets WHERE id = ?`, id,
	)

	d := &Dataset{}
	var createdAt, updatedAt string
	if err := row.Scan(&d.ID, &d.Name, &d.SourceType, &d.SourceRef,
		&d.LocalPath, &d.SizeBytes, &d.Status, &createdAt, &updatedAt, &d.ErrorMsg); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	d.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	d.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
	return d, nil
}

func (s *Store) List() ([]Dataset, error) {
	rows, err := s.db.Query(
		`SELECT id, name, source_type, source_ref, local_path, size_bytes, status, created_at, updated_at, COALESCE(error_msg, '')
		 FROM datasets ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var datasets []Dataset
	for rows.Next() {
		var d Dataset
		var createdAt, updatedAt string
		if err := rows.Scan(&d.ID, &d.Name, &d.SourceType, &d.SourceRef,
			&d.LocalPath, &d.SizeBytes, &d.Status, &createdAt, &updatedAt, &d.ErrorMsg); err != nil {
			return nil, err
		}
		d.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		d.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
		datasets = append(datasets, d)
	}
	if datasets == nil {
		datasets = []Dataset{}
	}
	return datasets, rows.Err()
}

func (s *Store) UpdateStatus(id, status string, sizeBytes int64, localPath, errorMsg string) error {
	_, err := s.db.Exec(
		`UPDATE datasets SET status = ?, size_bytes = ?, local_path = ?, error_msg = ?, updated_at = ? WHERE id = ?`,
		status, sizeBytes, localPath, errorMsg, now().Format(time.RFC3339), id,
	)
	return err
}

func (s *Store) Delete(id string) error {
	_, err := s.db.Exec(`DELETE FROM datasets WHERE id = ?`, id)
	return err
}
