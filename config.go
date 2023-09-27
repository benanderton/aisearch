package main

import "github.com/caarlos0/env"

type Config struct {
	OpenAIKey           string `env:"OPENAI_KEY,required"`
	PineconeKey         string `env:"PINECONE_KEY,required"`
	PineconeIndex       string `env:"PINECONE_INDEX" envDefault:"products-a893d2b"`
	PineconeEnvironment string `env:"PINECONE_ENVIRONMENT" envDefault:"asia-southeast1-gcp-free"`
	PineconeNamespace   string `env:"PINECONE_NAMESPACE" envDefault:"products"`
}

func parseConfig() (Config, error) {
	var c Config
	return c, env.Parse(&c)
}
