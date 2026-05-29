package dataaccess

import (
	"context"
	"errors"
	"log"
	"ships/models"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"golang.org/x/crypto/bcrypt"
)

func GetUserByUsername(username string) (*models.User, error) {
	var user models.User

	error := ExecuteSecurely(CollectionNames.users(), func(collection mongo.Collection) error {
		return collection.FindOne(nil, bson.D{{Key: "username", Value: username}}).Decode(&user)
	})
	user.Password = ""
	return &user, error
}

func GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	err := ExecuteSecurely(CollectionNames.users(), func(collection mongo.Collection) error {
		return collection.FindOne(nil, bson.D{{Key: "email", Value: email}}).Decode(&user)
	})
	user.Password = ""
	return &user, err
}

func GetUserByID(id bson.ObjectID) (*models.User, error) {
	var user models.User
	err := ExecuteSecurely(CollectionNames.users(), func(collection mongo.Collection) error {
		return collection.FindOne(nil, bson.D{{Key: "_id", Value: id}}).Decode(&user)
	})
	user.Password = ""
	return &user, err
}

func GetUserByEmailAndPassword(email string, password string) (*models.User, error) {
	var user *models.User
	err := ExecuteSecurely(CollectionNames.users(), func(collection mongo.Collection) error {
		return collection.FindOne(nil, bson.D{{Key: "email", Value: email}}).Decode(&user)
	})
	if err != nil || nil == user {
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, errors.New("Invalid password or username provided")
	}
	user.Password = ""
	return user, nil
}

func CreateUser(username string, email string, password string) (*models.User, error) {
	registeredUser, err := GetUserByEmail(email)
	if nil == err && registeredUser == nil {
		registeredUser, err = GetUserByUsername(username)
	}

	if err == nil { // User or email already exist
		return nil, errors.New("User already exists")
	}

	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user := models.User{
		Admin:    false,
		Username: username,
		Email:    email,
		Password: string(hashedPwd),
		Credits:  0,
		Kills:    0,
		Deaths:   0,
	}

	err = ExecuteSecurely(CollectionNames.users(), func(collection mongo.Collection) error {
		_, err := collection.InsertOne(context.TODO(), user)
		return err
	})

	if nil != err {
		log.Fatal("Error inserting user:", err)
		return nil, err
	}
	return &user, nil
}

func UpdateUser(user models.User) error {
	return ExecuteSecurely(CollectionNames.users(), func(collection mongo.Collection) error {
		_, err := collection.UpdateByID(context.TODO(), user.Id, bson.D{{Key: "$set", Value: user}})
		return err
	})
}

func DeleteUserById(id string) error {
	return ExecuteSecurely(CollectionNames.users(), func(collection mongo.Collection) error {
		_, error := collection.DeleteOne(context.TODO(), bson.D{{Key: "_id", Value: id}})
		return error
	})
}
