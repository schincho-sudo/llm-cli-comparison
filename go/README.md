# Go Implementation

OpenAI chatbot CLI built in Go. Compiled to a native binary, no runtime needed once built. Same features as the Python and TypeScript implementations.

## Setup

Requires **Go 1.22 or newer**. Check with:

```bash
go version
```

If you don't have Go installed:
- **Windows:** Download from https://go.dev/dl/ (pick the `.msi` installer)
- **macOS/Linux:** Same site, or use a package manager (`brew install go`, `apt install golang-go`)

Then:

```bash
# 1. Download dependencies
go mod tidy

# 2. Configure your API key
cp .env.example .env
# Open .env in a text editor and paste your real OPENAI_API_KEY
```

You can reuse the same OpenAI API key from the Python and TypeScript implementations.

## Usage

Go's flag library uses **single-dash** flags by default (different from the Python and TypeScript implementations which use both `-i` and `--interactive`). The functionality is identical.

```bash
# Single-prompt mode
go run main.go "What is the capital of France?"

# Interactive multi-turn mode
go run main.go -i

# Choose a different model
go run main.go -m gpt-4o "explain entropy"

# Disable streaming
go run main.go -no-stream "hello"

# Use a system prompt
go run main.go -s "You are a pirate" "tell me a story"

# Help
go run main.go -h
```

## Build a standalone binary

Unlike Python and TypeScript which need their runtimes installed, a compiled Go binary runs on its own:

```bash
go build -o chat main.go
./chat "what is the capital of france?"     # macOS/Linux
chat.exe "what is the capital of france?"   # Windows
```

The binary is around 11 MB and includes everything it needs (no Go installation required to run it on another machine of the same OS/arch).

## Features

- **Native compilation** to a single static binary
- **Streaming responses** by default using Go's iterator-style `for stream.Next()` pattern
- **Multi-turn conversations** in interactive mode
- **Configurable model, temperature, max tokens, and system prompt**
- **Explicit error handling** using Go's `(value, error)` return pattern instead of exceptions
- **Standard library `flag` package** for CLI parsing (no external library needed)

## Files

- `main.go` - Main CLI program (~240 lines)
- `go.mod` and `go.sum` - Go module manifest and dependency hashes
- `.env.example` - Template for the .env file

## What's different from Python and TypeScript

The behavior is the same. What's interesting are the language-level differences:

- **No exceptions.** Every function that can fail returns `(value, error)` and the caller must check for an error. The `reportError` function uses `errors.As` to inspect what kind of error came back, similar in spirit to `except SpecificError` in Python or `instanceof` in TypeScript, but more verbose.
- **Static, nominal typing with structs.** `ChatConfig` and `ChatSession` are plain structs with method receivers. No classes, no inheritance.
- **`context.Context` everywhere.** Idiomatic Go passes a `Context` through all API calls so they can be cancelled. The other implementations don't have a direct equivalent.
- **`for stream.Next()`** instead of `async for` (Python) or `for await` (TypeScript). The streaming iterator is explicit rather than async/await syntax sugar.
- **Compiled, not interpreted.** One `go build` produces a binary that starts in milliseconds, no Python interpreter or Node runtime needed.
- **Standard library is enough.** `flag`, `bufio`, `context`, `os`, `strings` come with Go. The only external dependencies are `openai-go` (the SDK) and `godotenv` (.env loading).
