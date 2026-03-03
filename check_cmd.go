//
// Check if feed URLs are valid and reachable.
//

package main

import (
	"flag"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/mmcdole/gofeed"
	"github.com/skx/rss2email/configfile"
)

// Structure for our options and state.
type checkCmd struct {
	// Check all feeds from config
	all bool

	// Timeout for HTTP requests
	timeout int
}

// Info is part of the subcommand-API.
func (c *checkCmd) Info() (string, string) {
	return "check", `Check if feed URLs are valid and reachable.

This sub-command tests one or more feed URLs to verify they are
reachable and contain valid RSS/Atom content.

With --all it checks every feed in your configuration file.

Examples:

    $ rss2email check https://blog.example.com/feed.xml
    $ rss2email check --all
`
}

// Arguments handles our flag-setup.
func (c *checkCmd) Arguments(f *flag.FlagSet) {
	f.BoolVar(&c.all, "all", false, "Check all feeds from the configuration file")
	f.IntVar(&c.timeout, "timeout", 15, "HTTP timeout in seconds")
}

// Entry-point.
func (c *checkCmd) Execute(args []string) int {

	var urls []string

	if c.all {
		conf := configfile.New()
		entries, err := conf.Parse()
		if err != nil {
			fmt.Printf("Error parsing config: %s\n", err.Error())
			return 1
		}
		for _, entry := range entries {
			urls = append(urls, entry.URL)
		}
		fmt.Printf("Checking %d feeds from %s\n\n", len(urls), conf.Path())
	} else {
		if len(args) < 1 {
			fmt.Printf("Usage: rss2email check <url> [url...]\n")
			fmt.Printf("       rss2email check --all\n")
			return 1
		}
		urls = args
	}

	httpClient := &http.Client{
		Timeout: time.Duration(c.timeout) * time.Second,
	}

	parser := gofeed.NewParser()
	parser.Client = httpClient

	errors := 0
	for _, feedURL := range urls {
		result := c.checkFeed(parser, feedURL)
		if !result {
			errors++
		}
	}

	if errors > 0 {
		fmt.Printf("\n%d/%d feeds had errors\n", errors, len(urls))
		return 1
	}

	if len(urls) > 1 {
		fmt.Printf("\nAll %d feeds OK\n", len(urls))
	}
	return 0
}

func (c *checkCmd) checkFeed(parser *gofeed.Parser, feedURL string) bool {
	feed, err := parser.ParseURL(feedURL)
	if err != nil {
		errStr := err.Error()

		// Provide more human-friendly error messages
		if strings.Contains(errStr, "no such host") {
			fmt.Printf("FAIL %s\n     DNS resolution failed\n", feedURL)
		} else if strings.Contains(errStr, "connection refused") {
			fmt.Printf("FAIL %s\n     Connection refused\n", feedURL)
		} else if strings.Contains(errStr, "timeout") || strings.Contains(errStr, "deadline exceeded") {
			fmt.Printf("FAIL %s\n     Timed out after %ds\n", feedURL, c.timeout)
		} else if strings.Contains(errStr, "404") {
			fmt.Printf("FAIL %s\n     Not found (404)\n", feedURL)
		} else {
			fmt.Printf("FAIL %s\n     %s\n", feedURL, errStr)
		}
		return false
	}

	fmt.Printf("OK   %s\n     %s — %d items\n", feedURL, feed.Title, len(feed.Items))
	return true
}
