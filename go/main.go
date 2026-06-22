// main.go - OpenAI chatbot CLI in Go
//
// Part of the multi-language LLM wrapper comparison project for ITCS 4102.
// Shows how Go handles APIs, error returns, structs, and concurrency.
//
// Usage:
//
//	go run main.go "What is the capital of France?"
//	go run main.go -i
//	go run main.go -m gpt-4o "tell me a joke"
//	go run main.go -no-stream "hello"
//	go run main.go -s "You are a pirate" "tell me a story"
package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

// ---------- Configuration & Session ----------

// ChatConfig holds the user-facing settings for a chat session.
type ChatConfig struct {
	Model        string
	Temperature  float64
	MaxTokens    int64
	SystemPrompt string
	Stream       bool
}

// ChatSession tracks the message history for multi-turn conversations.
type ChatSession struct {
	Config   ChatConfig
	Messages []openai.ChatCompletionMessageParamUnion
}

// NewChatSession builds a fresh session and seeds it with a system prompt
// if one was provided.
func NewChatSession(config ChatConfig) *ChatSession {
	s := &ChatSession{Config: config}
	if config.SystemPrompt != "" {
		s.Messages = append(s.Messages, openai.SystemMessage(config.SystemPrompt))
	}
	return s
}

// AddUser appends a user message to the session.
func (s *ChatSession) AddUser(content string) {
	s.Messages = append(s.Messages, openai.UserMessage(content))
}

// AddAssistant appends an assistant message to the session.
func (s *ChatSession) AddAssistant(content string) {
	s.Messages = append(s.Messages, openai.AssistantMessage(content))
}

// ---------- API interaction ----------

// buildParams constructs the request struct for the OpenAI API call.
func (s *ChatSession) buildParams() openai.ChatCompletionNewParams {
	return openai.ChatCompletionNewParams{
		Model:               openai.ChatModel(s.Config.Model),
		Messages:            s.Messages,
		Temperature:         openai.Float(s.Config.Temperature),
		MaxCompletionTokens: openai.Int(s.Config.MaxTokens),
	}
}

// sendMessage sends one user prompt and returns the assistant's full reply.
// In Go, errors are returned as values rather than thrown as exceptions.
func sendMessage(ctx context.Context, client *openai.Client, session *ChatSession, prompt string) (string, error) {
	session.AddUser(prompt)

	if session.Config.Stream {
		return sendStreaming(ctx, client, session)
	}
	return sendBlocking(ctx, client, session)
}

// sendStreaming uses the streaming API and prints chunks as they arrive.
func sendStreaming(ctx context.Context, client *openai.Client, session *ChatSession) (string, error) {
	stream := client.Chat.Completions.NewStreaming(ctx, session.buildParams())

	var chunks []string
	for stream.Next() {
		event := stream.Current()
		if len(event.Choices) > 0 {
			delta := event.Choices[0].Delta.Content
			if delta != "" {
				fmt.Print(delta)
				chunks = append(chunks, delta)
			}
		}
	}

	if err := stream.Err(); err != nil {
		return "", err
	}

	fmt.Println()
	full := strings.Join(chunks, "")
	session.AddAssistant(full)
	return full, nil
}

// sendBlocking waits for the full response, then prints it.
func sendBlocking(ctx context.Context, client *openai.Client, session *ChatSession) (string, error) {
	response, err := client.Chat.Completions.New(ctx, session.buildParams())
	if err != nil {
		return "", err
	}

	if len(response.Choices) == 0 {
		return "", errors.New("no response choices returned")
	}

	text := response.Choices[0].Message.Content
	fmt.Println(text)
	session.AddAssistant(text)
	return text, nil
}

// ---------- Interactive REPL ----------

func interactiveMode(ctx context.Context, client *openai.Client, session *ChatSession) error {
	streaming := "off"
	if session.Config.Stream {
		streaming = "on"
	}
	fmt.Printf("chat (%s, streaming=%s)\n", session.Config.Model, streaming)
	fmt.Println("Type 'exit' or press Ctrl+D to quit.")
	fmt.Println()

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			// Ctrl+D or input closed
			fmt.Println()
			return nil
		}

		prompt := strings.TrimSpace(scanner.Text())
		if prompt == "" {
			continue
		}
		lower := strings.ToLower(prompt)
		if lower == "exit" || lower == "quit" {
			return nil
		}

		if _, err := sendMessage(ctx, client, session, prompt); err != nil {
			return err
		}
		fmt.Println() // blank line between turns
	}
}

// ---------- Error handling ----------

// reportError prints a friendly error message based on the kind of error received.
// Go doesn't have exception classes, so we use errors.As to inspect the error chain.
func reportError(err error) {
	var apiErr *openai.Error
	if errors.As(err, &apiErr) {
		switch apiErr.StatusCode {
		case 401:
			die("Invalid API key. Check OPENAI_API_KEY in your environment.")
		case 429:
			die("Rate limit exceeded. Wait a moment and try again.")
		default:
			msg := apiErr.Message
			if msg == "" {
				msg = err.Error()
			}
			die(fmt.Sprintf("OpenAI API error: %s", msg))
		}
	}
	die(fmt.Sprintf("Error: %s", err.Error()))
}

func die(msg string) {
	fmt.Fprintf(os.Stderr, "error: %s\n", msg)
	os.Exit(1)
}

// ---------- CLI argument parsing ----------

type cliFlags struct {
	interactive bool
	model       string
	system      string
	temperature float64
	maxTokens   int64
	noStream    bool
}

func parseFlags() (cliFlags, []string) {
	var f cliFlags
	flag.BoolVar(&f.interactive, "i", false, "Force interactive multi-turn mode")
	flag.StringVar(&f.model, "m", "gpt-4o-mini", "Model to use")
	flag.StringVar(&f.system, "s", "", "Optional system prompt")
	flag.Float64Var(&f.temperature, "t", 0.7, "Sampling temperature 0.0-2.0")
	flag.Int64Var(&f.maxTokens, "max-tokens", 1024, "Maximum response tokens")
	flag.BoolVar(&f.noStream, "no-stream", false, "Disable streaming, wait for full response")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `usage: chat [options] [prompt...]

OpenAI chatbot CLI (Go implementation).

positional arguments:
  prompt           Prompt text. Omit for interactive mode.

options:
`)
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, `
examples:
  go run main.go "What is the capital of France?"
  go run main.go -i
  go run main.go -m gpt-4o -t 0.2 "explain entropy in one paragraph"
`)
	}

	flag.Parse()
	return f, flag.Args()
}

// ---------- Entry point ----------

func main() {
	// Load .env if present (no error if missing)
	_ = godotenv.Load()

	f, positional := parseFlags()

	if os.Getenv("OPENAI_API_KEY") == "" {
		die("OPENAI_API_KEY is not set. Add it to a .env file or export it.")
	}

	config := ChatConfig{
		Model:        f.model,
		Temperature:  f.temperature,
		MaxTokens:    f.maxTokens,
		SystemPrompt: f.system,
		Stream:       !f.noStream,
	}

	session := NewChatSession(config)

	client := openai.NewClient(
		option.WithAPIKey(os.Getenv("OPENAI_API_KEY")),
	)

	ctx := context.Background()
	prompt := strings.Join(positional, " ")

	var err error
	if f.interactive || prompt == "" {
		err = interactiveMode(ctx, &client, session)
	} else {
		_, err = sendMessage(ctx, &client, session, prompt)
	}

	if err != nil {
		reportError(err)
	}
}
