# Blockchain

This project consists of two major components:

1. **Client**: A Go-based client that interacts with the blockchain network, manages transactions, and provides wallet functionality.
2. **Node**: A blockchain node that supports decentralized consensus, block validation, mining, and transaction processing.

## Overview

The blockchain node and client are designed to work together in a decentralized blockchain network. The node is responsible for maintaining the blockchain ledger, validating transactions, and supporting mining operations, while the client provides the interface for interacting with the blockchain.

## Project Structure

```
├── client                # Client application for interacting with the blockchain
│   └── README.md         # Client README with details on building and running the client
├── node                  # Blockchain node implementation
│   └── README.md         # Node README with details on building and running the node
├── go.mod                # Go modules file for dependency management
├── go.sum                # Go checksum file
├── Makefile              # Build instructions for the entire project
```

## Building the Project

### 1. Building the Client

To build the client application:

1. Navigate to the `client/` directory:
   ```
   cd client
   ```

2. Run `go mod tidy` to install the necessary dependencies:
   ```
   go mod tidy
   ```

3. Build the client:
   ```
   make
   ```

4. Follow the [client README](client/README.md) for more detailed instructions on running and interacting with the client.

### 2. Building the Node

To build the blockchain node:

1. Navigate to the `node/` directory:
   ```
   cd node
   ```

2. Run `go mod tidy` to install the necessary dependencies:
   ```
   go mod tidy
   ```

3. Build the node:
   ```
   make
   ```

4. Follow the [node README](node/README.md) for more detailed instructions on running and interacting with the node.

## Running the Project

### Running the Client

To run the client application:

1. Navigate to the `client/` directory:
   ```
   cd client
   ```

2. Start the client:
   ```
   go run cmd/app/main.go
   ```

### Running the Node

To run the blockchain node:

1. Navigate to the `node/` directory:
   ```
   cd node
   ```

2. Start the blockchain node:
   ```
   go run cmd/app/main.go
   ```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
