package controllers

import (
	"net/http"
	"ships/models"

	"github.com/gin-gonic/gin"
)

func RegisterGameRoutes(router *gin.Engine) {
	router.GET("/game/data", getGameData)

	router.GET("/game/userShips", getUserShips)

	router.GET("/game/getShips", getShips)
}

var canvasWidth = 3840
var canvasHeight = 2160
var guestsAllowed = true

type gameData struct {
	title         string `bson:"title"`
	username      string `bson:"username"`
	credits       int    `bson:"credits"`
	canvasWidth   int    `bson:"canvasWidth"`
	canvasHeight  int    `bson:"canvasHeight"`
	guestsAllowed bool   `bson:"guestsAllowed"`
}

func getGameData(c *gin.Context) {
	session := GetSessionIfExist(c)
	var user *models.User = nil
	if session != nil {
		user, _ = userDataAccess.GetUserByID(session.UserIdAsBsonObject())
	}
	data := gameData{
		title:         "Game",
		username:      "",
		credits:       0,
		canvasWidth:   canvasWidth,
		canvasHeight:  canvasHeight,
		guestsAllowed: guestsAllowed,
	}
	if user != nil {
		data.username = user.Username
		data.credits = user.Credits
	}

	c.IndentedJSON(http.StatusOK, data)
}

func getUserShips(c *gin.Context) {
	session := ValidateSession(c)
	if nil == session {
		c.IndentedJSON(http.StatusOK, gin.H{"success": true, "userShips": []models.Ship{}})
		return
	}

	projects, err := paintingBoardDataAccess.GetProjectsByUserId(session.UserId)

	if nil != err {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"success": false, "errors": []string{err.Error()}})
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"success": true, "userShips": projects})
}

func getShips(c *gin.Context) {
	ships, err := paintingBoardDataAccess.GetPublicShips()
	if nil != err {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"success": false, "errors": []string{err.Error()}})
	}
	c.IndentedJSON(http.StatusOK, ships)
}
