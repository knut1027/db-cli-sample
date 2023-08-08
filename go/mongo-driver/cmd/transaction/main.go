package main

import (
	"context"
	"flag"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
	"go.uber.org/zap"
	"time"
)

type Book struct {
	ID     string `bson:"_id,omitempty"`
	Title  string `bson:"title"`
	Author string `bson:"author"`
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27018/?connect=direct"))
	if err != nil {
		panic("cannot connect to mongo")
	}

	logger, _ := zap.NewProduction()
	defer logger.Sync()

	deleted := flag.Bool("deleted", false, "delete all documents in the collection")
	flag.Parse()
	if *deleted {
		col := client.Database("test").Collection("bookInfo")
		col.DeleteMany(ctx, bson.M{})
		logger.Info("deleted")
	}

	session, err := client.StartSession()
	if err != nil {
		panic(err)
	}
	defer session.EndSession(context.TODO())

	book1 := Book{
		ID:     "1",
		Title:  "The Bluest Eye",
		Author: "Toni Morrison",
	}
	book2 := Book{
		ID:     "2",
		Title:  "Sula",
		Author: "Toni Morrison",
	}
	book3 := Book{
		ID:     "3",
		Title:  "Song of Solomon",
		Author: "Toni Morrison",
	}

	c := BookClient{
		cli:    client,
		logger: logger,
	}
	_, err = c.Transact(ctx, func(ctx context.Context) (interface{}, error) {
		err := c.InsertMany(ctx, []Book{book1, book2, book3})
		return nil, err
	})
	if err != nil {
		logger.Error("failed to insert", zap.Error(err))
	}
}

type BookClient struct {
	cli    *mongo.Client
	logger *zap.Logger
}

func (c *BookClient) InsertMany(ctx context.Context, books []Book) error {
	col := c.cli.Database("test").Collection("bookInfo")

	c.logger.Info("insert many...")
	for i := range books {
		_, err := col.InsertOne(ctx, books[i])
		if err != nil {
			return err
		}
		time.Sleep(2 * time.Second)
	}
	return nil
}

func (c *BookClient) Transact(ctx context.Context, f func(context.Context) (interface{}, error)) (interface{}, error) {
	wc := writeconcern.New(writeconcern.WMajority())
	txnOptions := options.Transaction().SetWriteConcern(wc)

	session, err := c.cli.StartSession()
	if err != nil {
		return nil, fmt.Errorf("failed to start session: %w", err)
	}
	c.logger.Info("start session")
	defer func() {
		c.logger.Info("end session")
		session.EndSession(context.Background())
	}()
	result, err := session.WithTransaction(
		ctx,
		func(sessCtx mongo.SessionContext) (interface{}, error) {
			return f(sessCtx)
		},
		txnOptions,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to transact: %w", err)
	}
	return result, nil
}
