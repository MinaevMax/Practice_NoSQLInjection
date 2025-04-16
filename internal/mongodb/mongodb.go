package mongodb

import (
	"context"
	"log"
	"time"
    "os"
	"fmt"
	"math/rand"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/bson/primitive"
)

type Fine struct {
	ID    primitive.ObjectID `bson:"_id" json:"id"`
	Name  string             `bson:"name" json:"name"`
	Role  string             `bson:"role" json:"role"`
	Value int                `bson:"value" json:"value"`
}

func Init(collection *mongo.Collection) {
	// Подключение к MongoDB
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

	filter := bson.M{"role": "admin"}

    var result bson.M
    err := collection.FindOne(ctx, filter).Decode(&result)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            fine := Fine{
				ID:    primitive.NewObjectID(),
				Name:  os.Getenv("FLAG"),
				Role:  "admin",
				Value: 0,
			}
			_, err := collection.InsertOne(ctx, fine)
			if err != nil {
				log.Fatal(err)
			}
        } else {
            log.Fatal("Ошибка при поиске:", err)
        }
    }

	for i := 1; i < 15; i++ {
		name := fmt.Sprintf("Name%v", i)
		fine := Fine{
			ID:    primitive.NewObjectID(),
			Name:  name,
			Role:  "user",
			Value: rand.Intn(10000),
		}
		_, err := collection.InsertOne(ctx, fine)
		if err != nil {
			log.Fatal(err)
		}
	}
}