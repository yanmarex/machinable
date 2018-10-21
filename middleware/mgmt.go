package middleware

import (
	"net/http"
	"strings"

	"bitbucket.org/nsjostrom/machinable/auth"
	"bitbucket.org/nsjostrom/machinable/management/database"
	"bitbucket.org/nsjostrom/machinable/management/models"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/bson/objectid"
)

func respondWithError(code int, message string, c *gin.Context) {
	resp := map[string]string{"error": message}

	c.JSON(code, resp)
	c.Abort()
}

// AppUserProjectAuthzMiddleware validates this app user has access to the project
func AppUserProjectAuthzMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if values, _ := c.Request.Header["Authorization"]; len(values) > 0 {

			tokenString := strings.Split(values[0], " ")[1]
			token, err := jwt.Parse(tokenString, auth.TokenLookup)

			if err == nil {
				// get project from context, inserted into context from subdomain
				project := c.GetString("project")
				if project == "" {
					respondWithError(http.StatusUnauthorized, "invalid project", c)
					return
				}

				// token is valid, get claims and perform authorization
				claims := token.Claims.(jwt.MapClaims)

				// get list of users' projects from claims
				projects, ok := claims["projects"].(map[string]interface{})
				if !ok {
					respondWithError(http.StatusUnauthorized, "improperly formatted access token", c)
					return
				}

				// get user from claims
				user, ok := claims["user"].(map[string]interface{})
				if !ok {
					respondWithError(http.StatusUnauthorized, "improperly formatted access token", c)
					return
				}

				_, ok = projects[project]
				if !ok {
					// the project is not in the claims, look in the database in case it was created with the last 5 minutes

					// create ObjectID from UserID string
					userObjectID, err := objectid.FromHex(user["id"].(string))
					if err != nil {
						respondWithError(http.StatusUnauthorized, "improperly formatted access token", c)
						return
					}
					// get the project collection
					col := database.Connect().Collection(database.Projects)

					// look up the user
					documentResult := col.FindOne(
						nil,
						bson.NewDocument(
							bson.EC.String("slug", project),
							bson.EC.ObjectID("user_id", userObjectID),
						),
						nil,
					)

					prj := &models.Project{}
					// decode user document
					err = documentResult.Decode(prj)
					if err != nil {
						respondWithError(http.StatusNotFound, "project not found", c)
						return
					}

					// project was found, continue request
					c.Next()
					return
				}
			}

			respondWithError(http.StatusUnauthorized, "invalid access token", c)
			return
		}

		respondWithError(http.StatusUnauthorized, "access token required", c)
		return
	}
}

// AppUserJwtAuthzMiddleware authorizes the JWT in the Authorization header for application users
func AppUserJwtAuthzMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		if values, _ := c.Request.Header["Authorization"]; len(values) > 0 {

			tokenString := strings.Split(values[0], " ")[1]
			token, err := jwt.Parse(tokenString, auth.TokenLookup)

			if err == nil {
				// token is valid, get claims and perform authorization
				claims := token.Claims.(jwt.MapClaims)

				projects, ok := claims["projects"].(map[string]interface{})
				if !ok {
					respondWithError(http.StatusUnauthorized, "improperly formatted access token", c)
					return
				}

				user, ok := claims["user"].(map[string]interface{})
				if !ok {
					respondWithError(http.StatusUnauthorized, "improperly formatted access token", c)
					return
				}

				userType, ok := user["type"].(string)
				if !ok || userType != "app" {
					respondWithError(http.StatusUnauthorized, "invalid access token", c)
					return
				}

				userIsActive, ok := user["active"].(bool)
				if !ok || !userIsActive {
					respondWithError(http.StatusUnauthorized, "user is not active, please confirm your account", c)
					return
				}

				// inject claims into context
				c.Set("projects", projects)
				c.Set("user_id", user["id"])
				c.Set("username", user["username"])

				c.Next()
				return
			}

			respondWithError(http.StatusUnauthorized, "invalid access token", c)
			return
		}

		respondWithError(http.StatusUnauthorized, "access token required", c)
		return
	}
}

// ValidateRefreshToken validates the refresh token
func ValidateRefreshToken() gin.HandlerFunc {
	return func(c *gin.Context) {

		if values, _ := c.Request.Header["Authorization"]; len(values) > 0 {

			tokenString := strings.Split(values[0], " ")[1]
			token, err := jwt.Parse(tokenString, auth.TokenLookup)

			if err == nil {
				// token is valid, validate it's a refresh token
				claims := token.Claims.(jwt.MapClaims)

				sessionID, ok := claims["session_id"].(string)
				if !ok {
					respondWithError(http.StatusUnauthorized, "invalid refresh token", c)
					return
				}

				userID, ok := claims["user_id"].(string)
				if !ok {
					respondWithError(http.StatusUnauthorized, "invalid refresh token", c)
					return
				}

				// inject claims into context
				c.Set("session_id", sessionID)
				c.Set("user_id", userID)

				c.Next()
				return
			}

			respondWithError(http.StatusUnauthorized, "invalid refresh token", c)
			return
		}

		respondWithError(http.StatusUnauthorized, "refresh token required", c)
		return
	}
}