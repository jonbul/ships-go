package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"ships/controllers/websocket"
	"ships/controllers/websocket/models"
)

func registerAdminRoutes(router *gin.Engine) {
	router.POST("/gameData", getAdminGameData)
	router.GET("/game/admin/data", getAdminData)
	router.POST("/game/admin", postAdminData)
}

type resolutionData struct {
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Name   string `json:"name"`
}

var resolutions = []resolutionData{
	{Width: 1920, Height: 1080, Name: "FullHD (1920x1080)"},
	{Width: 2560, Height: 1440, Name: "QHD (2560x1440)"},
	{Width: 3840, Height: 2160, Name: "4K UHD (3840x2160)"},
}
var allowedPlayerTypes = map[string]int{
	"All":        0,
	"Registered": 1,
}
var currentResolution = 2
var allowedPlayerType = allowedPlayerTypes["All"]

func getAdminGameData(c *gin.Context) {
	session := ValidateSession(c)
	if nil == session {
		c.IndentedJSON(http.StatusUnauthorized, gin.H{"success": false, "errors": []string{"Unauthorized"}})
		return
	}

	user, err := userDataAccess.GetUserByID(session.UserIdAsBsonObject())
	if nil != err {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"success": false, "errors": []string{err.Error()}})
		return
	}
	if !user.Admin {
		c.IndentedJSON(http.StatusUnauthorized, gin.H{"success": false, "errors": []string{"Unauthorized"}})
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{
		"players":     websocket.SnapshotPlayers(),
		"resultCards": websocket.BackgroundCards,
	})
}

func getAdminData(c *gin.Context) {
	session := ValidateSession(c)
	if nil == session {
		c.IndentedJSON(http.StatusUnauthorized, gin.H{"success": false, "errors": []string{"Unauthorized"}})
		return
	}

	user, err := userDataAccess.GetUserByID(session.UserIdAsBsonObject())
	if nil != err {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"success": false, "errors": []string{err.Error()}})
		return
	}
	if !user.Admin {
		c.IndentedJSON(http.StatusUnauthorized, gin.H{"success": false, "errors": []string{"Unauthorized"}})
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{
		"resolutions":        resolutions,
		"currentResolution":  currentResolution,
		"allowedPlayerTypes": allowedPlayerTypes,
		"allowedPlayerType":  allowedPlayerType,
		"npcSettings":        websocket.GetNpcSettings(),
	})
}

func postAdminData(c *gin.Context) {
	session := ValidateSession(c)
	if nil == session {
		c.IndentedJSON(http.StatusUnauthorized, gin.H{"success": false, "errors": []string{"Unauthorized"}})
		return
	}

	user, err := userDataAccess.GetUserByID(session.UserIdAsBsonObject())
	if nil != err {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"success": false, "errors": []string{err.Error()}})
		return
	}
	if !user.Admin {
		c.IndentedJSON(http.StatusUnauthorized, gin.H{"success": false, "errors": []string{"Unauthorized"}})
		return
	}

	var body adminDataBody
	if err := c.BindJSON(&body); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"success": false, "errors": []string{err.Error()}})
		return
	}

	currentResolution, _ = strconv.Atoi(body.Resolution)
	allowedPlayerType, _ = strconv.Atoi(body.AllowedPlayerType)
	// Stored here but simulated in ships-npc, so this also pushes the new
	// values down its websocket; they take effect on its next tick, with
	// nothing to restart. ships-npc clamps them, so no validation here.
	if body.NpcSettings != nil {
		websocket.SetNpcSettings(*body.NpcSettings)
	}
	c.IndentedJSON(http.StatusOK, gin.H{"success": true, "npcSettings": websocket.GetNpcSettings()})

}

type adminDataBody struct {
	AllowedPlayerType string                  `json:"allowedPlayerType"`
	Resolution        string                  `json:"resolution"`
	NpcSettings       *models.NpcSettingsData `json:"npcSettings"`
}
