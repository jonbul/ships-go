package controllers

import (
	"log"
	"time"

	"net/http"

	"github.com/gin-gonic/gin"

	"ships/dataAccess"
)

type BodyUser struct {
	Email      string `json:"email"`
	Username   string `json:"username"`
	Password   string `json:"password"`
	Cpassword  string `json:"cpassword"`  // only register
	RememberMe bool   `json:"rememberMe"` // only login
}

func RegisterUserRoutes(router *gin.Engine) {
	router.GET("/status", getStatus)

	router.POST("/register", registerUser)

	router.POST("/login", loginUser)
	// Add more user-related routes here
}

func getStatus(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, gin.H{"status": "ok"})
}

func registerUser(c *gin.Context) {
	var user BodyUser
	err := c.BindJSON(&user)
	if nil != err {
		_ = c.BindJSON(gin.H{"error": "invalid request body"})
	}
	log.Printf("register user %s - %s", user.Email, user.Username)
	if user.Username == "" || user.Password == "" || user.Password != user.Cpassword {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
	}
	dbUser, err := dataaccess.CreateUser(user.Username, user.Email, user.Password)

	if nil != err || dbUser == nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
	}

	c.IndentedJSON(http.StatusOK, gin.H{"success": true})
}

func loginUser(c *gin.Context) {
	var user BodyUser
	err := c.BindJSON(&user)
	if nil != err {
		_ = c.BindJSON(gin.H{"error": "invalid request body"})
	}
	log.Printf("register user %s", user.Email)
	if user.Email == "" || user.Password == "" {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
	}

	var expirationTime = int64(-1)
	if !user.RememberMe {
		now := time.Now().UnixMilli()
		expirationTime = now + 30*24*3600000
	}

	dbUser, err := dataaccess.GetUserByEmailAndPassword(user.Email, user.Password)

	if nil != err || nil == dbUser {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
	}

	c.IndentedJSON(http.StatusOK, gin.H{
		"success":        true,
		"user":           dbUser,
		"expirationTime": expirationTime,
	})
}
