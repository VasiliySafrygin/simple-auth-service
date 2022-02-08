package api

import (
	"auth-service/db"
	"auth-service/function"
	"auth-service/models"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Token  string  `json:"token"`
	Error  error   `json:"error"`
	Status bool        `json:"status"`
	User   models.User `json:"user"`
}

func About(c *gin.Context) {
	c.JSON(http.StatusOK, "OK")
}

func timing(start time.Time, comment string)  {
	t := time.Now()
	fmt.Fprintf(os.Stdout, "Elapsed Login: %v %s\n", t.Sub(start), comment)
}

func Login(c *gin.Context) {
	start := time.Now()
	response := Response{}
	response.Status = false
	var u models.User
	if err := c.ShouldBindJSON(&u); err != nil {
		c.JSON(http.StatusUnprocessableEntity, "Invalid json provided")
		return
	}
	dbUser, err := db.GetUserByName(u.UserName)
	timing(start, "GetUserByName")
	if err != nil {
		response.Error = err
		c.JSON(http.StatusUnprocessableEntity, response)
		return
	}
	if dbUser.UserName != u.UserName || !function.ComparePassword(dbUser.Password, u.Password) {
		response.Error = errors.New("Логин или пароль не совпадают. ")
		c.JSON(http.StatusForbidden, response)
		return
	}
	timing(start, "Verify")
	token, err := db.CreateJwtToken(dbUser.Id)
	if err != nil {
		response.Error = err
		c.JSON(http.StatusUnprocessableEntity, response)
		return
	}
	timing(start, "CreateJwtToken")
	response.Token = token
	response.Status = true
	response.User.Id = dbUser.Id
	response.User.UserName = dbUser.UserName
	response.User.FirstName = dbUser.FirstName
	response.User.LastName = dbUser.LastName
	response.User.MiddleName = dbUser.MiddleName
	timing(start, "Response")
	c.JSON(http.StatusOK, response)
}

func Logout(c *gin.Context) {
	response := Response{}
	response.Status = true

	userId, exist := c.Get("userId")
	if exist {
		err := db.DeleteToken(userId.(string))
		if err != nil {
			response.Error = err
			c.JSON(http.StatusUnprocessableEntity, response)
			return
		}
	}
	c.JSON(http.StatusOK, response)
}

func User(c *gin.Context) {
	response := Response{}
	response.Status = false
	username, _ := c.GetQuery("username")
	dbUser, err := db.GetUserByName(username)
	if err != nil {
		response.Error = err
		c.JSON(http.StatusUnprocessableEntity, response)
		return
	}
	response.Status = true
	response.User.Id = dbUser.Id
	response.User.UserName = dbUser.UserName
	response.User.FirstName = dbUser.FirstName
	response.User.LastName = dbUser.LastName
	response.User.MiddleName = dbUser.MiddleName
	c.JSON(http.StatusOK, response)
}

func Create(c *gin.Context) {
	var u models.User
	if err := c.ShouldBindJSON(&u); err != nil {
		c.JSON(http.StatusUnprocessableEntity, "Invalid json provided")
		return
	}
	fmt.Fprintf(os.Stdout, u.UserName, u.Password)
	user, err := db.CreateUser(u.UserName, u.FirstName, u.LastName, u.MiddleName, u.Password)
	c.JSON(http.StatusOK, gin.H{"result": user, "error": err})
}
