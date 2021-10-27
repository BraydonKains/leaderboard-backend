package handlers

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/speedrun-website/leaderboard-backend/auth"
	"github.com/speedrun-website/leaderboard-backend/data"
)

type UserIdentifierResponse struct {
	User *data.UserIdentifier `json:"user"`
}

type UserPersonalResponse struct {
	User *data.UserPersonal `json:"user"`
}

func GetUser(c *gin.Context) {
	// Maybe we shouldn't use the increment ID but generate a UUID instead to avoid
	// exposing the amount of users registered in the database.
	id, err := strconv.ParseUint(c.Param("id"), 10, 0)

	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	user, err := data.Users.GetUserIdentifierById(id)

	if err != nil {
		var code int
		if errors.Is(err, data.UserNotFoundError{ID: id}) {
			code = http.StatusNotFound
		} else {
			code = http.StatusInternalServerError
		}

		c.AbortWithStatusJSON(code, ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, UserIdentifierResponse{
		User: user,
	})
}

type RegisterUserRequest struct {
	Username        string `json:"username" binding:"required"`
	Email           string `json:"email" binding:"required,email"`
	Password        string `json:"password" binding:"required,min=8"`
	PasswordConfirm string `json:"password_confirm" binding:"eqfield=Password"`
}

func RegisterUser(c *gin.Context) {
	var registerValue RegisterUserRequest
	if err := c.BindJSON(&registerValue); err != nil {
		log.Println("Unable to bind value", err)
		return
	}

	hash, err := auth.HashAndSalt([]byte(registerValue.Password))
	if err != nil {
		log.Println(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	user := data.User{
		Username: registerValue.Username,
		Email:    registerValue.Email,
		Password: hash,
	}

	err = data.Users.CreateUser(user)

	if err != nil {
		if uniquenessErr, ok := err.(data.UserUniquenessError); ok {
			/*
			 * TODO: we probably don't want to reveal if an email is already in use.
			 * Maybe just give a 201 and send an email saying that someone tried to sign up as you.
			 * --Ted W
			 *
			 * I still think we should do as above, but for my refactor 2021/10/22 I left
			 * what was already here.
			 * --RageCage
			 */
			c.AbortWithStatusJSON(http.StatusConflict, ErrorResponse{
				Error: uniquenessErr.Error(),
			})
		} else {
			c.AbortWithStatus(http.StatusInternalServerError)
		}

		return
	}

	c.Header("Location", fmt.Sprintf("/api/v1/users/%d", user.ID))
	c.JSON(http.StatusCreated, UserIdentifierResponse{
		User: &data.UserIdentifier{
			ID:       user.ID,
			Username: user.Username,
		},
	})
}

func Me(c *gin.Context) {
	rawUser, ok := c.Get(auth.JwtConfig.IdentityKey)
	if ok {
		user, ok := rawUser.(*data.UserPersonal)
		if ok {
			userInfo, err := data.Users.GetUserPersonalById(uint64(user.ID))

			if err == nil {
				c.JSON(http.StatusOK, UserPersonalResponse{
					User: userInfo,
				})
			}
		}
	}

	c.AbortWithStatus(http.StatusInternalServerError)
}
