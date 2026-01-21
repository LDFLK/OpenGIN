// Copyright 2025 Lanka Data Foundation
// SPDX-License-Identifier: Apache-2.0

package mongorepository

import (
	"context"
	"lk/datafoundation/core-api/db/config"
	"log"

	pb "lk/datafoundation/core-api/lk/datafoundation/core-api"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/protobuf/types/known/anypb"
)

type MongoRepository struct {
	client *mongo.Client
	config *config.MongoConfig
}

// A custom wrapper struct for Entity to use MongoDB's _id field
type entityDocument struct {
	ID            string                            `bson:"_id"`
	Metadata      map[string]*anypb.Any             `bson:"metadata,omitempty"`
	Kind          *pb.Kind                          `bson:"kind,omitempty"`
	Created       string                            `bson:"created,omitempty"`
	Terminated    string                            `bson:"terminated,omitempty"`
	Name          *pb.TimeBasedValue                `bson:"name,omitempty"`
	Attributes    map[string]*pb.TimeBasedValueList `bson:"attributes,omitempty"`
	Relationships map[string]*pb.Relationship       `bson:"relationships,omitempty"`
}

// Convert protobuf Entity to MongoDB document
func toDocument(entity *pb.Entity) interface{} {
	return bson.M{
		"_id":      entity.Id,
		"metadata": entity.Metadata,
		// Map other entity fields as needed
	}
}

// Convert MongoDB document to protobuf Entity
func fromDocument(data *entityDocument) *pb.Entity {
	return &pb.Entity{
		Id:            data.ID,
		Metadata:      data.Metadata,
		Kind:          data.Kind,
		Created:       data.Created,
		Terminated:    data.Terminated,
		Name:          data.Name,
		Attributes:    data.Attributes,
		Relationships: data.Relationships,
	}
}

// NewMongoRepository initializes a MongoDB client
// TODO: Handle errors better
func NewMongoRepository(ctx context.Context, config *config.MongoConfig) *MongoRepository {
	clientOptions := options.Client().ApplyURI(config.URI)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	return &MongoRepository{
		client: client,
		config: config,
	}
}

func (repo *MongoRepository) collection() *mongo.Collection {
	return repo.client.Database(repo.config.DBName).Collection(repo.config.Collection)
}

// CreateEntity inserts a new entity in MongoDB
// FIXME: https://github.com/LDFLK/nexoan/issues/118
func (repo *MongoRepository) CreateEntity(ctx context.Context, entity *pb.Entity) (*mongo.InsertOneResult, error) {
	// Use the entity.Id as MongoDB's _id field
	doc := toDocument(entity)
	result, err := repo.collection().InsertOne(ctx, doc)
	return result, err
}

// ReadEntity fetches an entity by ID from MongoDB
func (repo *MongoRepository) ReadEntity(ctx context.Context, id string) (*pb.Entity, error) {
	var doc entityDocument
	err := repo.collection().FindOne(ctx, bson.M{"_id": id}).Decode(&doc)
	if err != nil {
		return nil, err
	}
	return fromDocument(&doc), nil
}

// UpdateEntity updates an entity's attributes in MongoDB
func (repo *MongoRepository) UpdateEntity(ctx context.Context, id string, updates bson.M) (*mongo.UpdateResult, error) {
	update := bson.M{"$set": updates}
	result, err := repo.collection().UpdateOne(ctx, bson.M{"_id": id}, update)
	return result, err
}

// DeleteEntity removes an entity from MongoDB
func (repo *MongoRepository) DeleteEntity(ctx context.Context, id string) (*mongo.DeleteResult, error) {
	result, err := repo.collection().DeleteOne(ctx, bson.M{"_id": id})
	return result, err
}
