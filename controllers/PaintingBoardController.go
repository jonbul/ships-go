package controllers

import (
	"net/http"
	"ships/dataAccess"
	"ships/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func RegisterPaintingBoardRoutes(router *gin.Engine) {
	router.GET("/paintingBoard/projects/all", getProjects)

	router.POST("/paintingBoard/save", saveProject)

	router.DELETE("/paintingBoard/projects", deleteProject)

	router.GET("/paintingBoard/projects/id", getProject)

}

var paintingBoardDataAccess = dataaccess.PaintingBoardDataAccess

type Project struct {
	Id      bson.ObjectID          `json:"_id" bson:"_id,omitempty"`
	Project models.PaintingProject `json:"project" bson:"project"`
}

func getProjects(c *gin.Context) {
	session := ValidateSession(c)
	if nil == session {
		return
	}

	projects, _ := paintingBoardDataAccess.GetProjectsByUserId(session.UserId)
	c.IndentedJSON(http.StatusOK, projects)
}

func saveProject(c *gin.Context) {
	session := ValidateSession(c)
	// boddy as string
	if nil == session {
		return
	}
	user, err := userDataAccess.GetUserByID(session.UserIdAsBsonObject())

	if nil != err || nil == user {
		InvalidateSession(c)
		return
	}

	var project Project
	err = c.BindJSON(&project)

	if err != nil {
		c.IndentedJSON(http.StatusUnauthorized, gin.H{"success": false, "errors": []string{err.Error()}})
		return
	}
	project.Project.UserId = session.UserId
	if project.Project.Id == bson.NilObjectID {
		err = paintingBoardDataAccess.SaveProject(&project.Project)
	} else {
		err = paintingBoardDataAccess.UpdateProject(&project.Project)
	}
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"success": false, "errors": []string{err.Error()}})
	}
	c.IndentedJSON(http.StatusOK, gin.H{"success": true, "id": project.Project.Id})
}

func deleteProject(c *gin.Context) {
	session := ValidateSession(c)
	if nil == session {
		return
	}

	id := c.Query("id")

	if id != "" {
		projectId, _ := bson.ObjectIDFromHex(id)
		err := paintingBoardDataAccess.DeleteProjectById(projectId)
		if nil == err {
			c.IndentedJSON(http.StatusOK, gin.H{"success": true})
			return
		}
		c.IndentedJSON(http.StatusBadRequest, gin.H{"errors": []string{"invalid request body"}})
	}
}

func getProject(c *gin.Context) {
	session := ValidateSession(c)
	if nil == session {
		return
	}

	id := c.Query("id")

	if id != "" {
		projectId, _ := bson.ObjectIDFromHex(id)
		project, err := paintingBoardDataAccess.GetProjectById(projectId)
		if nil == err {
			c.IndentedJSON(http.StatusOK, project)
			return
		}
		c.IndentedJSON(http.StatusBadRequest, gin.H{"errors": []string{"invalid request body"}})
	}
}
