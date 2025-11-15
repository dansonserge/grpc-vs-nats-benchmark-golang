# -----------------------------
# Variables
# -----------------------------
DOCKER_COMPOSE=docker-compose
BENCH_CONTAINER=benchmark

# Default values for environment variables
MODE ?= all           # can be grpc, nats, or all
REQUESTS ?= 50000
CONCURRENCY ?= 200
PAYLOAD_SIZE ?= 16
TIMEOUT_MS ?= 500

# -----------------------------
# Phony targets
# -----------------------------
.PHONY: all build up bench down clean report full-bench

# -----------------------------
# Build all Docker images
# -----------------------------
build:
	$(DOCKER_COMPOSE) build

# -----------------------------
# Start NATS and gRPC services
# -----------------------------
up:
	$(DOCKER_COMPOSE) up -d nats grpc-server

# -----------------------------
# Run benchmark
# -----------------------------
bench:
	$(DOCKER_COMPOSE) run --rm \
		-e MODE=$(MODE) \
		-e REQUESTS=$(REQUESTS) \
		-e CONCURRENCY=$(CONCURRENCY) \
		-e PAYLOAD_SIZE=$(PAYLOAD_SIZE) \
		-e TIMEOUT_MS=$(TIMEOUT_MS) \
		$(BENCH_CONTAINER)

# -----------------------------
# Generate report
# -----------------------------
report:
	go run ./cmd/report

# -----------------------------
# Full workflow: build, up, run benchmark, generate report
# -----------------------------
full-bench: build up bench report

# -----------------------------
# Stop all containers
# -----------------------------
down:
	$(DOCKER_COMPOSE) down

# -----------------------------
# Clean everything including images
# -----------------------------
clean: down
	$(DOCKER_COMPOSE) rm -f
	docker system prune -f
