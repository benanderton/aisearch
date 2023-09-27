package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type EmbeddingResponse struct {
	Matches   []Match `json:"matches"`
	Namespace string  `json:"namespace"`
}

type Match struct {
	ID       string                 `json:"id"`
	Score    float32                `json:"score"`
	Metadata map[string]interface{} `json:"metadata"`
}

// For some reason the Pinecone package I'm using gets the score wrong so I'm fetching directly from the API
func (h *Handler) queryPineconeViaHTTP(embedding []float32) (EmbeddingResponse, error) {
	var response EmbeddingResponse

	// Turn embedding into a comma seperated string
	var sb strings.Builder
	for i, v := range embedding {
		sb.WriteString(fmt.Sprintf("%f", v))
		if i != len(embedding)-1 {
			sb.WriteString(",")
		}
	}
	url := fmt.Sprintf("https://%s.svc.%s.pinecone.io/query", h.c.PineconeIndex, h.c.PineconeEnvironment)
	payload := strings.NewReader(fmt.Sprintf("{\"includeMetadata\":true,\"vector\":[%s],\"namespace\":\"%s\",\"topK\":5}", sb.String(), h.c.PineconeNamespace))
	req, _ := http.NewRequest("POST", url, payload)
	req.Header.Add("accept", "application/json")
	req.Header.Add("content-type", "application/json")
	req.Header.Add("Api-Key", h.c.PineconeKey)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return response, err
	}

	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	// marshall the response
	err = json.Unmarshal(body, &response)
	if err != nil {
		return response, err
	}

	return response, nil
}

func (h *Handler) search() error {
	// Get input from CLI:
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter a product description: ")
	productDescription, err := reader.ReadString('\n')
	if err != nil {
		return err
	}

	// Generate embeddings for the input
	embedding, err := h.generateEmbeddingForString(productDescription)
	if err != nil {
		return err
	}

	// Find items that match the embedding
	res, err := h.queryPineconeViaHTTP(embedding.Data[0].Embedding)
	if err != nil {
		return err
	}

	// Keep matches that were close
	closeMatches := make([]Match, 0)
	for _, match := range res.Matches {
		if match.Score <= 0.28 { // Only show close matches
			closeMatches = append(closeMatches, match)
		}
	}

	// If no close matches found, return, but push this new item to the catalogue as a non-actual item
	if len(closeMatches) == 0 {
		fmt.Println("No close matches found - we should log this as a new item in the catalogue")
		return nil
	}

	for _, result := range closeMatches {
		fmt.Println("----------------------------------------------")
		fmt.Printf("Product ID: %s\n", result.ID)
		fmt.Printf("Score: %f (low is better)\n", result.Score)
		fmt.Printf("OptionDesc: %s\n", result.Metadata["OptionDesc"])
	}

	// @TODO: Add any misses to a catalogue with an "isActual" flag set to false. Increment the counter in the metadata each time this is done

	return nil
}
