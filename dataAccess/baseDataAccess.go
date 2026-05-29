package dataaccess

import (
	"context"
	"errors"
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
	return "paintingProjects"
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

func getClient() *mongo.Client {
	client, err := mongo.Connect(options.Client().ApplyURI(MongoUri))
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

func getCollection(client *mongo.Client, collectionName string) *mongo.Collection {
	return client.Database("jaes", nil).Collection(collectionName)
}

func ExecuteSecurely(collectionName string, method func(mongo.Collection) error) error {
	mongoClient := getClient()
	collection := getCollection(mongoClient, collectionName)
	res := method(*collection)
	mongoClient.Disconnect(context.TODO())
	return res
}

func Test() {

	user, err := GetUserByUsername("test")

	log.Println("------------------------------")
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		log.Fatal("Error finding user:", err)
	} else if errors.Is(err, mongo.ErrNoDocuments) {
		log.Println("Connection works but no user found with username:", "jonbul")
	} else if nil != user && nil == err {
		log.Printf("User found with username: %s, email: %s\n", user.Username, user.Email)
	} else if nil != err {
		log.Printf("Something happened: %s\n", err.Error())
	} else {
		log.Println("¯\\_(ツ)_/¯")
	}
	log.Println("------------------------------")

}
