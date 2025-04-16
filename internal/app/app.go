
package app

import (
	"sync"
	"log"
	"os"
	"context"
	"time"
	"nosqli/internal/httpServer"
	"nosqli/internal/mongodb"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func ConnectWithRetry(uri string) (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(uri)
	
	// Настройки повторных попыток
	retryCount := 0
	maxRetries := 10
	retryInterval := 5 * time.Second

	for {
		client, err := mongo.Connect(ctx, clientOptions)
		if err == nil {
			// Проверка соединения
			err = client.Ping(ctx, nil)
			if err == nil {
				return client, nil
			}
		}

		retryCount++
		if retryCount > maxRetries {
			return nil, err
		}

		log.Printf("Connection attempt %d/%d failed. Retrying in %v...", 
			retryCount, maxRetries, retryInterval)
		time.Sleep(retryInterval)
	}
}

func Run() error {
	//db.Add()
	var wg sync.WaitGroup

	// Подключение к MongoDB
	client, err := ConnectWithRetry(os.Getenv("MONGODB_URI"))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(context.TODO())

	collection := client.Database("fines_db").Collection("fines")

	mongodb.Init(collection)

	wg.Add(1)
	httpServer.Start(&wg, collection)
	wg.Wait()
	return nil
}
