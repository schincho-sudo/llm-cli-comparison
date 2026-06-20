# Python Implementation

OpenAI chatbot CLI built in Python 3. This is the baseline implementation that the TypeScript and Go versions are compared against.

## Setup

Requires Python 3.10 or newer.

```bash
# 1. Create a virtual environment (recommended)
python -m venv venv
source venv/bin/activate          # On Windows: venv\Scripts\activate

# 2. Install dependencies
pip install -r requirements.txt

# 3. Configure your API key
cp .env.example .env
# Open .env in a text editor and paste your real OPENAI_API_KEY
```

Get an OpenAI API key at https://platform.openai.com/api-keys.

## Usage

```bash
# Single-prompt mode
python chat.py "What is the capital of France?"

# Interactive multi-turn mode
python chat.py -i

# Choose a different model
python chat.py -m gpt-4o "explain entropy"

# Disable streaming (wait for full response)
python chat.py --no-stream "hello"

# Use a system prompt
python chat.py -s "You are a pirate" "tell me a story"

# Lower temperature for more deterministic output
python chat.py -t 0.2 "what is 2 + 2?"

# Help
python chat.py --help
```

## Features

- **Streaming responses** by default (chunks appear as they're generated)
- **Multi-turn conversations** in interactive mode (`-i`) with full message history
- **Configurable model, temperature, max tokens, and system prompt**
- **Type hints throughout** using Python's dataclasses
- **Async I/O** using `asyncio` and the OpenAI async client
- **Clean error handling** for auth, rate limits, network errors, and API errors

## Files

- `chat.py` - Main CLI program (about 220 lines)
- `requirements.txt` - Python package dependencies
- `.env.example` - Template for the .env file

## How it works

The program uses the official `openai` SDK in async mode (`AsyncOpenAI`). When you send a prompt, it's added to the message history along with any system prompt, then the request goes out to OpenAI's chat completions endpoint. If streaming is enabled (the default), the response chunks are printed as they arrive using an `async for` loop. Otherwise the full response is fetched and printed at once.

In interactive mode, the same session and message history are kept across turns so the model has context for follow-up questions.

## Notes for the comparison

This implementation will be benchmarked and compared against the TypeScript and Go versions on:

- Lines of code
- Cold start time
- Per-call latency
- How the language handles async iteration over the streaming response
- How errors are caught and reported
- How types are declared (or not)
