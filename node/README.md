# Blockchain Node Project

This project is a blockchain node implementation that enables decentralized consensus, transaction processing, block validation, and mining functionalities within the blockchain network.

## Project Structure

This is a high-level overview of the project structure:

```
├── cmd
│   ├── app
│   │   └── main.go            # Main entry point for the application
│   └── cli
│       └── cli.go             # Command-line interface for the app
├── go.mod                      # Go modules file
├── go.sum                      # Go checksum file
├── internal
│   ├── block
│   │   ├── block.go           # Block-related logic
│   │   └── codec.go           # Block codec for encoding/decoding blocks
│   ├── blockchain
│   │   ├── blockchain.go      # Blockchain operations and logic
│   │   ├── constants.go       # Constants related to blockchain rules
│   │   └── proof
│   │       └── proof.go       # Proof logic (e.g., proof of work or stake)
│   ├── blockStore
│   │   ├── blockIO.go         # Block I/O operations
│   │   ├── blockStore.go      # Block store for persisting blocks
│   │   └── indexIO.go         # Index operations for blocks
│   ├── client
│   │   ├── client.go          # Client for interacting with the node
│   │   └── keys.go            # Key management functions for the client
│   ├── mempool
│   │   ├── mempool.go         # Mempool for unconfirmed transactions
│   │   └── mempool.sql        # SQL-based persistence for mempool
│   ├── miner
│   │   └── miner.go           # Mining-related logic and operations
│   ├── network
│   │   └── network.go         # Networking setup for node communication
│   ├── node
│   │   └── node.go            # Node setup and management logic
│   ├── server
│   │   ├── blockStream.go     # Block streaming from peers
│   │   ├── server.go          # Core server operations
│   │   └── txStream.go        # Transaction streaming from peers
│   ├── t_config
│   │   └── config.go          # Configuration settings for the node
│   ├── t_error
│   │   └── error.go           # Error handling utilities
│   ├── t_util
│   │   └── util.go            # Utility functions across the project
│   ├── transaction
│   │   ├── codec.go           # Transaction codec for encoding/decoding transactions
│   │   ├── opstack.go         # Operation stack for transaction processing
│   │   ├── script.go          # Script handling for transactions
│   │   ├── transaction.go     # Core transaction processing logic
│   │   └── txIndexIO.go       # Transaction indexing I/O operations
│   ├── utxoSet
│   │   └── utxoStore.go       # UTXO (Unspent Transaction Outputs) store for wallet
│   ├── validator
│   │   ├── blockValidator.go  # Block validation logic
│   │   └── txValidator.go     # Transaction validation logic
│   └── wallet
│       ├── wallet.go          # Wallet-related logic
│       └── walletController.go # Wallet management and controller logic
├── Makefile                    # Build instructions and commands for the project
```

## Installation

1. Ensure you have Go installed. You can download it from [here](https://golang.org/dl/).
2. Clone the repository:
   ```
   git clone <your-repository-url>
   cd <project-directory>
   ```
3. Run `go mod tidy` to install the necessary dependencies.
4. Build the project by running:
   ```
   make
   ```

## Usage

- To start the blockchain node, execute:
  ```
  go run cmd/app/main.go
  ```
- To interact with the node via the command-line interface (CLI), run:
  ```
  go run cmd/cli/cli.go [command]
  ```

