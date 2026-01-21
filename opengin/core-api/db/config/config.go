// Copyright 2025 Lanka Data Foundation
// SPDX-License-Identifier: Apache-2.0

package config

type MongoConfig struct {
	URI        string `env:"MONGO_URI"`
	DBName     string `env:"MONGO_DB_NAME"`
	Collection string `env:"MONGO_COLLECTION"`
}

type Neo4jConfig struct {
	URI      string `env:"NEO4J_URI"`
	Username string `env:"NEO4J_USER"`
	Password string `env:"NEO4J_PASSWORD"`
}

type PostgresConfig struct {
	Host     string `env:"POSTGRES_HOST"`
	Port     string `env:"POSTGRES_PORT"`
	User     string `env:"POSTGRES_USER"`
	Password string `env:"POSTGRES_PASSWORD"`
	DBName   string `env:"POSTGRES_DB"`
	SSLMode  string `env:"POSTGRES_SSL_MODE"`
}
