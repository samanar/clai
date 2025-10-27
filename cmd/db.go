package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/samanar/clai/man"
	"github.com/samanar/clai/vecdb"
	"github.com/spf13/cobra"
)

var dbCmd = &cobra.Command{
	Use:   "db",
	Short: "Database management commands",
	Long:  "Manage the man pages database used for context-aware command generation",
}

var dbResetCmd = &cobra.Command{
	Use:     "create",
	Short:   "Reset the database",
	Aliases: []string{"reset", "refresh"},
	Long:    "Drop the existing database and rebuild it from scratch by re-indexing all man pages",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("🗑️  Resetting database...")

		// Get database path
		dbPath, err := getDBPath()
		if err != nil {
			return fmt.Errorf("failed to get database path: %w", err)
		}

		// Remove existing database file if it exists
		if _, err := os.Stat(dbPath); err == nil {
			fmt.Printf("📁 Removing existing database: %s\n", dbPath)
			if err := os.Remove(dbPath); err != nil {
				return fmt.Errorf("failed to remove existing database: %w", err)
			}
		}

		// Create new database
		fmt.Println("🔧 Creating new database...")
		db, err := vecdb.NewVecDB()
		if err != nil {
			fmt.Println("database not created", err)
			return fmt.Errorf("failed to create database: %w", err)
		}
		defer db.Close()

		// Index man pages
		fmt.Println("📚 Indexing man pages...")
		m, err := man.NewMan()
		if err != nil {
			return fmt.Errorf("failed to create man instance: %w", err)
		}
		if err := db.IndexManPages(&m); err != nil {
			return fmt.Errorf("failed to index man pages: %w", err)
		}

		// Get count of indexed pages
		count, err := db.GetManPageCount()
		if err != nil {
			fmt.Printf("failed to get man page count: %w", err)
			os.Exit(1)
		}

		fmt.Printf("\n✅ Database reset complete! Indexed %d man pages\n", count)
		fmt.Printf("📍 Database location: %s\n", dbPath)

		return nil
	},
}

var dbShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show database information",
	Long:  "Display database path, size, and statistics if the database exists",
	RunE: func(cmd *cobra.Command, args []string) error {
		dbPath, err := getDBPath()
		if err != nil {
			return fmt.Errorf("failed to get database path: %w", err)
		}

		fmt.Printf("📍 Database path: %s\n", dbPath)

		// Check if database exists
		stat, err := os.Stat(dbPath)
		if os.IsNotExist(err) {
			fmt.Println("❌ Database does not exist")
			fmt.Println("💡 Run 'clai db reset' to create and populate the database")
			return nil
		}
		if err != nil {
			return fmt.Errorf("failed to check database file: %w", err)
		}

		// Show file size
		size := stat.Size()
		var sizeStr string
		if size < 1024 {
			sizeStr = fmt.Sprintf("%d B", size)
		} else if size < 1024*1024 {
			sizeStr = fmt.Sprintf("%.1f KB", float64(size)/1024)
		} else {
			sizeStr = fmt.Sprintf("%.1f MB", float64(size)/(1024*1024))
		}

		fmt.Printf("📊 Database size: %s\n", sizeStr)
		fmt.Printf("🕐 Last modified: %s\n", stat.ModTime().Format("2006-01-02 15:04:05"))

		// Try to connect and get statistics
		db, err := vecdb.NewVecDB()
		if err != nil {
			fmt.Printf("⚠️  Could not connect to database: %v\n", err)
			return nil
		}
		defer db.Close()

		// Get man page count
		count, err := db.GetManPageCount()
		if err != nil {
			fmt.Printf("⚠️  Could not get man page count: %v\n", err)
		} else {
			fmt.Printf("📚 Indexed man pages: %d\n", count)
		}

		// Get database schema info
		tables, err := db.GetTableNames()
		if err != nil {
			fmt.Printf("⚠️  Could not get table info: %v\n", err)
		} else {
			fmt.Println("🗂️  Database tables:")
			for _, table := range tables {
				fmt.Printf("   • %s\n", table)
			}
		}

		fmt.Println("✅ Database is healthy and accessible")

		return nil
	},
}

// getDBPath returns the database file path
func getDBPath() (string, error) {
	appDataDir, err := getAppDataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(appDataDir, "db", "manpages.db"), nil
}

// getAppDataDir returns the application data directory
func getAppDataDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	var appDataDir string
	switch {
	case os.Getenv("XDG_DATA_HOME") != "":
		appDataDir = filepath.Join(os.Getenv("XDG_DATA_HOME"), "clai")
	case fileExists(filepath.Join(homeDir, ".local", "share")):
		appDataDir = filepath.Join(homeDir, ".local", "share", "clai")
	default:
		appDataDir = filepath.Join(homeDir, "Library", "Application Support", "Clai")
	}

	return appDataDir, nil
}

// fileExists checks if a file or directory exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func init() {
	dbCmd.AddCommand(dbResetCmd)
	dbCmd.AddCommand(dbShowCmd)
	rootCmd.AddCommand(dbCmd)
}
