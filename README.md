# Multi-Language LLM Wrapper Comparison

Same AI chatbot CLI implemented in three languages: **Python**, **TypeScript**, and **Go**. Built to compare how each language handles async APIs, type systems, error handling, and concurrency.

Final project for ITCS 4102 - Programming Languages (UNCC).

## What it does

Each implementation is a command-line chat tool that talks to OpenAI's chat completions API. Same features in every version:

- Single-prompt mode (`chat "hello"`) and interactive multi-turn mode (`chat -i`)
- Streaming responses by default, can be disabled with `--no-stream`
- Configurable model, temperature, max tokens, and system prompt
- API key loaded from `OPENAI_API_KEY` environment variable or `.env` file
- Clean error handling for auth failures, rate limits, and network errors

## Project structure

```
.
├── python/         Python 3.10+ implementation (asyncio + openai SDK)
├── typescript/     TypeScript implementation (Node 20+ + openai SDK)
├── go/             Go 1.22+ implementation (goroutines + openai-go SDK)
├── benchmarks/     Scripts for comparing the three implementations
└── report/         Final report (comparison + GenAI reflection)
```

Each language folder has its own README with setup and run instructions.

## Quick start

Each implementation needs its own setup. Pick a folder and follow the README inside it.

```bash
# Python version
cd python && pip install -r requirements.txt && python chat.py "hello"

# TypeScript version (coming next)
cd typescript && npm install && npm run start "hello"

# Go version (coming after that)
cd go && go run main.go "hello"
```

You'll need an OpenAI API key. Get one at https://platform.openai.com/api-keys.

## Comparison angle

The same chat CLI in three languages exposes how each one handles common API integration challenges:

| Aspect | Python | TypeScript | Go |
|---|---|---|---|
| Type system | Dynamic + optional hints | Gradual (structural) | Static (nominal) |
| Async model | asyncio event loop | Promises + async/await | Goroutines + channels |
| Error handling | Exceptions | Exceptions + Promise rejection | Explicit error returns |
| Compilation | Interpreted | Transpiled to JS | Compiled to native binary |
| Package manager | pip | npm | go modules |

The final report has the detailed write-up with code snippets, benchmarks, lines-of-code comparison, and reflection on what each language did well or badly.

## Status

- [x] Python implementation
- [ ] TypeScript implementation
- [ ] Go implementation
- [ ] Benchmarks
- [ ] Final report
