FROM rust:1.92-slim AS builder

WORKDIR /app

# Copy shared proto files
COPY ../shared/proto /app/shared/proto

# Copy matching engine source
COPY . /app/matching-engine

WORKDIR /app/matching-engine

# Install protobuf compiler
RUN apt-get update && apt-get install -y protobuf-compiler && rm -rf /var/lib/apt/lists/*

RUN cargo build --release

FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/matching-engine/target/release/computecoin-matching-engine /usr/local/bin/matching-engine

EXPOSE 50051

CMD ["matching-engine"]
