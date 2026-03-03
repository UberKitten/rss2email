//
// Show feeds and their contents
//

package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/skx/rss2email/state"
	"go.etcd.io/bbolt"
)

// Structure for our options and state.
type seenCmd struct {
	// Show only counts, not individual items
	count bool
}

// Info is part of the subcommand-API.
func (s *seenCmd) Info() (string, string) {
	return "seen", `Show all the feed-items we've seen.

This sub-command will report upon all the feeds to which
you're subscribed, and show the link to each feed-entry
to which you've been notified in the past.

You can optionally filter by feed URL pattern (substring match).

Examples:

    $ rss2email seen
    $ rss2email seen example.com
    $ rss2email seen --count
    $ rss2email seen --count blog.example.com
`
}

// Arguments handles our flag-setup.
func (s *seenCmd) Arguments(f *flag.FlagSet) {
	f.BoolVar(&s.count, "count", false, "Show only the number of items per feed, not individual items")
}

// Entry-point.
func (s *seenCmd) Execute(args []string) int {

	// Optional pattern filter
	pattern := ""
	if len(args) > 0 {
		pattern = args[0]
	}

	// Ensure we have a state-directory.
	dir := state.Directory()
	errM := os.MkdirAll(dir, 0666)
	if errM != nil {
		logger.Error("failed to create directory", slog.String("directory", dir), slog.String("error", errM.Error()))
		return 1
	}

	// Now create the database, if missing, or open it if it exists.
	dbPath := filepath.Join(dir, "state.db")
	db, err := bbolt.Open(dbPath, 0666, nil)
	if err != nil {
		logger.Error("failed to open database", slog.String("database", dbPath), slog.String("error", err.Error()))
		return 1
	}

	// Ensure we close when we're done
	defer db.Close()

	// Keep track of buckets here
	var bucketNames [][]byte

	err = db.View(func(tx *bbolt.Tx) error {
		err = tx.ForEach(func(bucketName []byte, _ *bbolt.Bucket) error {
			bucketNames = append(bucketNames, bucketName)
			return nil
		})
		return err
	})
	if err != nil {
		logger.Error("failed to find bucket names", slog.String("database", dbPath), slog.String("error", err.Error()))
		return 1
	}

	matched := 0

	// Now we have a list of buckets, we'll show the contents
	for _, buck := range bucketNames {

		name := string(buck)

		// Apply pattern filter
		if pattern != "" && !strings.Contains(strings.ToLower(name), strings.ToLower(pattern)) {
			continue
		}

		matched++

		if s.count {
			// Count-only mode
			itemCount := 0
			err = db.View(func(tx *bbolt.Tx) error {
				b := tx.Bucket([]byte(buck))
				c := b.Cursor()
				for k, _ := c.First(); k != nil; k, _ = c.Next() {
					itemCount++
				}
				return nil
			})
			if err != nil {
				logger.Error("failed iterating over bucket", slog.String("database", dbPath), slog.String("bucket", name), slog.String("error", err.Error()))
				return 1
			}
			fmt.Printf("%s (%d items)\n", name, itemCount)
		} else {
			fmt.Printf("%s\n", name)

			err = db.View(func(tx *bbolt.Tx) error {
				b := tx.Bucket([]byte(buck))
				c := b.Cursor()
				for k, _ := c.First(); k != nil; k, _ = c.Next() {
					key := string(k)
					fmt.Printf("\t%s\n", key)
				}
				return nil
			})

			if err != nil {
				logger.Error("failed iterating over bucket", slog.String("database", dbPath), slog.String("bucket", name), slog.String("error", err.Error()))
				return 1
			}
		}
	}

	if pattern != "" && matched == 0 {
		fmt.Printf("No feeds matching '%s'\n", pattern)
	}

	return 0
}
