package emailer

import (
	"errors"
	"strings"
	"testing"
)

func TestMakeListIdHeader(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "http://example.com/feed",
			expected: "example.com.feed.localhost",
		},
		{
			input:    "https://example.com/feed",
			expected: "example.com.feed.localhost",
		},
		{
			input:    "https://example.com/foo/bar?baz=qux",
			expected: "example.com.foo.bar.baz=qux.localhost",
		},
		{
			input:    "example.com/feed",
			expected: "example.com.feed.localhost",
		},
		{
			input:    "https://example.com/foo//bar///baz",
			expected: "example.com.foo.bar.baz.localhost",
		},
		{
			input:    "https://example.com/foo@bar#baz",
			expected: "example.com.foo.bar.baz.localhost",
		},
	}

	for _, tt := range tests {
		got := makeListIdHeader(tt.input)
		if got != tt.expected {
			t.Errorf("makeListIdHeader(%q) = %q; want %q", tt.input, got, tt.expected)
		}
		// Check that every character in got is allowed
		// from https://datatracker.ietf.org/doc/html/rfc2822#section-3.2.4
		allowed := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!#$%&'*+-/=?^_`{|}~."
		for i, c := range got {
			if !strings.ContainsRune(allowed, c) {
				t.Errorf("makeListIdHeader(%q) produced invalid char %q at position %d", tt.input, c, i)
			}
		}
	}
}

func TestIsTransientError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"nil error", nil, false},
		{"permanent error", errors.New("550 mailbox not found"), false},
		{"random error", errors.New("connection refused"), false},
		{"azure rate limit", errors.New("450 4.5.127 Message rejected. Excessive message rate from sender."), true},
		{"generic rate limit", errors.New("rate limit exceeded"), true},
		{"throttled", errors.New("request throttled"), true},
		{"too many requests", errors.New("too many requests"), true},
		{"try again later", errors.New("try again later"), true},
		{"temporary failure", errors.New("temporary failure"), true},
		{"smtp 4xx start", errors.New("421 service not available"), true},
		{"office365 throttle", errors.New("4.7.427 throttled"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isTransientError(tt.err)
			if got != tt.expected {
				t.Errorf("isTransientError(%v) = %v; want %v", tt.err, got, tt.expected)
			}
		})
	}
}
