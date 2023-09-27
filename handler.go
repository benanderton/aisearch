package main

import (
	"context"
	"fmt"

	"github.com/nexidian/gocliselect"
	"github.com/rexliu0715/go-pinecone-rest"
	"github.com/sashabaranov/go-openai"
)

type Handler struct {
	ctx            context.Context
	c              Config
	openAIClient   *openai.Client
	pineconeClient pinecone.Client
}

func NewHandler(ctx context.Context, c Config, openAIClient *openai.Client, pineconeClient pinecone.Client) *Handler {
	return &Handler{
		ctx:            ctx,
		c:              c,
		openAIClient:   openAIClient,
		pineconeClient: pineconeClient,
	}
}

func (h *Handler) start() error {
	menu := buildMenu()
	choice := menu.Display()

	var err error
	switch choice {
	case menuItemEmbedData:
		err = h.handleEmbeddingsGeneration()
	case menuItemQueryCatalogue:
		err = h.search()
	default:
		return fmt.Errorf("invalid choice: %s", choice)
	}

	if err != nil {
		return err
	}

	return nil
}

func buildMenu() *gocliselect.Menu {
	menu := gocliselect.NewMenu("Choose an option")
	for _, item := range menuItems {
		menu.AddItem(item.Name, item.Value)
	}

	return menu
}
