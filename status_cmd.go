//
// Show status information about feeds and state.
//

package main

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/skx/rss2email/config"
	"github.com/skx/rss2email/configfile"
	"github.com/skx/rss2email/state"
	"github.com/skx/subcommands"
	"go.etcd.io/bbolt"
)

// Structure for our options and state.
type statusCmd struct {

	// We embed the NoFlags option, because we accept no command-line flags.
	subcommands.NoFlags
}

// Info is part of the subcommand-API.
func (s *statusCmd) Info() (string, string) {
	return "status", `Show status information about feeds and configuration.

This sub-command displays an overview of your rss2email setup:

  - Configuration file location and feed count
  - SMTP delivery configuration (config file vs env vars)
  - State database statistics (items per feed)

Example:

    $ rss2email status
`
}

// Entry-point.
func (s *statusCmd) Execute(args []string) int {

	// Show config location
	conf := configfile.New()
	fmt.Printf("Config file: %s\n", conf.Path())

	// Parse feeds
	entries, err := conf.Parse()
	if err != nil {
		fmt.Printf("  Error parsing config: %s\n", err.Error())
		return 1
	}
	fmt.Printf("Feeds:       %d\n", len(entries))

	// Show SMTP config
	cfg, cfgErr := config.Load()
	if cfgErr != nil {
		fmt.Printf("\nSMTP:        not configured (error loading config: %s)\n", cfgErr.Error())
	} else if cfg.HasSMTP() {
		fmt.Printf("\nSMTP:\n")
		fmt.Printf("  Host:      %s:%d\n", cfg.SMTP.Host, cfg.SMTP.Port)
		fmt.Printf("  Username:  %s\n", cfg.SMTP.Username)
		fmt.Printf("  Password:  %s\n", strings.Repeat("*", len(cfg.SMTP.Password)))
		if cfg.From != "" {
			fmt.Printf("  From:      %s\n", cfg.From)
		}

		// Show source
		configPath := config.Path()
		if _, statErr := os.Stat(configPath); statErr == nil {
			fmt.Printf("  Source:    %s\n", configPath)
		} else {
			fmt.Printf("  Source:    environment variables\n")
		}
	} else {
		fmt.Printf("\nSMTP:        not configured (will use /usr/sbin/sendmail)\n")
	}

	// Show state database
	dir := state.Directory()
	dbPath := filepath.Join(dir, "state.db")

	if _, statErr := os.Stat(dbPath); os.IsNotExist(statErr) {
		fmt.Printf("\nState DB:    not found (no feeds processed yet)\n")
		return 0
	}

	db, err := bbolt.Open(dbPath, 0666, nil)
	if err != nil {
		logger.Error("failed to open database", slog.String("database", dbPath), slog.String("error", err.Error()))
		return 1
	}
	defer db.Close()

	// Collect bucket stats
	type feedStat struct {
		name  string
		count int
	}
	var stats []feedStat
	totalItems := 0

	err = db.View(func(tx *bbolt.Tx) error {
		return tx.ForEach(func(bucketName []byte, b *bbolt.Bucket) error {
			count := 0
			c := b.Cursor()
			for k, _ := c.First(); k != nil; k, _ = c.Next() {
				count++
			}
			stats = append(stats, feedStat{name: string(bucketName), count: count})
			totalItems += count
			return nil
		})
	})

	if err != nil {
		logger.Error("failed to read database", slog.String("error", err.Error()))
		return 1
	}

	// Sort by count descending
	sort.Slice(stats, func(i, j int) bool {
		return stats[i].count > stats[j].count
	})

	fmt.Printf("\nState DB:    %s\n", dbPath)
	fmt.Printf("Total seen:  %d items across %d feeds\n", totalItems, len(stats))

	if len(stats) > 0 {
		fmt.Printf("\nFeed                                                           Seen\n")
		fmt.Printf("%-60s %5s\n", strings.Repeat("-", 60), "-----")

		for _, st := range stats {
			name := st.name
			if len(name) > 60 {
				name = name[:57] + "..."
			}
			fmt.Printf("%-60s %5d\n", name, st.count)
		}
	}

	return 0
}
