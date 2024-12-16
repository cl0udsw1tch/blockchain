# Blockchain Client

This is a Go-based client project that provides various functionalities for interacting with a networked service, handling transactions, client operations, and managing the wallet's UTXO set.

## Project Structure

This is a high-level overview of the project structure:

```
├── cmd
│   ├── app
│   │   └── main.go          # Main entry point for the application
│   └── cli
│       └── cli.go           # Command-line interface for the app
├── go.mod                    # Go modules file
├── go.sum                    # Go checksum file
├── internal
│   ├── client
│   │   ├── client.go         # Client implementation
│   │   └── keys.go           # Key management related code
│   ├── network
│   │   ├── apiController.go  # Network API controller
│   │   └── network.go        # Network-related utilities
│   ├── t_config
│   │   └── t_config.go       # Configuration handling
│   ├── t_error
│   │   └── error.go          # Error handling utilities
│   ├── t_util
│   │   └── util.go           # Utility functions
│   ├── transaction
│   │   ├── codec.go          # Transaction codec for encoding/decoding
│   │   ├── opstack.go        # Op stack for transaction processing
│   │   ├── script.go         # Script handling for transactions
│   │   └── transaction.go    # Core transaction logic
│   ├── utxoSet
│   │   └── utxoStore.go      # UTXO Store for managing unspent transaction outputs
│   └── wallet
│       ├── wallet.go         # Wallet related functionality
│       └── walletController.go # Wallet controller logic
├── Makefile                  # Build instructions and commands for the project
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

- To run the application, execute:
  ```
  go run cmd/app/main.go
  ```
- To interact with the command-line interface:
  ```
  go run cmd/cli/cli.go [command]
  ```


