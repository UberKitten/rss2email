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

## Runtime security maintenance

The maintained Dockerfile and CI pin supported Go `1.26.10`; the generic image
uses Alpine `3.24.3`. Rebuild from the exact reviewed fork revision, using the
committed modules and Go's source compatibility guarantee. Run the existing
`go build -v ./...` and `go test -v ./...` gates with an isolated home/config.
Do not invoke `daemon`, `cron`, `test`, or feed/network commands against real
configuration to verify a build.
The targeted `golang.org/x/net v0.55.0` security update fixes reachable HTML
parsing/rendering advisories; its required closure is `x/sys v0.45.0` and
`x/text v0.37.0`. These modules require Go 1.25 or later; the pinned compiler
remains 1.26.10. No feed, mail-delivery, or state-management feature is changed.

The established private-apps installation is a local image overlay, not the
generic GHCR image above. It retains the vendor's tini/entrypoint, msmtp setup,
UID/GID 1000, `/app/rss2email`, `/home/rss2email`, and existing sendmail wrapper.
Its runtime build context replaces only the static fork binary and updates
the supported vendor base/packages. Do not substitute the generic image's
`/app/.rss2email` layout or change the installation channel during maintenance.
The site-owned Compose/deployment contract remains authoritative for the
environment, mounts, poll interval, recipient configuration, and image pin.

Before the single-instance cutover, retain the old image and a root-only,
stopped-consistent checkpoint of the complete state/config bind, including the
BoltDB seen-item state, feed list, templates, SMTP configuration and existing
logs. Preserve the read-only feed-file mount. There must never be a second
poller or an intentional resend; acceptance uses binary build metadata, existing
service health/process state, and read-only state invariants without enumerating
feed URLs or sending mail. A runtime-only rebuild does not authorize adding
feeds, creating Azure senders, rule deployment, clearing seen state, or test SMTP.
Rollback is deliberate and operator-assisted with the retained image and
coupled checkpoint; never restore vulnerable code automatically or overwrite
newer seen-item/config state.

## License

[MIT](LICENSE)

---

Originally by [Steve Kemp](https://steve.kemp.fi/). Fork maintained by [UberKitten](https://github.com/UberKitten).
