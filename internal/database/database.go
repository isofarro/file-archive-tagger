package database

import (
	"database/sql"
	"fmt"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	*sql.DB
}

// New creates a new database connection
func New(dbPath string) (*DB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	return &DB{db}, nil
}

// Initialize creates the database schema
func (db *DB) Initialize() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS files (
            id INTEGER PRIMARY KEY,
            filename TEXT NOT NULL,
            path TEXT NOT NULL,
            hash TEXT NOT NULL,
            size INTEGER NOT NULL,
            modified_at DATETIME NOT NULL,
            UNIQUE(path, filename)
        )`,
		`CREATE TABLE IF NOT EXISTS taxonomies (
            id INTEGER PRIMARY KEY,
            name TEXT NOT NULL UNIQUE
        )`,
		`CREATE TABLE IF NOT EXISTS tags (
            id INTEGER PRIMARY KEY,
            taxonomy_id INTEGER NOT NULL,
            name TEXT NOT NULL,
            FOREIGN KEY(taxonomy_id) REFERENCES taxonomies(id),
            UNIQUE(taxonomy_id, name)
        )`,
		`CREATE TABLE IF NOT EXISTS file_tags (
            file_id INTEGER NOT NULL,
            tag_id INTEGER NOT NULL,
            PRIMARY KEY(file_id, tag_id),
            FOREIGN KEY(file_id) REFERENCES files(id),
            FOREIGN KEY(tag_id) REFERENCES tags(id)
        )`,
		`CREATE TABLE IF NOT EXISTS stage_directory (
            id INTEGER PRIMARY KEY,
            path TEXT NOT NULL UNIQUE
        )`,
		`INSERT OR IGNORE INTO taxonomies (name) VALUES ('tags')`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("failed to execute query: %w", err)
		}
	}
	return nil
}

// AddFile adds a file to the database
func (db *DB) AddFile(filename, path, hash string, size int64, modifiedAt string) error {
	query := `
        INSERT INTO files (filename, path, hash, size, modified_at)
        VALUES (?, ?, ?, ?, ?)
        ON CONFLICT(path, filename) DO UPDATE SET
            hash = excluded.hash,
            size = excluded.size,
            modified_at = excluded.modified_at
    `
	_, err := db.Exec(query, filename, path, hash, size, modifiedAt)
	if err != nil {
		return fmt.Errorf("failed to add file: %w", err)
	}
	return nil
}

// FileExists checks if a file exists in the database by its hash
func (db *DB) FileExists(hash string) (bool, error) {
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM files WHERE hash = ?)", hash).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check file existence: %w", err)
	}
	return exists, nil
}

// AddTaxonomy creates a new taxonomy
func (db *DB) AddTaxonomy(name string) error {
	_, err := db.Exec("INSERT INTO taxonomies (name) VALUES (?)", name)
	if err != nil {
		return fmt.Errorf("failed to add taxonomy: %w", err)
	}
	return nil
}

// TagFile adds a tag to a file
func (db *DB) TagFile(filePath, taxonomyName, tagName string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Get path components
	filename := filepath.Base(filePath)

	// Get file ID using path components
	var fileID int64
	err = tx.QueryRow("SELECT id FROM files WHERE path = ? AND filename = ?", "players", filename).Scan(&fileID)
	if err != nil {
		return fmt.Errorf("file not found: %w", err)
	}

	// Get or create taxonomy
	var taxonomyID int64
	err = tx.QueryRow(`
        INSERT INTO taxonomies (name) 
        VALUES (?) 
        ON CONFLICT(name) DO UPDATE SET name = excluded.name 
        RETURNING id`, taxonomyName).Scan(&taxonomyID)
	if err != nil {
		return fmt.Errorf("failed to get/create taxonomy: %w", err)
	}

	// Insert or get tag
	var tagID int64
	err = tx.QueryRow(`
        INSERT INTO tags (taxonomy_id, name) 
        VALUES (?, ?)
        ON CONFLICT(taxonomy_id, name) DO UPDATE SET name = excluded.name
        RETURNING id
    `, taxonomyID, tagName).Scan(&tagID)
	if err != nil {
		return fmt.Errorf("failed to create/get tag: %w", err)
	}

	// Link file to tag
	_, err = tx.Exec(`
        INSERT INTO file_tags (file_id, tag_id)
        VALUES (?, ?)
        ON CONFLICT(file_id, tag_id) DO NOTHING
    `, fileID, tagID)
	if err != nil {
		return fmt.Errorf("failed to tag file: %w", err)
	}

	return tx.Commit()
}

// SearchByTag returns all files with a specific tag
func (db *DB) SearchByTag(taxonomyName, tagName string) ([]string, error) {
	query := `
        SELECT f.path || '/' || f.filename
        FROM files f
        JOIN file_tags ft ON f.id = ft.file_id
        JOIN tags t ON ft.tag_id = t.id
        JOIN taxonomies tax ON t.taxonomy_id = tax.id
        WHERE tax.name = ? AND t.name = ?
    `
	rows, err := db.Query(query, taxonomyName, tagName)
	if err != nil {
		return nil, fmt.Errorf("failed to search files: %w", err)
	}
	defer rows.Close()

	var files []string
	for rows.Next() {
		var file string
		if err := rows.Scan(&file); err != nil {
			return nil, err
		}
		files = append(files, file)
	}
	return files, nil
}

// GetFilePathByHash returns the filepath of a file with the given hash
func (db *DB) GetFilePathByHash(hash string) (string, error) {
	var path, filename string
	err := db.QueryRow("SELECT path, filename FROM files WHERE hash = ?", hash).Scan(&path, &filename)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("failed to query file path: %w", err)
	}

	return filepath.Join(path, filename), nil
}

// GetAllFiles returns all file paths in the database
func (db *DB) GetAllFiles() ([]string, error) {
    query := `SELECT path || '/' || filename FROM files`
    rows, err := db.Query(query)
    if err != nil {
        return nil, fmt.Errorf("failed to query files: %w", err)
    }
    defer rows.Close()

    var files []string
    for rows.Next() {
        var file string
        if err := rows.Scan(&file); err != nil {
            return nil, err
        }
        files = append(files, file)
    }
    return files, nil
}
