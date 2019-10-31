package jsontree

import (
	"github.com/anothrnick/machinable/dsi/interfaces"
	"github.com/anothrnick/machinable/middleware"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
)

// Handler is an interface to the JSON key/val HTTP handler functions.
type Handler interface {
	ListRootKeys(c *gin.Context)
	CreateRootKey(c *gin.Context)
	ReadRootKey(c *gin.Context)
	DeleteRootKey(c *gin.Context)
	ReadJSONKey(c *gin.Context)
	CreateJSONKey(c *gin.Context)
	UpdateJSONKey(c *gin.Context)
	DeleteJSONKey(c *gin.Context)
}

// SetRoutes sets all of the appropriate routes to handlers for the application
func SetRoutes(engine *gin.Engine, datastore interfaces.Datastore, cache redis.UniversalClient) error {
	handler := NewHandlers(datastore)

	return setRoutes(engine, handler, datastore, cache)
}

// abstraction for dependency injection
func setRoutes(engine *gin.Engine, h Handler, datastore interfaces.Datastore, cache redis.UniversalClient) error {
	jsonKeys := engine.Group("/json")

	// set middleware on json key resources
	jsonKeys.Use(middleware.JSONStatsMiddleware(datastore))
	jsonKeys.Use(middleware.ProjectUserAuthzMiddleware(datastore))
	jsonKeys.Use(middleware.RequestRateLimit(datastore, cache))

	// initialize routes
	jsonKeys.GET("/:rootKey/*keys", h.ReadJSONKey)
	jsonKeys.POST("/:rootKey/*keys", h.CreateJSONKey)
	jsonKeys.PUT("/:rootKey/*keys", h.UpdateJSONKey)
	jsonKeys.DELETE("/:rootKey/*keys", h.DeleteJSONKey)

	// App mgmt routes with different authz policy
	mgmt := engine.Group("/mgmt")
	mgmtAPI := mgmt.Group("/json")
	mgmtAPI.Use(middleware.AppUserJwtAuthzMiddleware())
	mgmtAPI.Use(middleware.AppUserProjectAuthzMiddleware(datastore))
	mgmtAPI.GET("/", h.ListRootKeys)             // returns all root keys
	mgmtAPI.GET("/:rootKey", h.ReadRootKey)      // returns entire root tree
	mgmtAPI.POST("/:rootKey", h.CreateRootKey)   // create a new tree at `rootKey`
	mgmtAPI.DELETE("/:rootKey", h.DeleteRootKey) // root tree must be empty to delete

	return nil
}
