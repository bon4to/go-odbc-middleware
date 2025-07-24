# Base image with Go
FROM golang:1.23

# Install curl and unzip tools
RUN apt-get update && apt-get install -y \
    curl \
    tar \
    unzip \
    libxml2 \
    && rm -rf /var/lib/apt/lists/*

# Set working directory
WORKDIR /app

# Download and install IBM CLI Driver
RUN curl -LO https://public.dhe.ibm.com/ibmdl/export/pub/software/data/db2/drivers/odbc_cli/linuxx64_odbc_cli.tar.gz && \
    tar -xvzf linuxx64_odbc_cli.tar.gz && \
    rm linuxx64_odbc_cli.tar.gz && \
    mkdir -p /opt/ibm && \
    mv clidriver /opt/ibm/clidriver

# Set environment variables for CGO
ENV CGO_CFLAGS="-I/opt/ibm/clidriver/include"
ENV CGO_LDFLAGS="-L/opt/ibm/clidriver/lib"
ENV LD_LIBRARY_PATH="/opt/ibm/clidriver/lib"

# Copy Go files
COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .

# Build the binary
RUN go build -o go-odbc-middleware ./cmd/server

# Expose port
EXPOSE 40500

# Start the application
CMD ["./go-odbc-middleware"]