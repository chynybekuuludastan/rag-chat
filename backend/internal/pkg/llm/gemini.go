package llm

import (
	"context"
	"fmt"

	"google.golang.org/genai"
)

type GeminiClient struct {
	client         *genai.Client
	chatModel      string
	embeddingModel string
}

func NewGeminiClient(ctx context.Context, apiKey, chatModel, embeddingModel string) (*GeminiClient, error) {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("create gemini client: %w", err)
	}

	return &GeminiClient{
		client:         client,
		chatModel:      chatModel,
		embeddingModel: embeddingModel,
	}, nil
}

func (c *GeminiClient) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, nil
	}

	allEmbeddings := make([][]float32, 0, len(texts))

	// Gemini EmbedContent takes a single content input, so we loop over texts.
	for i, text := range texts {
		dim := int32(768)
		result, err := c.client.Models.EmbedContent(
			ctx,
			c.embeddingModel,
			genai.Text(text),
			&genai.EmbedContentConfig{
				TaskType:             "RETRIEVAL_DOCUMENT",
				OutputDimensionality: &dim,
			},
		)
		if err != nil {
			return nil, fmt.Errorf("embed content (index %d): %w", i, err)
		}

		if len(result.Embeddings) == 0 {
			return nil, fmt.Errorf("embed content (index %d): no embeddings returned", i)
		}

		allEmbeddings = append(allEmbeddings, result.Embeddings[0].Values)
	}

	return allEmbeddings, nil
}

func (c *GeminiClient) ChatStream(ctx context.Context, messages []ChatMessage) (<-chan string, <-chan error) {
	contentCh := make(chan string, 64)
	errCh := make(chan error, 1)

	go func() {
		defer close(contentCh)
		defer close(errCh)

		// Build contents from messages. The last message is the user prompt;
		// everything before it is history.
		var systemParts []*genai.Part
		var contents []*genai.Content

		for _, msg := range messages {
			role := msg.Role
			if role == "system" {
				systemParts = append(systemParts, &genai.Part{Text: msg.Content})
				continue
			}
			// Gemini uses "model" instead of "assistant"
			if role == "assistant" {
				role = "model"
			}
			contents = append(contents, &genai.Content{
				Role:  role,
				Parts: []*genai.Part{{Text: msg.Content}},
			})
		}

		config := &genai.GenerateContentConfig{}
		if len(systemParts) > 0 {
			config.SystemInstruction = &genai.Content{
				Parts: systemParts,
			}
		}

		if len(contents) == 0 {
			errCh <- fmt.Errorf("no messages to send")
			return
		}

		// Last message is the user prompt; prior messages are history.
		prompt := *contents[len(contents)-1].Parts[0]
		var history []*genai.Content
		if len(contents) > 1 {
			history = contents[:len(contents)-1]
		}

		chat, err := c.client.Chats.Create(ctx, c.chatModel, config, history)
		if err != nil {
			errCh <- fmt.Errorf("create chat: %w", err)
			return
		}

		for result, err := range chat.SendMessageStream(ctx, prompt) {
			if err != nil {
				errCh <- fmt.Errorf("receive stream chunk: %w", err)
				return
			}

			text := result.Text()
			if text != "" {
				select {
				case contentCh <- text:
				case <-ctx.Done():
					errCh <- ctx.Err()
					return
				}
			}
		}
	}()

	return contentCh, errCh
}
