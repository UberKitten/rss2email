//
// Send a test email to verify SMTP configuration.
//

package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net/smtp"
	"strings"
	"time"

	"github.com/skx/rss2email/config"
)

// Structure for our options and state.
type testCmd struct {
	// verbose output
	verbose bool
}

// Info is part of the subcommand-API.
func (t *testCmd) Info() (string, string) {
	return "test", `Send a test email to verify SMTP configuration.

This sub-command sends a simple test email to the specified address
to verify that your SMTP configuration is working correctly.

It reads SMTP settings from the config file (~/.rss2email/config.yaml)
with fallback to environment variables (SMTP_HOST, etc).

Example:

    $ rss2email test user@example.com
`
}

// Arguments handles our flag-setup.
func (t *testCmd) Arguments(f *flag.FlagSet) {
	f.BoolVar(&t.verbose, "verbose", false, "Show detailed connection information")
}

// Entry-point.
func (t *testCmd) Execute(args []string) int {

	if len(args) < 1 {
		fmt.Printf("Usage: rss2email test <email-address>\n")
		return 1
	}

	addr := args[0]
	if !strings.Contains(addr, "@") {
		fmt.Printf("Error: '%s' doesn't look like an email address\n", addr)
		return 1
	}

	// Load config
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Error loading config: %s\n", err.Error())
		return 1
	}

	// Validate SMTP config
	if !cfg.HasSMTP() {
		fmt.Printf("Error: SMTP is not configured.\n\n")
		fmt.Printf("Configure SMTP in %s:\n\n", config.Path())
		fmt.Printf("  smtp:\n")
		fmt.Printf("    host: smtp.example.com\n")
		fmt.Printf("    port: 587\n")
		fmt.Printf("    username: user@example.com\n")
		fmt.Printf("    password: your-password\n")
		fmt.Printf("\nOr set environment variables: SMTP_HOST, SMTP_PORT, SMTP_USERNAME, SMTP_PASSWORD\n")
		return 1
	}

	issues := cfg.Validate()
	if len(issues) > 0 {
		fmt.Printf("Configuration issues:\n")
		for _, issue := range issues {
			fmt.Printf("  - %s\n", issue)
		}
		return 1
	}

	// Build the test email
	from := addr
	if cfg.From != "" {
		from = cfg.From
	}

	now := time.Now().Format(time.RFC1123Z)
	subject := fmt.Sprintf("rss2email test message - %s", time.Now().Format("2006-01-02 15:04:05"))

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nDate: %s\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\nThis is a test message from rss2email.\n\nSMTP Host: %s:%d\nTimestamp: %s\n\nIf you received this, your SMTP configuration is working correctly.\n",
		from, addr, subject, now, cfg.SMTP.Host, cfg.SMTP.Port, now)

	// Send it
	smtpAddr := fmt.Sprintf("%s:%d", cfg.SMTP.Host, cfg.SMTP.Port)

	if t.verbose {
		fmt.Printf("Connecting to %s...\n", smtpAddr)
		fmt.Printf("  Username: %s\n", cfg.SMTP.Username)
		fmt.Printf("  From:     %s\n", from)
		fmt.Printf("  To:       %s\n", addr)
	}

	auth := smtp.PlainAuth("", cfg.SMTP.Username, cfg.SMTP.Password, cfg.SMTP.Host)
	err = smtp.SendMail(smtpAddr, auth, from, []string{addr}, []byte(msg))
	if err != nil {
		logger.Error("failed to send test email",
			slog.String("to", addr),
			slog.String("smtp", smtpAddr),
			slog.String("error", err.Error()))
		fmt.Printf("Error: failed to send email: %s\n", err.Error())
		return 1
	}

	fmt.Printf("Test email sent successfully to %s\n", addr)
	return 0
}
