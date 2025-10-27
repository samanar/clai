package vecdb

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
	"github.com/samanar/clai/model"
)

type VecDB struct {
	db *sql.DB
}

// NewVecDB creates or opens the vector database
func NewVecDB() (*VecDB, error) {
	dbPath, err := getDBPath()
	if err != nil {
		return nil, fmt.Errorf("failed to get db path: %w", err)
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create db directory: %w", err)
	}

	// Open with FTS5 extension enabled
	db, err := sql.Open("sqlite3", dbPath+"?_fts5=1")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	vecdb := &VecDB{db: db}

	// Initialize schema
	if err := vecdb.initSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return vecdb, nil
}

// getDBPath returns the path to the vector database file
func getDBPath() (string, error) {
	appDataDir, err := model.AppDataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(appDataDir, "db", "manpages.db"), nil
}

// initSchema creates the necessary tables
func (v *VecDB) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS man_pages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		section INTEGER NOT NULL,
		description TEXT,
		content TEXT NOT NULL,
		file_path TEXT NOT NULL,
		indexed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(name, section)
	);

	CREATE INDEX IF NOT EXISTS idx_man_name ON man_pages(name);
	CREATE INDEX IF NOT EXISTS idx_man_section ON man_pages(section);

	-- Virtual table for full-text search
	CREATE VIRTUAL TABLE IF NOT EXISTS man_pages_fts USING fts5(
		name, 
		description, 
		content,
		content='man_pages',
		content_rowid='id'
	);

	-- Triggers to keep FTS in sync
	CREATE TRIGGER IF NOT EXISTS man_pages_ai AFTER INSERT ON man_pages BEGIN
		INSERT INTO man_pages_fts(rowid, name, description, content)
		VALUES (new.id, new.name, new.description, new.content);
	END;

	CREATE TRIGGER IF NOT EXISTS man_pages_ad AFTER DELETE ON man_pages BEGIN
		DELETE FROM man_pages_fts WHERE rowid = old.id;
	END;

	CREATE TRIGGER IF NOT EXISTS man_pages_au AFTER UPDATE ON man_pages BEGIN
		DELETE FROM man_pages_fts WHERE rowid = old.id;
		INSERT INTO man_pages_fts(rowid, name, description, content)
		VALUES (new.id, new.name, new.description, new.content);
	END;
	`

	_, err := v.db.Exec(schema)
	return err
}

// Close closes the database connection
func (v *VecDB) Close() error {
	if v.db != nil {
		return v.db.Close()
	}
	return nil
}

// GetManPageCount returns the total number of indexed man pages
func (v *VecDB) GetManPageCount() (int, error) {
	var count int
	err := v.db.QueryRow("SELECT COUNT(*) FROM man_pages").Scan(&count)
	return count, err
}

// IsIndexed checks if database has been populated
func (v *VecDB) IsIndexed() (bool, error) {
	count, err := v.GetManPageCount()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetTableNames returns a list of all tables in the database
func (v *VecDB) GetTableNames() ([]string, error) {
	rows, err := v.db.Query("SELECT name FROM sqlite_master WHERE type='table' ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		tables = append(tables, name)
	}

	return tables, rows.Err()
}
