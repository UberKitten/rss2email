# RSS2Email

A self-hosted RSS/Atom feed reader that delivers new posts to your inbox via email. Fork of [skx/rss2email](https://github.com/skx/rss2email) (archived) with active maintenance and improvements.

## What's Different From Upstream

- **YAML config file** for SMTP settings (no more env var juggling)
- **`from` per-feed option** for custom sender addresses
- **Category filtering** (`include-category` / `exclude-category`)
- **New CLI commands**: `status`, `test`, `check`
- **Improved `seen`** with pattern filtering and `--count` mode
- **Cleaner Dockerfile** (no sendmail/msmtp wrapper needed)
- GitHub Actions CI

## Quick Start

### Install

```bash
go install github.com/skx/rss2email@latest
```

### Docker

```bash
docker run -d \
  -v /path/to/config:/app/.rss2email \
  ghcr.io/uberkitten/rss2email:latest \
  daemon user@example.com
```

### Configure SMTP

Create `~/.rss2email/config.yaml`:

```yaml
smtp:
  host: smtp.example.com
  port: 587
  username: user@example.com
  password: your-password
from: notifications@example.com
```

Verify it works:

```bash
rss2email test user@example.com
```

> **Env var fallback**: `SMTP_HOST`, `SMTP_PORT`, `SMTP_USERNAME`, `SMTP_PASSWORD`, and `FROM` still work. Config file values take precedence.

### Add Feeds

```bash
rss2email add https://blog.example.com/feed.xml
```

Or edit `~/.rss2email/feeds.txt` directly:

```
https://blog.example.com/feed.xml

https://news.example.com/rss
 - tag:news
 - from:news@example.com
```

### Run

```bash
# One-shot (for cron)
rss2email cron user@example.com

# Daemon mode (polls every 5 minutes)
rss2email daemon user@example.com

# Dry run (don't actually send)
rss2email cron -send=false user@example.com
```

## Commands

| Command | Description |
|---------|-------------|
| `add <url>` | Add a feed |
| `delete <url>` | Remove a feed |
| `list` | List all configured feeds |
| `check <url>` | Validate a feed URL is reachable |
| `check --all` | Validate all configured feeds |
| `status` | Show config, SMTP, and state overview |
| `test <email>` | Send a test email |
| `cron <email>` | Process all feeds, send emails |
| `daemon <email>` | Run continuously (5-min poll interval) |
| `seen [pattern]` | Show seen items (optionally filtered) |
| `seen --count` | Show item counts per feed |
| `unsee <url>` | Mark an item as unseen (triggers re-send) |
| `config` | Show configuration documentation |
| `import <file>` | Import feeds from OPML |
| `export` | Export feeds as OPML |

## Per-Feed Options

Options go below the feed URL in `feeds.txt`, prefixed with ` - `:

```
https://example.com/feed.xml
 - from:custom-sender@example.com
 - tag:tech
 - exclude-title:(?i)sponsored
 - include-category:golang
 - frequency:60
 - notify:special@example.com
```

| Option | Description |
|--------|-------------|
| `from` | Custom sender address for this feed |
| `tag` | Tag added to email subject: `[rss2email] [tag] Title` |
| `exclude` | Skip items matching regex (body) |
| `exclude-title` | Skip items matching regex (title) |
| `exclude-category` | Skip items with category matching regex |
| `exclude-older` | Skip items older than N days |
| `include` | Only include items matching regex (body) |
| `include-title` | Only include items matching regex (title) |
| `include-category` | Only include items with category matching regex |
| `notify` | Override recipient list (comma-separated) |
| `frequency` | Minimum minutes between fetches |
| `template` | Custom email template file |
| `sleep` | Seconds to wait before fetching |
| `retry` | Max retry attempts for failed fetches |
| `delay` | Seconds between retries |
| `user-agent` | Custom User-Agent header |
| `insecure` | Ignore TLS errors (`true`/`yes`) |

## Email Customization

The default email template can be overridden by placing a file at `~/.rss2email/email.tmpl`. Per-feed templates are supported via the `template` option.

See the default template:

```bash
rss2email list-default-template
```

### Template Variables

| Variable | Description |
|----------|-------------|
| `{{.Feed}}` | Feed URL |
| `{{.FeedTitle}}` | Feed title |
| `{{.From}}` | From header (display name + address) |
| `{{.FromAddr}}` | Just the email address |
| `{{.Link}}` | Item URL |
| `{{.Subject}}` | Item title |
| `{{.To}}` | Recipient address |
| `{{.Tag}}` | Feed tag |
| `{{.Text}}` | Plain text content |
| `{{.HTML}}` | HTML content |
| `{{.RSSFeed}}` | Full feed object |
| `{{.RSSItem}}` | Full item object (e.g. `{{.RSSItem.GUID}}`) |

### Template Functions

| Function | Example |
|----------|---------|
| `env` | `{{env "USER"}}` |
| `quoteprintable` | `{{quoteprintable .Link}}` |
| `encodeHeader` | `{{encodeHeader .Subject}}` |
| `split` | `{{split "a:b" ":"}}` |
| `makeListIdHeader` | `{{makeListIdHeader .Feed}}` |

## Logging

Set `LOG_LEVEL` to `DEBUG`, `WARN`, or `ERROR`:

```bash
LOG_LEVEL=DEBUG rss2email cron user@example.com
```

Or use `-verbose` flag on `cron`/`daemon` commands.

Logs go to stderr and optionally to a file (`rss2email.log` by default, override with `LOG_FILE_PATH`).

## State

State is stored in `~/.rss2email/state.db` (BoltDB). Each feed gets a bucket, and seen item URLs are stored as keys.

When a feed item falls out of the remote feed, it's automatically pruned from state. If a feed is removed from `feeds.txt`, its bucket is pruned on next run.

## License

[MIT](LICENSE)

---

Originally by [Steve Kemp](https://steve.kemp.fi/). Fork maintained by [UberKitten](https://github.com/UberKitten).
