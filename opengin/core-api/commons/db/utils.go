// Copyright 2025 Lanka Data Foundation
// SPDX-License-Identifier: Apache-2.0

package dbcommons

import (
	"context"
	"fmt"
	"os"

	"lk/datafoundation/core-api/db/config"
	mongorepository "lk/datafoundation/core-api/db/repository/mongo"
	neo4jrepository "lk/datafoundation/core-api/db/repository/neo4j"
	postgresrepository "lk/datafoundation/core-api/db/repository/postgres"
)

// GetNeo4jConfig creates a Neo4jConfig from environment variables
func GetNeo4jConfig() *config.Neo4jConfig {
	return &config.Neo4jConfig{
		URI:      os.Getenv("NEO4J_URI"),
		Username: os.Getenv("NEO4J_USER"),
		Password: os.Getenv("NEO4J_PASSWORD"),
	}
}

// GetMongoConfig creates a MongoConfig from environment variables
func GetMongoConfig() *config.MongoConfig {
	return &config.MongoConfig{
		URI:        os.Getenv("MONGO_URI"),
		DBName:     os.Getenv("MONGO_DB_NAME"),
		Collection: os.Getenv("MONGO_COLLECTION"),
	}
}

// GetNeo4jRepository retrieves a Neo4j repository
func GetNeo4jRepository(ctx context.Context) (*neo4jrepository.Neo4jRepository, error) {
	cfg := GetNeo4jConfig()
	repo, err := neo4jrepository.NewNeo4jRepository(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("[Commons] failed to create Neo4j repository: %w", err)
	}
	return repo, nil
}

// GetMongoRepository retrieves a Mongo repository
// TODO: Handle errors better
func GetMongoRepository(ctx context.Context) *mongorepository.MongoRepository {
	cfg := GetMongoConfig()
	return mongorepository.NewMongoRepository(ctx, cfg)
}

// GetPostgresConfig creates a PostgresConfig from environment variables
func GetPostgresConfig() postgresrepository.Config {
	return postgresrepository.Config{
		Host:     os.Getenv("POSTGRES_HOST"),
		Port:     os.Getenv("POSTGRES_PORT"),
		User:     os.Getenv("POSTGRES_USER"),
		Password: os.Getenv("POSTGRES_PASSWORD"),
		DBName:   os.Getenv("POSTGRES_DB"),
		SSLMode:  os.Getenv("POSTGRES_SSL_MODE"),
	}
}

// GetPostgresRepository retrieves a Postgres repository
func GetPostgresRepository(ctx context.Context) (*postgresrepository.PostgresRepository, error) {
	cfg := GetPostgresConfig()
	repo, err := postgresrepository.NewPostgresRepository(cfg)
	if err != nil {
		return nil, fmt.Errorf("[Commons] failed to create Postgres repository: %w", err)
	}
	return repo, nil
}
