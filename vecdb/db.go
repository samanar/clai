package vecdb

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	_ "github.com/mattn/go-sqlite3"
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
	appDataDir, err := getAppDataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(appDataDir, "db", "manpages.db"), nil
}

// getAppDataDir returns the application data directory (copied from model package to avoid circular import)
func getAppDataDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	if runtime.GOOS == "darwin" {
		return filepath.Join(homeDir, "Library", "Application Support", "Clai"), nil
	}
	// default to Linux behaviour
	if dir := os.Getenv("XDG_DATA_HOME"); dir != "" {
		return filepath.Join(dir, "clai"), nil
	}
	return filepath.Join(homeDir, ".local", "share", "clai"), nil
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
		embedding BLOB,
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

// UpdateEmbedding updates the embedding vector for a man page
func (v *VecDB) UpdateEmbedding(id int, embedding []float32) error {
	bytes := Float32SliceToBytes(embedding)
	_, err := v.db.Exec("UPDATE man_pages SET embedding = ? WHERE id = ?", bytes, id)
	return err
}

// GetManPagesWithoutEmbeddings returns man pages that don't have embeddings yet
func (v *VecDB) GetManPagesWithoutEmbeddings(limit int) ([]struct {
	ID          int
	Name        string
	Description string
	Content     string
}, error) {
	rows, err := v.db.Query(`
		SELECT id, name, description, content 
		FROM man_pages 
		WHERE embedding IS NULL 
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []struct {
		ID          int
		Name        string
		Description string
		Content     string
	}

	for rows.Next() {
		var r struct {
			ID          int
			Name        string
			Description string
			Content     string
		}
		if err := rows.Scan(&r.ID, &r.Name, &r.Description, &r.Content); err != nil {
			continue
		}
		results = append(results, r)
	}

	return results, rows.Err()
}

// SearchByEmbedding finds similar man pages using cosine similarity
func (v *VecDB) SearchByEmbedding(queryEmbedding []float32, limit int) ([]SearchResult, error) {
	rows, err := v.db.Query(`
		SELECT id, name, section, description, content, embedding
		FROM man_pages
		WHERE embedding IS NOT NULL
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type scored struct {
		result SearchResult
		score  float64
	}

	var candidates []scored

	for rows.Next() {
		var id, section int
		var name, description, content string
		var embeddingBytes []byte

		if err := rows.Scan(&id, &name, &section, &description, &content, &embeddingBytes); err != nil {
			continue
		}

		if len(embeddingBytes) == 0 {
			continue
		}

		embedding := BytesToFloat32Slice(embeddingBytes)
		similarity := CosineSimilarity(queryEmbedding, embedding)

		candidates = append(candidates, scored{
			result: SearchResult{
				Name:        name,
				Section:     section,
				Description: description,
				Content:     content,
				Relevance:   similarity,
			},
			score: similarity,
		})
	}

	// Sort by similarity descending
	for i := 0; i < len(candidates); i++ {
		for j := i + 1; j < len(candidates); j++ {
			if candidates[j].score > candidates[i].score {
				candidates[i], candidates[j] = candidates[j], candidates[i]
			}
		}
	}

	// Return top results
	var results []SearchResult
	for i := 0; i < limit && i < len(candidates); i++ {
		results = append(results, candidates[i].result)
	}

	return results, nil
}
