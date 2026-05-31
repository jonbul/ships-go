package controllers

import (
	"log"
	"time"

	"net/http"

	"github.com/gin-gonic/gin"
)

type BodyUser struct {
	Email      string `json:"email"`
	Username   string `json:"username" default:""`
	Password   string `json:"password"`
	Cpassword  string `json:"cpassword" default:""`  // only register
	RememberMe bool   `json:"rememberMe" default:""` // only login
}

func RegisterUserRoutes(router *gin.Engine) {
	router.GET("/status", getStatus)

	router.POST("/register", registerUser)

	router.POST("/login", loginUser)

	router.GET("/userInfo", userInfo)

	router.POST("/logout", logout)

	router.POST("/refreshToken", refreshToken)
}

func getStatus(c *gin.Context) {

	c.IndentedJSON(http.StatusOK, gin.H{"status": "ok"})
}

func userInfo(c *gin.Context) {

	session := GetSessionIfExist(c)
	if nil == session || session.IsExpired() {
		return
	}

	user, err := userDataAccess.GetUserByID(session.UserIdAsBsonObject())
	if err != nil {

	}
	c.IndentedJSON(http.StatusOK, user)
}

func registerUser(c *gin.Context) {
	var user BodyUser
	err := c.BindJSON(&user)
	if nil != err {
		_ = c.BindJSON(gin.H{"errors": []string{"invalid request body"}})
	}
	log.Printf("register user %s - %s", user.Email, user.Username)
	if user.Username == "" || user.Password == "" || user.Password != user.Cpassword {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"errors": []string{"invalid request body"}})
		return
	}
	dbUser, err := userDataAccess.CreateUser(user.Username, user.Email, user.Password)

	if nil != err || dbUser == nil {
		var text = "Something happened while creating user."
		if nil != err {
			text = err.Error()
		}
		c.IndentedJSON(http.StatusBadRequest, gin.H{"success": false, "errors": []string{text}})
	} else {
		c.IndentedJSON(http.StatusOK, gin.H{"success": true})
	}
}

func loginUser(c *gin.Context) {
	var user BodyUser
	err := c.BindJSON(&user)
	if nil != err {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"errors": []string{"invalid request body"}})
		return
	}
	log.Printf("login user %s", user.Email)
	if user.Email == "" || user.Password == "" {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"errors": []string{"Empty fields"}})
		return
	}

	var expirationTime = int64(-1)
	if !user.RememberMe {
		now := time.Now().UnixMilli()
		expirationTime = now + 30*24*3600000
	}

	dbUser, err := userDataAccess.GetUserByEmailAndPassword(user.Email, user.Password)

	if nil != err || nil == dbUser {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"errors": []string{"invalid request body"}})
		return
	}

	session := NewSession(dbUser.IdAsString(), dbUser.Admin, user.RememberMe)
	err = sessionDataAccess.InsertSession(session)
	if err != nil {
		log.Fatal(err)
	}
	//SetCookie(name, value string, maxAge int, path, domain string, secure, httpOnly bool) {
	c.SetCookie("token", session.Token, int(expirationTime/1000), "", "", true, true)

	c.IndentedJSON(http.StatusOK, gin.H{
		"success":        true,
		"user":           dbUser,
		"expirationTime": expirationTime,
	})
}

func logout(c *gin.Context) {
	session := GetSessionIfExist(c)
	if nil == session {
		c.IndentedJSON(http.StatusOK, gin.H{"success": true})
		return
	}
	session.Persistent = false
	session.LoggedOut = true
	err := sessionDataAccess.UpdateSession(session)
	if nil != err {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"success": false})
		return
	}
	invalidateSession(c)
	c.IndentedJSON(http.StatusOK, gin.H{"success": true})
}

func refreshToken(c *gin.Context) {
	session := GetSessionIfExist(c)
	if nil == session {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"success": false})
		return
	}
	RefreshToken(c, session)
	c.IndentedJSON(http.StatusOK, gin.H{"success": true, "expirationTime": session.ExpirationTime})
}
