package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ConnectMongo connects to MongoDB and returns the client and a database handle.
func ConnectMongo(ctx context.Context, uri string, dbName string) (*mongo.Client, *mongo.Database, error) {
    clientOpts := options.Client().ApplyURI(uri)
    client, err := mongo.Connect(ctx, clientOpts)
    if err != nil {
        return nil, nil, fmt.Errorf("mongo.Connect: %w", err)
    }
    // Ping to verify connection
    ctxPing, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()
    if err := client.Ping(ctxPing, nil); err != nil {
        return nil, nil, fmt.Errorf("mongo ping: %w", err)
    }
    db := client.Database(dbName)
    
    // Create indexes
    if err := createIndexes(ctx, db); err != nil {
        log.Printf("[MongoDB] ⚠️  Failed to create indexes: %v\n", err)
    } else {
        log.Println("[MongoDB] ✅ Indexes created successfully")
    }
    
    return client, db, nil
}

// createIndexes creates necessary indexes for performance
func createIndexes(ctx context.Context, db *mongo.Database) error {
	// Users collection indexes
	usersCol := db.Collection("users")
	_, err := usersCol.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "uid", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "email", Value: 1}}},
	})
	if err != nil {
		return fmt.Errorf("users indexes: %w", err)
	}
	
	// Wallets collection indexes
	walletsCol := db.Collection("wallets")
	_, err = walletsCol.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "user_id", Value: 1}}, Options: options.Index().SetUnique(true)},
	})
	if err != nil {
		return fmt.Errorf("wallets indexes: %w", err)
	}
	
	// Transactions collection indexes
	transactionsCol := db.Collection("transactions")
	_, err = transactionsCol.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "created_at", Value: -1}}},
		{Keys: bson.D{{Key: "category", Value: 1}}},
	})
	if err != nil {
		return fmt.Errorf("transactions indexes: %w", err)
	}
	
	// Games collection indexes
	gamesCol := db.Collection("games")
	_, err = gamesCol.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "created_at", Value: -1}}},
		{Keys: bson.D{{Key: "game_type", Value: 1}}},
		{Keys: bson.D{{Key: "created_at", Value: -1}}},
	})
	if err != nil {
		return fmt.Errorf("games indexes: %w", err)
	}
	
	// Payment requests collection indexes
	paymentRequestsCol := db.Collection("payment_requests")
	_, err = paymentRequestsCol.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "created_at", Value: -1}}},
		{Keys: bson.D{{Key: "status", Value: 1}}},
	})
	if err != nil {
		return fmt.Errorf("payment_requests indexes: %w", err)
	}
	
	return nil
}

// UpsertUser inserts or updates a user document keyed by uid.
func UpsertUser(ctx context.Context, db *mongo.Database, uid, email, name string) error {
    col := db.Collection("users")
    now := time.Now().UTC()
    filter := bson.M{"uid": uid}
    update := bson.M{
        "$set": bson.M{
            "email":      email,
            "name":       name,
            "lastSeenAt": now,
        },
        "$setOnInsert": bson.M{
            "createdAt": now,
            "role":      "user",
        },
    }
    opts := options.Update().SetUpsert(true)
    _, err := col.UpdateOne(ctx, filter, update, opts)
    if err != nil {
        return fmt.Errorf("upsert user: %w", err)
    }
    return nil
}
