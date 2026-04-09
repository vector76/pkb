# --- Build stage: compile pkb and raymond ---
FROM golang:1.25-alpine AS build

RUN apk add --no-cache git

# pkb: clone and build from source (module path not yet updated in
# the latest published release, so go install @latest won't work).
RUN git clone --depth 1 https://github.com/vector76/pkb.git /src/pkb
WORKDIR /src/pkb
RUN CGO_ENABLED=0 go build -o /go/bin/pkb .

# raymond: clone and build from source.
RUN git clone --depth 1 https://github.com/vector76/raymond.git /src/raymond
WORKDIR /src/raymond
RUN CGO_ENABLED=0 go build -o /go/bin/raymond ./cmd/raymond

# --- Runtime stage ---
FROM node:22-alpine

RUN apk add --no-cache git bash

# Install Claude Code CLI.
RUN npm install -g @anthropic-ai/claude-code

# Create a non-root user for running services.
RUN adduser -D -h /home/pkb pkb

# Copy Go binaries from build stage.
COPY --from=build /go/bin/pkb /usr/local/bin/pkb
COPY --from=build /go/bin/raymond /usr/local/bin/raymond

USER pkb
WORKDIR /kb

EXPOSE 4242

CMD ["pkb", "-C", "/kb", "-addr", "0.0.0.0:4242"]
