package llm

import (
	"context"
	"errors"
	"fmt"
	"io"

	openai "github.com/sashabaranov/go-openai"
)

type OpenAIClient struct {
	client         *openai.Client
	chatModel      string
	embeddingModel openai.EmbeddingModel
}

func NewOpenAIClient(apiKey, chatModel, embeddingModel string) *OpenAIClient {
	return &OpenAIClient{
		client:         openai.NewClient(apiKey),
		chatModel:      chatModel,
		embeddingModel: openai.EmbeddingModel(embeddingModel),
	}
}

func (c *OpenAIClient) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, nil
	}

	const batchSize = 100
	var allEmbeddings [][]float32

	for i := 0; i < len(texts); i += batchSize {
		end := i + batchSize
		if end > len(texts) {
			end = len(texts)
		}
		batch := texts[i:end]

		resp, err := c.client.CreateEmbeddings(ctx, openai.EmbeddingRequest{
			Input: batch,
			Model: c.embeddingModel,
		})
		if err != nil {
			return nil, fmt.Errorf("create embeddings (batch %d-%d): %w", i, end, err)
		}

		for _, emb := range resp.Data {
			allEmbeddings = append(allEmbeddings, emb.Embedding)
		}
	}

	return allEmbeddings, nil
}

func (c *OpenAIClient) ChatStream(ctx context.Context, messages []ChatMessage) (<-chan string, <-chan error) {
	contentCh := make(chan string, 64)
	errCh := make(chan error, 1)

	var openaiMessages []openai.ChatCompletionMessage
	for _, msg := range messages {
		openaiMessages = append(openaiMessages, openai.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	go func() {
		defer close(contentCh)
		defer close(errCh)

		stream, err := c.client.CreateChatCompletionStream(ctx, openai.ChatCompletionRequest{
			Model:    c.chatModel,
			Messages: openaiMessages,
			Stream:   true,
		})
		if err != nil {
			errCh <- fmt.Errorf("create chat stream: %w", err)
			return
		}
		defer stream.Close()

		for {
			resp, err := stream.Recv()
			if errors.Is(err, io.EOF) {
				return
			}
			if err != nil {
				errCh <- fmt.Errorf("receive stream chunk: %w", err)
				return
			}

			if len(resp.Choices) > 0 {
				delta := resp.Choices[0].Delta.Content
				if delta != "" {
					select {
					case contentCh <- delta:
					case <-ctx.Done():
						errCh <- ctx.Err()
						return
					}
				}
			}
		}
	}()

	return contentCh, errCh
}
