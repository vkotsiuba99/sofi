package internal

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"time"
)

const mongoURI = "mongodb://database:27017"

type Database struct {
	client *mongo.Client
	db     *mongo.Database
}

func NewDatabase() *Database {
	return &Database{}
}

func (d *Database) Connect() error {
	client, err := mongo.NewClient(options.Client().ApplyURI(mongoURI))
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	if err != nil {
		return err
	}

	err = client.Ping(context.Background(), readpref.Primary())
	if err != nil {
		return err
	}

	d.client = client
	return nil
}

func (d *Database) InitDatabase() error {
	db := d.client.Database("sofi")
	coll := db.Collection("logs")
	if coll == nil {
		err := db.CreateCollection(context.TODO(), "logs")
		if err != nil {
			return err
		}
	}

	d.db = db
	return nil
}

func (d *Database) Insert(log interface{}) (interface{}, error) {
	logs := d.db.Collection("logs")
	result, err := logs.InsertOne(context.TODO(), log)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (d *Database) Disconnect() {
	d.client.Disconnect(context.Background())
}
