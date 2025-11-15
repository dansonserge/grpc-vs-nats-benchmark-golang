# GRPC vs NATS



# Run both benchmarks sequentially
make bench MODE=all

# Run only gRPC benchmark
make bench MODE=grpc

# Run only NATS benchmark
make bench MODE=nats

# Full workflow
make full-bench