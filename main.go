package main

import (
	"context"
	"log"

	"github.com/rexliu0715/go-pinecone-rest"
	"github.com/sashabaranov/go-openai"
)

const (
	menuItemEmbedData      = "menuItemEmbedData"
	menuItemQueryCatalogue = "menuItemQueryCatalogue"
)

// Define struct to hold a menu item
type MenuItem struct {
	Name  string
	Value string
}

// Define a slice of menu items
var menuItems = []MenuItem{
	{
		Name:  "Search catalogue for similar products",
		Value: menuItemQueryCatalogue,
	},
	{
		Name:  "Generate embeddings for data",
		Value: menuItemEmbedData,
	},
}

func main() {
	config, err := parseConfig()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	pineconeClient := pinecone.NewClient(pinecone.Config{
		APIKey:      config.PineconeKey,
		Index:       config.PineconeIndex,
		Environment: config.PineconeEnvironment,
	})

	openAIClient := openai.NewClient(config.OpenAIKey)

	handler := NewHandler(ctx, config, openAIClient, *pineconeClient)
	err = handler.start()

	if err != nil {
		log.Fatal(err)
	}
}
