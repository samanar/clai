package cmd

import (
	"fmt"
	"os"
	"sync"
	"sync/atomic"

	"github.com/samanar/clai/model"
	"github.com/samanar/clai/vecdb"
	"github.com/spf13/cobra"
)

var dbEmbedCmd = &cobra.Command{
	Use:   "embed",
	Short: "Generate embeddings for man pages",
	Long:  "Generate vector embeddings for all man pages to enable semantic search",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("🧠 Generating embeddings for man pages...")

		// Configuration
		numWorkers := 4
		batchSize := 100

		// Get model
		m, err := model.NewModel()
		if err != nil {
			fmt.Printf("failed to create model: %w", err)
			os.Exit(1)
		}

		// Ensure assets are available
		if err := m.EnsureAssets(); err != nil {
			fmt.Printf("failed to ensure assets: %w", err)
			os.Exit(1)
		}

		// Open database
		db, err := vecdb.NewVecDB()
		if err != nil {
			fmt.Printf("failed to open database: %w", err)
			os.Exit(1)
		}
		defer db.Close()

		// Statistics
		var totalProcessed atomic.Int64
		var totalErrors atomic.Int64

		// Process in batches
		for {
			// Get batch of man pages without embeddings
			pages, err := db.GetManPagesWithoutEmbeddings(batchSize)
			if err != nil {
				fmt.Printf("failed to get man pages: %w", err)
				os.Exit(1)
			}

			if len(pages) == 0 {
				break // All done
			}

			fmt.Printf("📝 Processing batch of %d pages with %d workers...\n", len(pages), numWorkers)

			// Create job channel and results channel
			type job struct {
				id          int
				name        string
				description string
				content     string
			}

			type result struct {
				id        int
				embedding []float32
				err       error
			}

			jobs := make(chan job, len(pages))
			results := make(chan result, len(pages))

			// Start worker pool
			var wg sync.WaitGroup
			for w := 0; w < numWorkers; w++ {
				wg.Add(1)
				go func(workerID int) {
					defer wg.Done()
					for j := range jobs {
						// Create summary for embedding
						summary := fmt.Sprintf("%s: %s\n\n%s",
							j.name,
							j.description,
							vecdb.ExtractKeyContent(j.content, 6000))

						// Generate embedding using Model
						embedding, err := m.GenerateEmbedding(summary)
						results <- result{
							id:        j.id,
							embedding: embedding,
							err:       err,
						}
					}
				}(w)
			}

			// Send jobs
			for _, page := range pages {
				jobs <- job{
					id:          page.ID,
					name:        page.Name,
					description: page.Description,
					content:     page.Content,
				}
			}
			close(jobs)

			// Collect results in a separate goroutine
			go func() {
				wg.Wait()
				close(results)
			}()

			// Process results
			processed := 0
			for result := range results {
				if result.err != nil {
					fmt.Printf("⚠️  Failed to generate embedding: %v\n", result.err)
					totalErrors.Add(1)
					continue
				}

				// Update database
				if err := db.UpdateEmbedding(result.id, result.embedding); err != nil {
					fmt.Printf("⚠️  Failed to save embedding: %v\n", err)
					totalErrors.Add(1)
					continue
				}

				processed++
				totalProcessed.Add(1)
				if processed%10 == 0 {
					fmt.Printf("   Processed %d/%d in batch (Total: %d)\n", processed, len(pages), totalProcessed.Load())
				}
			}

			fmt.Printf("✅ Batch complete. Processed: %d, Errors: %d\n", processed, totalErrors.Load())
		}

		fmt.Printf("🎉 Embedding generation complete! Total processed: %d, Total errors: %d\n", totalProcessed.Load(), totalErrors.Load())

		return nil
	},
}

func init() {
	dbCmd.AddCommand(dbEmbedCmd)
}
