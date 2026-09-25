GO       ?= go
BUN      ?= bun
GOFLAGS  ?=
LDFLAGS  ?= -s -w
BUILDDIR ?= build
BINARY   ?= pab-go

.PHONY: all build frontend backend clean

all: build

build: frontend backend

frontend:
	cd frontend && $(BUN) install --frozen-lockfile
	cd frontend && $(BUN) run build

backend:
	@mkdir -p $(BUILDDIR)
	CGO_ENABLED=0 $(GO) build $(GOFLAGS) -trimpath -tags=nomsgpack -ldflags="$(LDFLAGS)" -o $(BUILDDIR)/$(BINARY) ./cmd/pab-go

clean:
	rm -rf $(BUILDDIR) internal/web/assets
