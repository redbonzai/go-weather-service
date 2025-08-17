# ---- Config ----
APP_NAME        ?= weather-service
PKG_MAIN        ?= ./cmd/server
BIN_DIR         ?= bin
BIN_PATH        ?= $(BIN_DIR)/$(APP_NAME)

GO_VERSION      ?= 1.24
GOFLAGS         ?=
LDFLAGS         ?= -s -w
CGO_ENABLED     ?= 0

# Docker
DOCKERFILE      ?= Dockerfile
IMAGE_REGISTRY  ?= ghcr.io
IMAGE_OWNER     ?= $(shell git config --get user.username 2>/dev/null || echo your-org)
IMAGE_NAME      ?= $(IMAGE_REGISTRY)/$(IMAGE_OWNER)/go-weather-service
IMAGE_TAG       ?= $(shell git rev-parse --short=8 HEAD)
PLATFORMS       ?= linux/amd64
BUILD_ARGS      ?= --build-arg GO_VERSION=$(GO_VERSION)

# Runtime defaults (override at run time)
NWS_USER_AGENT  ?= go-weather-service/1.0 (dev@example.com)
COLD_LT_F       ?= 50
HOT_GE_F        ?= 80
PORT            ?= 8080

# ---- Tasks ----
.PHONY: all
all: tidy vet test build

$(BIN_DIR):
	mkdir -p $(BIN_DIR)

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: vet
vet:
	go vet ./...

.PHONY: test
test:
	go test -race -count=1 ./...

.PHONY: build
build: $(BIN_DIR)
	CGO_ENABLED=$(CGO_ENABLED) go build -trimpath -ldflags="$(LDFLAGS)" -o $(BIN_PATH) $(PKG_MAIN)

.PHONY: run
run:
	NWS_USER_AGENT="$(NWS_USER_AGENT)" \
	COLD_LT_F="$(COLD_LT_F)" \
	HOT_GE_F="$(HOT_GE_F)" \
	go run $(PKG_MAIN)

.PHONY: docker-build
docker-build:
	docker build $(BUILD_ARGS) -f $(DOCKERFILE) -t $(IMAGE_NAME):$(IMAGE_TAG) .

.PHONY: docker-run
docker-run:
	docker run --rm -p $(PORT):8080 \
		-e NWS_USER_AGENT="$(NWS_USER_AGENT)" \
		-e COLD_LT_F="$(COLD_LT_F)" \
		-e HOT_GE_F="$(HOT_GE_F)" \
		$(IMAGE_NAME):$(IMAGE_TAG)

# Push only if you've logged in: `echo $GITHUB_TOKEN | docker login ghcr.io -u USERNAME --password-stdin`
.PHONY: docker-push
docker-push:
	docker push $(IMAGE_NAME):$(IMAGE_TAG)

# Optional: buildx multi-arch
.PHONY: docker-buildx
docker-buildx:
	docker buildx build --platform $(PLATFORMS) $(BUILD_ARGS) -f $(DOCKERFILE) \
		-t $(IMAGE_NAME):$(IMAGE_TAG) --load .

.PHONY: compose-up
compose-up:
	docker compose up --build

.PHONY: compose-down
compose-down:
	docker compose down

.PHONY: lint
lint:
	@echo "Add your linter (golangci-lint) here if desired"
