package controllers

import (
	dataaccess "ships/dataAccess"
	"ships/models"

	"github.com/gin-gonic/gin"
)

var paintingBoardDataAccess = dataaccess.PaintingBoardDataAccess
var sessionDataAccess = dataaccess.SessionDataAccess
var userDataAccess = dataaccess.UserDataAccess

func ValidateSession(c *gin.Context) *models.Session {
	cookie, err := c.Cookie("token")
	session, err := sessionDataAccess.GetSessionByToken(cookie)
	if nil != err || nil == session || session.IsExpired() {
		InvalidateSession(c)
		return nil
	}
	return session
}

func GetSessionIfExist(c *gin.Context) *models.Session {
	cookie, _ := c.Cookie("token")
	if "" != cookie {
		session, _ := sessionDataAccess.GetSessionByToken(cookie)
		return session
	}
	return nil
}

func InvalidateSession(c *gin.Context) {
	c.SetCookie("token", "", -1, "", "", true, true)
}
