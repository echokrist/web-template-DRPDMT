# DEVELOPMENT STAGE
FROM rust:alpine AS development

WORKDIR /backend

# Install necessary tools
RUN apk add --no-cache build-base musl-dev

# Install cargo-watch for hot reloading
RUN cargo install cargo-watch

# Copy source code
COPY ./src ./src

# Expose application port for development
EXPOSE 8090

# Command to watch for changes and run the application
CMD ["cargo-watch", "-x", "run"]

########################################################

# BUILDER STAGE (Used for both staging and production)
FROM rust:1.84-alpine AS builder

WORKDIR /backend

# Install required dependencies
RUN apk add --no-cache build-base musl-dev

# Copy Cargo files and fetch dependencies for caching
COPY Cargo.toml Cargo.lock ./
RUN cargo fetch

# Copy source code and build the application
COPY ./src ./src
RUN cargo build --release

########################################################

# PRODUCTION/STAGING STAGE
FROM alpine:3.21.2 AS production

WORKDIR /backend

# Install minimal dependencies
RUN apk add --no-cache libc6-compat

# Copy the built binary from the builder stage
COPY --from=builder /backend/target/release/api /backend/api

# Expose the application port
EXPOSE 8090

# Command to run the application
CMD ["/backend/api"]
