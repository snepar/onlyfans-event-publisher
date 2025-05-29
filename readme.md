# OnlyFans Event Publisher

A Go application that simulates and publishes Only Fans Events to Redpanda.

## Features

- Configurable simulation parameters (number of devices, interval, abnormal temperature probability)
- Publishes readings to Redpanda using the franz-go client
- Docker and docker-compose support for easy deployment
- VS Code integration

## Prerequisites

- Go 1.19 or later
- Redpanda (or Kafka) broker
- Docker and Docker Compose (optional)
- VS Code (optional)

## Project Structure

```
gpu-temp-publisher/
├── .vscode/               # VS Code configuration
├── cmd/                   # Application entry points
│   └── publisher/         # Main publisher application
├── internal/              # Internal packages
│   ├── config/            # Configuration handling
│   ├── model/             # Data models
│   ├── publisher/         # Redpanda publishing logic
│   └── simulator/         # simulation logic
├── Dockerfile             # Docker build configuration
├── docker-compose.yml     # Docker Compose configuration
└── go.mod                 # Go module definition
```

## Getting Started

### Local Development

1. Clone the repository:
   ```bash
   git clone https://github.com/snepar/onlyfans-event-publisher.git
   cd  onlyfans-event-publisher
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Update the module name in go.mod to match your GitHub username or organization.

4. Run the application:
   ```bash
   go run cmd/publisher/main.go
   ```

### Environment Variables

The application can be configured using the following environment variables:

- `REDPANDA_BROKERS`: Comma-separated list of Redpanda brokers (default: `localhost:9092`)
- `REDPANDA_TOPIC`: Topic to publish temperature readings to (default: `gpu-temperature`)
- `NUM_DEVICES`: Number of GPU devices to simulate (default: `5`)
- `INTERVAL_MS`: Interval between readings in milliseconds (default: `1000`)
- `ABNORMAL_PROBABILITY`: Probability of generating abnormal temperature readings (default: `0.05`)

### Using VS Code

The project includes VS Code configurations:

1. Open the project in VS Code:
   ```bash
   code .
   ```

2. Install the Go extension if not already installed.

3. Use the "Launch Publisher" debug configuration to run the application with debugging.

### Using Docker

Build and run using Docker:

```bash
docker build -t gpu-temp-publisher .
docker run -e REDPANDA_BROKERS=host.docker.internal:9092 gpu-temp-publisher
```

### Using Docker Compose

The included docker-compose.yml sets up:
- A Redpanda broker
- Redpanda Console (Web UI)
- The GPU temperature publisher

Start the entire stack:

```bash
docker-compose up -d
```

Access Redpanda Console at http://localhost:8080 to view the published messages.

## Customizing the Simulation

You can customize the temperature simulation by modifying the parameters in the `internal/simulator/temp_simulator.go` file