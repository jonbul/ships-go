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

type UserDataAccessType struct {
	*baseDataAccess
}

var UserDataAccess = UserDataAccessType{
	baseDataAccess: &BaseDataAccess,
}

func (dataAccess UserDataAccessType) GetUserByUsername(username string) (*models.User, error) {
	var user models.User

	err := dataAccess.ExecuteSecurely(CollectionNames.users(), func(collection mongo.Collection) error {
		return collection.FindOne(context.TODO(), bson.D{{Key: "username", Value: username}}).Decode(&user)
	})
	user.Password = ""
	return &user, err
}

func (dataAccess UserDataAccessType) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	err := dataAccess.ExecuteSecurely(CollectionNames.users(), func(collection mongo.Collection) error {
		return collection.FindOne(context.TODO(), bson.D{{Key: "email", Value: email}}).Decode(&user)
	})
	user.Password = ""
	return &user, err
}

func (dataAccess UserDataAccessType) GetUserByID(id bson.ObjectID) (*models.User, error) {
	var user models.User
	err := dataAccess.ExecuteSecurely(CollectionNames.users(), func(collection mongo.Collection) error {
		return collection.FindOne(context.TODO(), bson.D{{Key: "_id", Value: id}}).Decode(&user)
	})
	user.Password = ""
	return &user, err
}

func (dataAccess UserDataAccessType) GetUserByEmailAndPassword(email string, password string) (*models.User, error) {
	var user *models.User
	err := dataAccess.ExecuteSecurely(CollectionNames.users(), func(collection mongo.Collection) error {
		return collection.FindOne(nil, bson.D{{Key: "email", Value: email}}).Decode(&user)
	})
	if err != nil || nil == user {
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, errors.New("invalid password or username provided")
	}
	user.Password = ""
	return user, nil
}

func (dataAccess UserDataAccessType) CreateUser(username string, email string, password string) (*models.User, error) {
	registeredUser, err := dataAccess.GetUserByEmail(email)
	if nil == err && registeredUser == nil {
		registeredUser, err = dataAccess.GetUserByUsername(username)
	}

	if err == nil { // User or email already exist
		return nil, errors.New("user already exists")
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

	err = dataAccess.ExecuteSecurely(CollectionNames.users(), func(collection mongo.Collection) error {
		_, err := collection.InsertOne(context.TODO(), user)
		return err
	})

	if nil != err {
		log.Fatal("Error inserting user:", err)
		return nil, err
	}
	return &user, nil
}

func (dataAccess UserDataAccessType) UpdateUser(user models.User) error {
	return dataAccess.ExecuteSecurely(CollectionNames.users(), func(collection mongo.Collection) error {
		_, err := collection.UpdateByID(context.TODO(), user.Id, bson.D{{Key: "$set", Value: user}})
		return err
	})
}

func (dataAccess UserDataAccessType) DeleteUserById(id string) error {
	return dataAccess.ExecuteSecurely(CollectionNames.users(), func(collection mongo.Collection) error {
		_, err := collection.DeleteOne(context.TODO(), bson.D{{Key: "_id", Value: id}})
		return err
	})
}

func Test() {

	user, err := UserDataAccess.GetUserByUsername("test")

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
