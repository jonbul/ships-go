package controllers

import (
	dataaccess "ships/dataAccess"
	"ships/models"

	"crypto/rand"
	"math/big"
	"time"

	"github.com/gin-gonic/gin"
)

const SessionDuration = 2592000000

var paintingBoardDataAccess = dataaccess.PaintingBoardDataAccess
var sessionDataAccess = dataaccess.SessionDataAccess
var userDataAccess = dataaccess.UserDataAccess

func RegisterRoutes(router *gin.Engine) {
	registerUserRoutes(router)
	registerPaintingBoardRoutes(router)
	registerGameRoutes(router)
	registerWebSocket(router)
	registerAdminRoutes(router)
	registerPrometheusRoutes(router)
}

func ValidateSession(c *gin.Context) *models.Session {
	cookie, err := c.Cookie("token")
	session, err := sessionDataAccess.GetSessionByToken(cookie)
	if nil != err || nil == session || session.IsExpired() {
		invalidateSession(c)
		return nil
	}
	RefreshToken(c, session)
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

func invalidateSession(c *gin.Context) {
	c.SetCookie("token", "", -1, "", "", true, true)
}

func RefreshToken(c *gin.Context, session *models.Session) {
	if !session.Persistent && !session.IsExpired() {
		session.Token = getNewBearerToken()
		session.ExpirationTime = time.Now().UnixMilli() + SessionDuration
		err := sessionDataAccess.UpdateSession(session)

		if nil == err {
			c.SetCookie("token", session.Token, int(session.ExpirationTime/1000), "", "", true, true)
		}
	}
}

func NewSession(userId string, admin bool, persistent bool) models.Session {
	session := models.Session{
		Admin:            admin,
		UserId:           userId,
		SessionTimeStamp: time.Now().UnixMilli(),
		Persistent:       persistent,
		Token:            getNewBearerToken(),
		LoggedOut:        false,
	}
	if persistent {
		session.ExpirationTime = time.Now().Add(365 * 24 * time.Hour).UnixMilli()
	} else {
		session.ExpirationTime = time.Now().UnixMilli() + SessionDuration
	}
	return session
}

func getNewBearerToken() string {
	const chars = "0123456789abcdefghijklmnopqrstuvwxyz"
	b := make([]byte, 60)
	for i := range b {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		b[i] = chars[n.Int64()]
	}
	return string(b)
}
