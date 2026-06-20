"""
chat.py - OpenAI chatbot CLI in Python

Part of the multi-language LLM wrapper comparison project for ITCS 4102.
Shows how Python handles async APIs, error handling, type hints, and CLI parsing.

Usage:
    python chat.py "What is the capital of France?"        # single prompt
    python chat.py -i                                       # interactive mode
    python chat.py -m gpt-4o "tell me a joke"               # custom model
    python chat.py --no-stream "hello"                      # disable streaming
    python chat.py -s "You are a pirate" "tell a story"     # system prompt
"""

import argparse
import asyncio
import os
import sys
from dataclasses import dataclass, field
from typing import Optional

from dotenv import load_dotenv
from openai import (
    AsyncOpenAI,
    APIError,
    APIConnectionError,
    AuthenticationError,
    OpenAIError,
    RateLimitError,
)


# ---------- Configuration & Session ----------

@dataclass
class ChatConfig:
    """User-facing configuration for a chat session."""
    model: str = "gpt-4o-mini"
    temperature: float = 0.7
    max_tokens: int = 1024
    system_prompt: Optional[str] = None
    stream: bool = True


@dataclass
class ChatSession:
    """Tracks the message history for multi-turn conversations."""
    config: ChatConfig
    messages: list[dict] = field(default_factory=list)

    def __post_init__(self):
        # Seed the conversation with a system prompt if one was given.
        if self.config.system_prompt:
            self.messages.append({
                "role": "system",
                "content": self.config.system_prompt,
            })

    def add_user(self, content: str) -> None:
        self.messages.append({"role": "user", "content": content})

    def add_assistant(self, content: str) -> None:
        self.messages.append({"role": "assistant", "content": content})


# ---------- API interaction ----------

async def send_message(client: AsyncOpenAI, session: ChatSession, prompt: str) -> str:
    """
    Send one user prompt, print the response (streamed or all-at-once),
    and append both turns to the session history.

    Returns the assistant's full response text.
    """
    session.add_user(prompt)

    try:
        if session.config.stream:
            return await _send_streaming(client, session)
        return await _send_blocking(client, session)
    except AuthenticationError:
        _die("Invalid API key. Check OPENAI_API_KEY in your environment.")
    except RateLimitError:
        _die("Rate limit exceeded. Wait a moment and try again.")
    except APIConnectionError as e:
        _die(f"Network error talking to OpenAI: {e}")
    except APIError as e:
        _die(f"OpenAI API error: {e}")
    except OpenAIError as e:
        _die(f"OpenAI client error: {e}")


async def _send_streaming(client: AsyncOpenAI, session: ChatSession) -> str:
    """Stream the response chunk-by-chunk to stdout as it arrives."""
    chunks: list[str] = []

    stream = await client.chat.completions.create(
        model=session.config.model,
        messages=session.messages,
        temperature=session.config.temperature,
        max_tokens=session.config.max_tokens,
        stream=True,
    )

    async for event in stream:
        # Each event has a .choices[0].delta.content which may be None or a string.
        delta = event.choices[0].delta.content
        if delta:
            print(delta, end="", flush=True)
            chunks.append(delta)

    print()  # final newline after streaming finishes
    full = "".join(chunks)
    session.add_assistant(full)
    return full


async def _send_blocking(client: AsyncOpenAI, session: ChatSession) -> str:
    """Wait for the full response, then print it."""
    response = await client.chat.completions.create(
        model=session.config.model,
        messages=session.messages,
        temperature=session.config.temperature,
        max_tokens=session.config.max_tokens,
        stream=False,
    )
    text = response.choices[0].message.content or ""
    print(text)
    session.add_assistant(text)
    return text


# ---------- Interactive REPL ----------

async def interactive_mode(client: AsyncOpenAI, session: ChatSession) -> None:
    """Multi-turn chat loop. Conversation history is kept inside the session."""
    streaming = "on" if session.config.stream else "off"
    print(f"chat ({session.config.model}, streaming={streaming})")
    print("Type 'exit' or press Ctrl+D to quit.\n")

    while True:
        try:
            prompt = input("> ").strip()
        except (EOFError, KeyboardInterrupt):
            print()
            return

        if not prompt:
            continue
        if prompt.lower() in ("exit", "quit"):
            return

        await send_message(client, session, prompt)
        print()  # blank line between turns


# ---------- CLI plumbing ----------

def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        prog="chat",
        description="OpenAI chatbot CLI (Python implementation).",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog=(
            "examples:\n"
            "  chat \"What is the capital of France?\"\n"
            "  chat -i\n"
            "  chat -m gpt-4o -t 0.2 \"explain entropy in one paragraph\"\n"
        ),
    )
    parser.add_argument("prompt", nargs="*",
                        help="Prompt text. Omit for interactive mode.")
    parser.add_argument("-i", "--interactive", action="store_true",
                        help="Force interactive multi-turn mode.")
    parser.add_argument("-m", "--model", default="gpt-4o-mini",
                        help="Model to use (default: gpt-4o-mini).")
    parser.add_argument("-s", "--system", default=None,
                        help="Optional system prompt.")
    parser.add_argument("-t", "--temperature", type=float, default=0.7,
                        help="Sampling temperature 0.0-2.0 (default: 0.7).")
    parser.add_argument("--max-tokens", type=int, default=1024,
                        help="Maximum response tokens (default: 1024).")
    parser.add_argument("--no-stream", action="store_true",
                        help="Disable streaming, wait for full response.")
    return parser.parse_args()


def _die(msg: str) -> None:
    """Print an error to stderr and exit with code 1."""
    print(f"error: {msg}", file=sys.stderr)
    sys.exit(1)


async def amain() -> None:
    args = parse_args()
    load_dotenv()

    if not os.getenv("OPENAI_API_KEY"):
        _die("OPENAI_API_KEY is not set. Add it to a .env file or export it.")

    config = ChatConfig(
        model=args.model,
        temperature=args.temperature,
        max_tokens=args.max_tokens,
        system_prompt=args.system,
        stream=not args.no_stream,
    )

    session = ChatSession(config=config)
    client = AsyncOpenAI()  # picks up OPENAI_API_KEY from env automatically

    try:
        if args.interactive or not args.prompt:
            await interactive_mode(client, session)
        else:
            prompt = " ".join(args.prompt)
            await send_message(client, session, prompt)
    finally:
        await client.close()


def main() -> None:
    try:
        asyncio.run(amain())
    except KeyboardInterrupt:
        print()
        sys.exit(130)


if __name__ == "__main__":
    main()
