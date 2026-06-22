/**
 * chat.ts - OpenAI chatbot CLI in TypeScript
 *
 * Part of the multi-language LLM wrapper comparison project for ITCS 4102.
 * Shows how TypeScript handles async APIs, type safety, error handling,
 * and CLI parsing through Node.js.
 *
 * Usage:
 *   npm run start -- "What is the capital of France?"
 *   npm run start -- -i
 *   npm run start -- -m gpt-4o "tell me a joke"
 *   npm run start -- --no-stream "hello"
 *   npm run start -- -s "You are a pirate" "tell me a story"
 */

import 'dotenv/config';
import * as readline from 'node:readline/promises';
import { stdin as input, stdout as output } from 'node:process';

import OpenAI, {
  APIError,
  APIConnectionError,
  AuthenticationError,
  RateLimitError,
} from 'openai';
import type { ChatCompletionMessageParam } from 'openai/resources/chat/completions';

// ---------- Types ----------

interface ChatConfig {
  model: string;
  temperature: number;
  maxTokens: number;
  systemPrompt: string | null;
  stream: boolean;
}

interface ParsedArgs {
  prompt: string;
  interactive: boolean;
  config: ChatConfig;
}

class ChatSession {
  readonly config: ChatConfig;
  readonly messages: ChatCompletionMessageParam[] = [];

  constructor(config: ChatConfig) {
    this.config = config;
    if (config.systemPrompt) {
      this.messages.push({ role: 'system', content: config.systemPrompt });
    }
  }

  addUser(content: string): void {
    this.messages.push({ role: 'user', content });
  }

  addAssistant(content: string): void {
    this.messages.push({ role: 'assistant', content });
  }
}

// ---------- API interaction ----------

async function sendMessage(
  client: OpenAI,
  session: ChatSession,
  prompt: string,
): Promise<string> {
  session.addUser(prompt);

  try {
    if (session.config.stream) {
      return await sendStreaming(client, session);
    }
    return await sendBlocking(client, session);
  } catch (err) {
    handleError(err);
  }
}

async function sendStreaming(client: OpenAI, session: ChatSession): Promise<string> {
  const chunks: string[] = [];

  const stream = await client.chat.completions.create({
    model: session.config.model,
    messages: session.messages,
    temperature: session.config.temperature,
    max_tokens: session.config.maxTokens,
    stream: true,
  });

  for await (const event of stream) {
    const delta = event.choices[0]?.delta?.content;
    if (delta) {
      process.stdout.write(delta);
      chunks.push(delta);
    }
  }

  process.stdout.write('\n');
  const full = chunks.join('');
  session.addAssistant(full);
  return full;
}

async function sendBlocking(client: OpenAI, session: ChatSession): Promise<string> {
  const response = await client.chat.completions.create({
    model: session.config.model,
    messages: session.messages,
    temperature: session.config.temperature,
    max_tokens: session.config.maxTokens,
    stream: false,
  });

  const text = response.choices[0]?.message.content ?? '';
  console.log(text);
  session.addAssistant(text);
  return text;
}

// ---------- Interactive REPL ----------

async function interactiveMode(client: OpenAI, session: ChatSession): Promise<void> {
  const streaming = session.config.stream ? 'on' : 'off';
  console.log(`chat (${session.config.model}, streaming=${streaming})`);
  console.log("Type 'exit' or press Ctrl+D to quit.\n");

  const rl = readline.createInterface({ input, output });

  try {
    while (true) {
      let prompt: string;
      try {
        prompt = (await rl.question('> ')).trim();
      } catch {
        // Ctrl+D or other input closure
        console.log();
        return;
      }

      if (!prompt) continue;
      if (prompt.toLowerCase() === 'exit' || prompt.toLowerCase() === 'quit') {
        return;
      }

      await sendMessage(client, session, prompt);
      console.log(); // blank line between turns
    }
  } finally {
    rl.close();
  }
}

// ---------- Error handling ----------

function handleError(err: unknown): never {
  if (err instanceof AuthenticationError) {
    die('Invalid API key. Check OPENAI_API_KEY in your environment.');
  }
  if (err instanceof RateLimitError) {
    die('Rate limit exceeded. Wait a moment and try again.');
  }
  if (err instanceof APIConnectionError) {
    die(`Network error talking to OpenAI: ${err.message}`);
  }
  if (err instanceof APIError) {
    die(`OpenAI API error: ${err.message}`);
  }
  if (err instanceof Error) {
    die(`Unexpected error: ${err.message}`);
  }
  die(`Unexpected error: ${String(err)}`);
}

function die(msg: string): never {
  console.error(`error: ${msg}`);
  process.exit(1);
}

// ---------- CLI argument parsing ----------

function parseArgs(argv: string[]): ParsedArgs {
  // Skip 'node' and the script path
  const args = argv.slice(2);

  const config: ChatConfig = {
    model: 'gpt-4o-mini',
    temperature: 0.7,
    maxTokens: 1024,
    systemPrompt: null,
    stream: true,
  };
  let interactive = false;
  const positional: string[] = [];

  for (let i = 0; i < args.length; i++) {
    const arg = args[i];
    switch (arg) {
      case '-h':
      case '--help':
        printHelp();
        process.exit(0);
      case '-i':
      case '--interactive':
        interactive = true;
        break;
      case '-m':
      case '--model':
        config.model = requireValue(args, ++i, arg);
        break;
      case '-s':
      case '--system':
        config.systemPrompt = requireValue(args, ++i, arg);
        break;
      case '-t':
      case '--temperature':
        config.temperature = parseFloat(requireValue(args, ++i, arg));
        break;
      case '--max-tokens':
        config.maxTokens = parseInt(requireValue(args, ++i, arg), 10);
        break;
      case '--no-stream':
        config.stream = false;
        break;
      default:
        positional.push(arg);
    }
  }

  return {
    prompt: positional.join(' '),
    interactive,
    config,
  };
}

function requireValue(args: string[], idx: number, flag: string): string {
  const v = args[idx];
  if (v === undefined) die(`${flag} requires a value`);
  return v;
}

function printHelp(): void {
  console.log(`usage: chat [options] [prompt...]

OpenAI chatbot CLI (TypeScript implementation).

positional arguments:
  prompt                Prompt text. Omit for interactive mode.

options:
  -h, --help            Show this help message and exit.
  -i, --interactive     Force interactive multi-turn mode.
  -m, --model MODEL     Model to use (default: gpt-4o-mini).
  -s, --system PROMPT   Optional system prompt.
  -t, --temperature N   Sampling temperature 0.0-2.0 (default: 0.7).
  --max-tokens N        Maximum response tokens (default: 1024).
  --no-stream           Disable streaming, wait for full response.

examples:
  npm run start -- "What is the capital of France?"
  npm run start -- -i
  npm run start -- -m gpt-4o -t 0.2 "explain entropy in one paragraph"
`);
}

// ---------- Entry point ----------

async function main(): Promise<void> {
  const { prompt, interactive, config } = parseArgs(process.argv);

  if (!process.env.OPENAI_API_KEY) {
    die('OPENAI_API_KEY is not set. Add it to a .env file or export it.');
  }

  const session = new ChatSession(config);
  const client = new OpenAI(); // picks up OPENAI_API_KEY from env automatically

  if (interactive || !prompt) {
    await interactiveMode(client, session);
  } else {
    await sendMessage(client, session, prompt);
  }
}

// Run, handling Ctrl+C gracefully
main().catch((err) => {
  if (err instanceof Error && err.message.includes('SIGINT')) {
    process.exit(130);
  }
  console.error('fatal:', err);
  process.exit(1);
});
