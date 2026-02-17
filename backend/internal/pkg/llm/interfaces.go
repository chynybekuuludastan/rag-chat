package llm

import "context"

type ChatMessage struct {
	Role    string
	Content string
}

type LLMClient interface {
	Embed(ctx context.Context, texts []string) ([][]float32, error)
	ChatStream(ctx context.Context, messages []ChatMessage) (<-chan string, <-chan error)
}
