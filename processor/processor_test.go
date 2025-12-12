package processor

import (
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/skx/rss2email/configfile"
)

var (
	// logger contains a shared logging handle, the code we're testing assumes it exists.
	logger *slog.Logger
)

// init runs at test-time.
func init() {

	// setup logging-level
	lvl := &slog.LevelVar{}
	lvl.Set(slog.LevelWarn)

	// create a handler
	opts := &slog.HandlerOptions{Level: lvl}
	handler := slog.NewTextHandler(os.Stderr, opts)

	// ensure the global-variable is set.
	logger = slog.New(handler)
}

func TestSendEmail(t *testing.T) {

	p, err := New()

	if err != nil {
		t.Fatalf("error creating processor %s", err.Error())
	}
	defer p.Close()

	if !p.send {
		t.Fatalf("unexpected default to sending mail")
	}

	p.SetSendEmail(false)

	if p.send {
		t.Fatalf("unexpected send-setting")
	}

}

func TestVerbose(t *testing.T) {

	p, err := New()

	if err != nil {
		t.Fatalf("error creating processor %s", err.Error())
	}

	defer p.Close()
}

// TestSkipExclude ensures that we can exclude items by regexp
func TestSkipExclude(t *testing.T) {

	feed := configfile.Feed{
		URL: "blah",
		Options: []configfile.Option{
			{Name: "exclude", Value: "foo"},
			{Name: "exclude-title", Value: "test"},
		},
	}

	// Create the new processor
	x, err := New()

	if err != nil {
		t.Fatalf("error creating processor %s", err.Error())
	}
	defer x.Close()

	if !x.shouldSkip(logger, feed, "Title here", "<p>foo, bar baz</p>") {
		t.Fatalf("failed to skip entry by regexp")
	}

	if !x.shouldSkip(logger, feed, "test", "<p>This matches the title</p>") {
		t.Fatalf("failed to skip entry by title")
	}

	// With no options we're not going to skip
	feed = configfile.Feed{
		URL:     "blah",
		Options: []configfile.Option{},
	}

	if x.shouldSkip(logger, feed, "Title here", "<p>foo, bar baz</p>") {
		t.Fatalf("skipped something with no options!")
	}

}

// TestSkipInclude ensures that we can exclude items by regexp
func TestSkipInclude(t *testing.T) {

	feed := configfile.Feed{
		URL: "blah",
		Options: []configfile.Option{
			{Name: "include", Value: "good"},
		},
	}

	// Create the new processor
	x, err := New()

	if err != nil {
		t.Fatalf("error creating processor %s", err.Error())
	}
	defer x.Close()

	if x.shouldSkip(logger, feed, "Title here", "<p>This is good</p>") {
		t.Fatalf("this should be included because it contains good")
	}

	if !x.shouldSkip(logger, feed, "Title here", "<p>This should be excluded.</p>") {
		t.Fatalf("This should be excluded; doesn't contain 'good'")
	}

	// If we don't try to make a mandatory include setting
	// nothing should be skipped
	feed = configfile.Feed{
		URL:     "blah",
		Options: []configfile.Option{},
	}

	if x.shouldSkip(logger, feed, "Title here", "<p>This is good</p>") {
		t.Fatalf("nothing specified, shouldn't be skipped")
	}
}

// TestSkipIncludeTitle ensures that we can exclude items by regexp
func TestSkipIncludeTitle(t *testing.T) {

	feed := configfile.Feed{
		URL: "blah",
		Options: []configfile.Option{
			{Name: "include", Value: "good"},
			{Name: "include-title", Value: "(?i)cake"},
		},
	}

	// Create the new processor
	x, err := New()
	if err != nil {
		t.Fatalf("error creating processor %s", err.Error())
	}

	if x.shouldSkip(logger, feed, "Title here", "<p>This is good</p>") {
		t.Fatalf("this should be included because it contains good")
	}
	if x.shouldSkip(logger, feed, "I like Cake!", "<p>Food is good.</p>") {
		t.Fatalf("this should be included because of the title")
	}

	//
	// Second test, only include titles
	//
	feed = configfile.Feed{
		URL: "blah",
		Options: []configfile.Option{
			{Name: "include-title", Value: "(?i)cake"},
			{Name: "include-title", Value: "(?i)pie"},
		},
	}

	//
	// Some titles which are OK
	//
	valid := []string{"I like cake", "I like pie", "piecemeal", "cupcake", "pancake"}
	bogus := []string{"I do not like food", "I don't like cooked goods", "cheese is dead milk", "books are fun", "tv is good"}

	// Create the new processor
	x.Close()
	x, err = New()
	if err != nil {
		t.Fatalf("error creating processor %s", err.Error())
	}
	defer x.Close()

	// include
	for _, entry := range valid {
		if x.shouldSkip(logger, feed, entry, "content") {
			t.Fatalf("this should be included due to include-title")
		}
	}

	// exclude
	for _, entry := range bogus {
		if !x.shouldSkip(logger, feed, entry, "content") {
			t.Fatalf("this shouldn't be included!")
		}
	}
}

// TestSkipOlder ensures that we can exclude items by age
func TestSkipOlder(t *testing.T) {

	feed := configfile.Feed{
		URL: "blah",
		Options: []configfile.Option{
			{Name: "exclude-older", Value: "1"},
		},
	}

	// Create the new processor
	x, err := New()

	if err != nil {
		t.Fatalf("error creating processor %s", err.Error())
	}
	defer x.Close()

	if x.shouldSkipOlder(logger, feed, "X") {
		t.Fatalf("failed to skip non correct published-date")
	}

	if !x.shouldSkipOlder(logger, feed, "Fri, 02 Dec 2022 16:43:04 +0000") {
		t.Fatalf("failed to skip old entry by age")
	}

	if !x.shouldSkipOlder(logger, feed, time.Now().Add(-time.Hour*24*2).Format(time.RFC1123)) {
		t.Fatalf("failed to skip newer entry by age")
	}

	if x.shouldSkipOlder(logger, feed, time.Now().Add(-time.Hour*12).Format(time.RFC1123)) {
		t.Fatalf("skipped new entry by age")
	}

	// With no options we're not going to skip
	feed = configfile.Feed{
		URL:     "blah",
		Options: []configfile.Option{},
	}

	if x.shouldSkipOlder(logger, feed, time.Now().Add(-time.Hour*24*128).String()) {
		t.Fatalf("skipped age with no options!")
	}
}

// TestSkipExcludeCategory ensures that we can exclude items by category regexp
func TestSkipExcludeCategory(t *testing.T) {

	feed := configfile.Feed{
		URL: "blah",
		Options: []configfile.Option{
			{Name: "exclude-category", Value: "(?i)sports"},
		},
	}

	// Create the new processor
	x, err := New()
	if err != nil {
		t.Fatalf("error creating processor %s", err.Error())
	}
	defer x.Close()

	// Should skip because "Sports" matches "(?i)sports"
	if !x.shouldSkipCategory(logger, feed, []string{"News", "Sports", "Entertainment"}) {
		t.Fatalf("failed to skip entry by category regexp")
	}

	// Should not skip because no category matches "(?i)sports"
	if x.shouldSkipCategory(logger, feed, []string{"News", "Entertainment"}) {
		t.Fatalf("skipped entry that doesn't match category regexp")
	}

	// Empty categories should not be skipped
	if x.shouldSkipCategory(logger, feed, []string{}) {
		t.Fatalf("skipped entry with empty categories")
	}

	// With no options we're not going to skip
	feed = configfile.Feed{
		URL:     "blah",
		Options: []configfile.Option{},
	}

	if x.shouldSkipCategory(logger, feed, []string{"Sports", "News"}) {
		t.Fatalf("skipped something with no options!")
	}
}

// TestSkipIncludeCategory ensures that we can include items by category regexp
func TestSkipIncludeCategory(t *testing.T) {

	feed := configfile.Feed{
		URL: "blah",
		Options: []configfile.Option{
			{Name: "include-category", Value: "(?i)tech"},
		},
	}

	// Create the new processor
	x, err := New()
	if err != nil {
		t.Fatalf("error creating processor %s", err.Error())
	}
	defer x.Close()

	// Should not skip because "Technology" matches "(?i)tech"
	if x.shouldSkipCategory(logger, feed, []string{"Technology", "News"}) {
		t.Fatalf("skipped entry that should be included by category")
	}

	// Should skip because no category matches "(?i)tech"
	if !x.shouldSkipCategory(logger, feed, []string{"Sports", "Entertainment"}) {
		t.Fatalf("failed to skip entry that doesn't match include-category")
	}

	// With no options we're not going to skip
	feed = configfile.Feed{
		URL:     "blah",
		Options: []configfile.Option{},
	}

	if x.shouldSkipCategory(logger, feed, []string{"Sports", "News"}) {
		t.Fatalf("skipped something with no options!")
	}
}

// TestSkipMultipleIncludeCategory ensures that multiple include-category options work
func TestSkipMultipleIncludeCategory(t *testing.T) {

	feed := configfile.Feed{
		URL: "blah",
		Options: []configfile.Option{
			{Name: "include-category", Value: "(?i)tech"},
			{Name: "include-category", Value: "(?i)programming"},
		},
	}

	// Create the new processor
	x, err := New()
	if err != nil {
		t.Fatalf("error creating processor %s", err.Error())
	}
	defer x.Close()

	// Should not skip because "Programming" matches second include-category
	if x.shouldSkipCategory(logger, feed, []string{"Programming"}) {
		t.Fatalf("skipped entry that should be included by second include-category")
	}

	// Should not skip because "Technology" matches first include-category
	if x.shouldSkipCategory(logger, feed, []string{"Technology"}) {
		t.Fatalf("skipped entry that should be included by first include-category")
	}

	// Should skip because no category matches any include-category
	if !x.shouldSkipCategory(logger, feed, []string{"Sports", "Entertainment"}) {
		t.Fatalf("failed to skip entry that doesn't match any include-category")
	}
}

// TestSkipInvalidCategoryRegex ensures that invalid regex patterns don't cause panics
func TestSkipInvalidCategoryRegex(t *testing.T) {

	// Test with invalid regex in exclude-category
	feed := configfile.Feed{
		URL: "blah",
		Options: []configfile.Option{
			{Name: "exclude-category", Value: "[invalid"},
		},
	}

	// Create the new processor
	x, err := New()
	if err != nil {
		t.Fatalf("error creating processor %s", err.Error())
	}
	defer x.Close()

	// Should not panic and should not skip (invalid regex is logged as warning)
	if x.shouldSkipCategory(logger, feed, []string{"Sports", "Entertainment"}) {
		t.Fatalf("skipped entry with invalid regex pattern")
	}

	// Test with invalid regex in include-category
	feed = configfile.Feed{
		URL: "blah",
		Options: []configfile.Option{
			{Name: "include-category", Value: "[invalid"},
		},
	}

	// Should skip because include-category was specified but none matched
	// (invalid regex fails to match)
	if !x.shouldSkipCategory(logger, feed, []string{"Sports"}) {
		t.Fatalf("failed to skip entry when include-category has invalid regex")
	}
}
