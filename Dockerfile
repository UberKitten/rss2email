#
# Dockerfile for rss2email.
#
# Build:
#   docker build -t rss2email:latest .
#
# Run with config file (recommended):
#   docker run -d \
#        -v /path/to/config:/app/.rss2email \
#        rss2email:latest daemon user@example.com
#
# The config directory should contain:
#   - feeds.txt    (feed list)
#   - config.yaml  (SMTP settings)
#   - state.db     (created automatically)
#
# Legacy: env vars still work as fallback:
#   docker run -d \
#        --env SMTP_HOST=smtp.gmail.com \
#        --env SMTP_USERNAME=user@example.com \
#        --env SMTP_PASSWORD=secret \
#        rss2email:latest daemon user@example.com
#

# STEP1 - Build-image
###########################################################################
FROM golang:alpine AS builder

ARG VERSION

LABEL org.opencontainers.image.source=https://github.com/UberKitten/rss2email

# Create a working-directory
WORKDIR $GOPATH/src/github.com/skx/rss2email/

# Copy the source to it
COPY . .

# Build the binary - ensuring we pass the build-argument
RUN go build -ldflags "-X main.version=$VERSION" -o /go/bin/rss2email

# STEP2 - Deploy-image
###########################################################################
FROM alpine

# Install CA certificates for HTTPS feed fetching and TLS SMTP
RUN apk add --no-cache ca-certificates

# Copy the binary.
COPY --from=builder /go/bin/rss2email /usr/local/bin/

# Set entrypoint
ENTRYPOINT [ "/usr/local/bin/rss2email" ]

# Set default command
CMD help

# Create a group and user
RUN addgroup app && adduser -D -G app -h /app app

# Tell docker that all future commands should run as the app user
USER app

# Set working directory and HOME so state goes to /app/.rss2email
WORKDIR /app
ENV HOME=/app
