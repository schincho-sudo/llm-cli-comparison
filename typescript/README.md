# TypeScript Implementation

OpenAI chatbot CLI built in TypeScript. Runs on Node.js. Same features as the Python implementation, but with TypeScript's static type checking and Node's async model.

## Setup

Requires **Node.js 20 or newer**. Check with:

```bash
node --version
```

If you don't have Node, install it from https://nodejs.org/ (pick the LTS version).

```bash
# 1. Install dependencies
npm install

# 2. Configure your API key
cp .env.example .env
# Open .env in a text editor and paste your real OPENAI_API_KEY
```

You can reuse the same OpenAI API key from the Python implementation.

## Usage

The CLI takes the same flags as the Python version. The `--` after `npm run start` is required so npm passes the arguments through to the script:

```bash
# Single-prompt mode
npm run start -- "What is the capital of France?"

# Interactive multi-turn mode
npm run start -- -i

# Choose a different model
npm run start -- -m gpt-4o "explain entropy"

# Disable streaming
npm run start -- --no-stream "hello"

# Use a system prompt
npm run start -- -s "You are a pirate" "tell me a story"

# Help
npm run start -- --help
```

## How it runs

The `npm run start` script uses **tsx** to execute the TypeScript source file directly without a separate compile step. This is the fastest way to develop, since changes show up immediately.

For a production build, you can compile to JavaScript:

```bash
npm run build       # compiles src/chat.ts -> dist/chat.js
npm run run-built   # runs the compiled JS with plain node
```

## Features

- **Strict TypeScript mode** with full type annotations
- **Streaming responses** by default (using `for await` over the async iterable returned by the OpenAI SDK)
- **Multi-turn conversations** in interactive mode
- **Configurable model, temperature, max tokens, and system prompt**
- **Manual CLI argument parsing** (no argparse equivalent in Node's standard library)
- **Clean error handling** using TypeScript's `instanceof` narrowing on the OpenAI SDK error classes

## Files

- `src/chat.ts` - Main CLI program (~270 lines)
- `package.json` - npm dependencies and scripts
- `tsconfig.json` - TypeScript compiler config (strict mode on)
- `.env.example` - Template for the .env file

## What's different from the Python version

The behavior is identical. The interesting differences are how each language gets there:

- **Types:** Python uses runtime type hints via dataclasses, TypeScript checks types at compile time and erases them at runtime.
- **Async iteration:** Both use `async for` (Python) / `for await` (TS) over the streaming response. Almost identical syntax.
- **Error handling:** Python catches specific exception types in `except` clauses. TypeScript uses `instanceof` narrowing inside a single `catch` block since `catch` only gives you one binding.
- **CLI parsing:** Python has the excellent `argparse` in the standard library. Node doesn't have an equivalent, so the parsing is hand-rolled (the alternative is adding a library like `commander` or `yargs`).
- **Imports:** Python imports look like `from openai import AsyncOpenAI`. TypeScript uses `import OpenAI, { AuthenticationError } from 'openai';`.
- **Compilation:** Python runs the source directly. TypeScript here uses `tsx` to skip the build step during development, but in production it would compile to JavaScript first.
