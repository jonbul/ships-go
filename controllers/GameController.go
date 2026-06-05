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
	router.GET("/game/getPlayers", getPlayers)
}

var canvasWidth = 3840
var canvasHeight = 2160
var guestsAllowed = true

type gameData struct {
	Title         string `json:"title" bson:"title"`
	Username      string `json:"username" bson:"username"`
	Credits       int    `json:"credits" bson:"credits"`
	CanvasWidth   int    `json:"canvasWidth" bson:"canvasWidth"`
	CanvasHeight  int    `json:"canvasHeight" bson:"canvasHeight"`
	GuestsAllowed bool   `json:"guestsAllowed" bson:"guestsAllowed"`
}

func getGameData(c *gin.Context) {
	session := GetSessionIfExist(c)
	var user *models.User = nil
	if session != nil {
		user, _ = userDataAccess.GetUserByID(session.UserIdAsBsonObject())
	}
	data := gameData{
		Title:         "Game",
		Username:      "",
		Credits:       0,
		CanvasWidth:   canvasWidth,
		CanvasHeight:  canvasHeight,
		GuestsAllowed: guestsAllowed,
	}
	if user != nil {
		data.Username = user.Username
		data.Credits = user.Credits
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

func getPlayers(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, players)
}
