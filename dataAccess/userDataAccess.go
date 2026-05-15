package dataaccess

import (
	"context"
	"log"
	"os"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"golang.org/x/crypto/bcrypt"

	"ships/models"
)

var MongoUri = os.Getenv("MONGODB_URI")

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

func Init() {

	user, err := GetUserByUsername("jonbul")

	log.Println("------------------------------")
	if err != nil && err != mongo.ErrNoDocuments {
		log.Fatal("Error finding user:", err)
	} else if err == mongo.ErrNoDocuments {
		log.Println("Connection works but no user found with username:", "jonbul")
	} else {
		log.Printf("User found with username: %s, email: %s", user.Username, user.Email)
	}
	log.Println("------------------------------")

}

func getCollection() *mongo.Collection {
	client, err := mongo.Connect(options.Client().ApplyURI(MongoUri))
	if err != nil {
		client.Disconnect(context.TODO())
		log.Fatal("Error connecting to MongoDB:", err)
	}
	return client.Database("jaes", nil).Collection("users")
}

func GetUserByUsername(username string) (*models.User, error) {
	var user models.User
	coll := getCollection()
	err := coll.FindOne(context.TODO(), bson.D{{Key: "username", Value: username}}).Decode(&user)
	user.Password = ""
	return &user, err
}

func GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	coll := getCollection()
	err := coll.FindOne(context.TODO(), bson.D{{Key: "email", Value: email}}).Decode(&user)
	return &user, err
}

func GetUserByEmailAndPassword(email string, password string) (*models.User, error) {
	var user, err = GetUserByEmail(email)
	if err != nil {
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, err
	}

	return user, nil
}

func GetUserByID(id string) (*models.User, error) {
	var user models.User
	coll := getCollection()
	err := coll.FindOne(context.TODO(), bson.D{{Key: "_id", Value: id}}).Decode(&user)
	return &user, err
}

func CreateUser(username string, email string, password string) (*models.User, error) {
	coll := getCollection()
	hasehdPwd, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user := models.User{
		Admin:    false,
		Username: username,
		Email:    email,
		Password: string(hasehdPwd),
		Credits:  0,
		Kills:    0,
		Deaths:   0,
	}
	_, err = coll.InsertOne(context.TODO(), user)
	return &user, err
}

func DeleteUser(id string) error {
	coll := getCollection()
	_, err := coll.DeleteOne(context.TODO(), bson.D{{Key: "_id", Value: id}})
	return err
}
