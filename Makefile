.PHONY: build test test-v lint format format-check pre-commit-install ci install-cli js-build js-test js-bundle-check e2e docs docs-build clean

BIN := bin/amarra-cais

test:
	go test ./... -race -count=1

test-v:
	go test ./... -v -count=1

build:
	CGO_ENABLED=0 go build -ldflags="-s -w" -o $(BIN) ./cmd/amarra-cais

lint:
	golangci-lint run ./...

format:
	npm run format

format-check:
	npm run format:check

js-build:
	npm run js:build

js-test:
	npm run js:test

e2e:
	npm run e2e

# Committed bundle (amarra.js) must match pkg/amarra/js sources.
js-bundle-check:
	npm run js:build
	git diff --exit-code pkg/cais/pwa/assets/amarra.js

pre-commit-install:
	pre-commit install

ci: test js-test lint format-check js-bundle-check

install-cli:
	go install ./cmd/amarra-cais

# Docs site (Astro Starlight) lives in website/.
docs:
	cd website && npm install && npm run dev

docs-build:
	cd website && npm ci && npm run build

clean:
	rm -rf bin/ tmp/
