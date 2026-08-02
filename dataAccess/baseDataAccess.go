package dataaccess

import (
	"context"
	"log"
	"os"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var MongoUri = os.Getenv("MONGODB_URI")

type collectionNames struct {
}

func (collectionNames) users() string {
	return "users"
}
func (collectionNames) sessions() string {
	return "sessions"
}
func (collectionNames) ships() string {
	return "ships"
}
func (collectionNames) paintingProjects() string {
	return "paintingprojects"
}

var CollectionNames = collectionNames{}

func init() {
	log.Default()
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	MongoUri = os.Getenv("MONGODB_URI")
	if MongoUri == "" {
		log.Fatal("MONGODB_URI is not set in the environment variables")
	}
	log.Println("MongoUri loaded in DataAccess: " + MongoUri[:4] + "...")
}

type baseDataAccess struct{}

var BaseDataAccess = baseDataAccess{}

func (baseDataAccess) getClient() *mongo.Client {
	bsonOpts := &options.BSONOptions{
		AllowTruncatingDoubles: true,
	}

	client, err := mongo.Connect(
		options.Client().
			ApplyURI(MongoUri).
			SetBSONOptions(bsonOpts))
	if err != nil && nil != client {
		err := client.Disconnect(context.TODO())
		if err != nil {
			return nil
		}
		log.Fatal("Error connecting to MongoDB:", err)
	}
	if nil == client {
		log.Fatal("Error connecting to MongoDB")
	}
	return client
}

func (baseDataAccess) getCollection(client *mongo.Client, collectionName string) *mongo.Collection {
	return client.Database("jaes", nil).Collection(collectionName)
}

func (da baseDataAccess) ExecuteSecurely(collectionName string, method func(mongo.Collection) error) error {
	mongoClient := da.getClient()
	collection := da.getCollection(mongoClient, collectionName)
	err1 := method(*collection)
	_ = mongoClient.Disconnect(context.TODO())
	return err1
}
