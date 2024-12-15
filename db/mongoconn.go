package db

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var MessagesCollection1 *mongo.Collection // Exposed for use in other files

func InitMongoDB2() {
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI("mongodb://umersheraxk:3FMqbxdwsHJxqdvn@drawzycluster-shard-00-00.2kfwi.mongodb.net:27017,drawzycluster-shard-00-01.2kfwi.mongodb.net:27017,drawzycluster-shard-00-02.2kfwi.mongodb.net:27017/DATABASE?ssl=true&authSource=admin").SetServerAPIOptions(serverAPI)
	client, err := mongo.Connect(context.TODO(), opts)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err = client.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()
	if err := client.Database("admin").RunCommand(context.TODO(), bson.D{{Key: "ping", Value: 1}}).Err(); err != nil {
		panic(err)
	}

	MessagesCollection1 = client.Database("draw_game").Collection("messages")

	fmt.Println("Pinged your deployment. You successfully connected to MongoDB!")
}

func InitMongoDB() {
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI("mongodb://umersheraxk:3FMqbxdwsHJxqdvn@drawzycluster-shard-00-00.2kfwi.mongodb.net:27017,drawzycluster-shard-00-01.2kfwi.mongodb.net:27017,drawzycluster-shard-00-02.2kfwi.mongodb.net:27017/DATABASE?ssl=true&authSource=admin").SetServerAPIOptions(serverAPI)

	client, err := mongo.Connect(context.TODO(), opts)
	if err != nil {
		panic(fmt.Sprintf("Error connecting to MongoDB: %v", err))
	}

	// Ping MongoDB to ensure connectivity
	if err := client.Database("admin").RunCommand(context.TODO(), bson.D{{Key: "ping", Value: 1}}).Err(); err != nil {
		panic(fmt.Sprintf("Error pinging MongoDB: %v", err))
	}

	// Set the messages collection
	MessagesCollection1 = client.Database("draw_game").Collection("messages")
	fmt.Println("Connected to MongoDB and initialized 'messages' collection")
}
