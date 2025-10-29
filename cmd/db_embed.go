package cmd

import (
	"fmt"
	"os"
	"runtime"
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
		availableCPU := runtime.NumCPU()
		if availableCPU < 1 {
			availableCPU = 1
		}

		workersFlag, _ := cmd.Flags().GetInt("workers")
		numWorkers := workersFlag
		if numWorkers <= 0 {
			numWorkers = availableCPU / 2
			if numWorkers < 1 {
				numWorkers = 1
			}
		}

		dbWorkers := numWorkers / 2
		if dbWorkers < 1 {
			dbWorkers = 1
		}

		batchSizeFlag, _ := cmd.Flags().GetInt("batch-size")
		batchSize := batchSizeFlag
		if batchSize <= 0 {
			batchSize = numWorkers * 50
			if batchSize < 50 {
				batchSize = 50
			}
		}

		fmt.Printf("🧵 Using %d embedding workers and %d database workers (batch size %d)\n", numWorkers, dbWorkers, batchSize)

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
				name      string
				embedding []float32
				err       error
			}

			jobs := make(chan job, len(pages))
			results := make(chan result, len(pages))

			// Start worker pool
			var wg sync.WaitGroup
			for w := 0; w < numWorkers; w++ {
				wg.Add(1)
				go func() {
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
							name:      j.name,
							embedding: embedding,
							err:       err,
						}
					}
				}()
			}

			var batchProcessed atomic.Int64
			var batchErrors atomic.Int64

			var dbWg sync.WaitGroup
			for w := 0; w < dbWorkers; w++ {
				dbWg.Add(1)
				go func() {
					defer dbWg.Done()
					for res := range results {
						if res.err != nil {
							fmt.Printf("⚠️  Failed to generate embedding for %s: %v\n", res.name, res.err)
							totalErrors.Add(1)
							batchErrors.Add(1)
							continue
						}

						if err := db.UpdateEmbedding(res.id, res.embedding); err != nil {
							fmt.Printf("⚠️  Failed to save embedding for %s: %v\n", res.name, err)
							totalErrors.Add(1)
							batchErrors.Add(1)
							continue
						}

						processed := batchProcessed.Add(1)
						totalProcessed.Add(1)
						if processed%10 == 0 || processed == int64(len(pages)) {
							fmt.Printf("   Processed %d/%d in batch (Total: %d)\n", processed, len(pages), totalProcessed.Load())
						}
					}
				}()
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
			dbWg.Wait()

			fmt.Printf("✅ Batch complete. Processed: %d, Errors: %d\n", batchProcessed.Load(), batchErrors.Load())
		}

		fmt.Printf("🎉 Embedding generation complete! Total processed: %d, Total errors: %d\n", totalProcessed.Load(), totalErrors.Load())

		return nil
	},
}

func init() {
	dbEmbedCmd.Flags().Int("workers", 0, "Number of embedding workers to run concurrently (0 = auto)")
	dbEmbedCmd.Flags().Int("batch-size", 0, "Number of man pages fetched per batch (0 = auto)")
	dbCmd.AddCommand(dbEmbedCmd)
}
