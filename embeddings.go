package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/rexliu0715/go-pinecone-rest"
	"github.com/sashabaranov/go-openai"
)

type Product struct {
	OptionId        int    `json:"Option Id"`
	OptionDesc      string `json:"Option Desc"`
	DivisionDesc    string `json:"Division Desc"`
	LWOrigRetailGBP int    `json:"LW Orig Retail GBP"`
	ColourDescr     string `json:"Colour Descr"`
	GrossProfit     string `json:"Gross Profit"`
	Vector          []float32
}

func (h *Handler) handleEmbeddingsGeneration() error {
	products, err := h.getProductsFromFile()
	if err != nil {
		fmt.Println(err)
		return err
	}

	// DEBUG: Keep only the first 10 products
	// products = products[:3]

	fmt.Printf("Generating Embeddings for %d products\n", len(products))

	productsWithEmbeddings, err := h.generateEmbeddingsForProducts(products)
	if err != nil {
		return err
	}
	fmt.Printf("Successfully generated embeddings for %d products\n", len(productsWithEmbeddings))

	err = h.upsertEmbeddingsToPinecone(productsWithEmbeddings)
	if err != nil {
		return err
	}

	return nil
}

func (h *Handler) getProductsFromFile() ([]Product, error) {
	data, err := os.ReadFile("input/top_100_yd.json")
	if err != nil {
		return nil, err
	}

	var products []Product
	err = json.Unmarshal(data, &products)
	if err != nil {
		return nil, err
	}

	return products, nil
}

func (h *Handler) generateEmbeddingsForProducts(products []Product) ([]Product, error) {
	// Combine fields
	for i, product := range products {
		fmt.Println("Generating embeddings for product: ", product.OptionId)
		combined := fmt.Sprintf("%s %s %s", product.OptionDesc, product.DivisionDesc, product.ColourDescr)
		resp, err := h.openAIClient.CreateEmbeddings(
			h.ctx,
			openai.EmbeddingRequest{
				Input: combined,
				Model: openai.AdaEmbeddingV2,
			},
		)

		if err != nil {
			return nil, err
		}

		products[i].Vector = resp.Data[0].Embedding
	}
	return products, nil
}

func (h *Handler) generateEmbeddingForString(searchTerm string) (openai.EmbeddingResponse, error) {
	embedding, err := h.openAIClient.CreateEmbeddings(
		h.ctx,
		openai.EmbeddingRequest{
			Input: searchTerm,
			Model: openai.AdaEmbeddingV2,
		},
	)
	if err != nil {
		return openai.EmbeddingResponse{}, err
	}

	return embedding, nil
}

func (h *Handler) upsertEmbeddingsToPinecone(products []Product) error {
	namespace := h.c.PineconeNamespace

	var pineconeVectors []pinecone.Vector
	for i, product := range products {
		metadata := make(map[string]interface{})
		metadata["OptionDesc"] = product.OptionDesc
		metadata["IsActual"] = true

		pineconeVectors = append(pineconeVectors, pinecone.Vector{
			ID:       fmt.Sprintf("%d", product.OptionId),
			Values:   &products[i].Vector,
			Metadata: &metadata,
		})
	}

	res, err := h.pineconeClient.Upsert(&pinecone.UpsertRequest{
		Namespace: &namespace,
		Vectors:   pineconeVectors,
	})

	if err != nil {
		return err
	}

	fmt.Printf("Upserted %d vectors into Pinecone\n", res.UpsertedCount)
	return nil
}
