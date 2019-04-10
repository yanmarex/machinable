package collections

import (
	"github.com/anothrnick/machinable/dsi/interfaces"
	"github.com/anothrnick/machinable/middleware"
	"github.com/gin-gonic/gin"
)

// SetRoutes sets all of the appropriate routes to handlers for project collections
func SetRoutes(engine *gin.Engine, datastore interfaces.Datastore) error {
	// create new Collections handler with datastore, set routes -> handlers
	handler := New(datastore)

	// routes for http api access to collections
	collections := engine.Group("/collections")

	collections.Use(middleware.ProjectLoggingMiddleware(datastore))
	collections.Use(middleware.CollectionStatsMiddleware(datastore))
	collections.Use(middleware.ProjectUserAuthzMiddleware(datastore))
	collections.Use(middleware.ProjectAuthzBuildFiltersMiddleware(datastore))

	// collections.GET("/", handler.GetCollections)
	// collections.POST("/", handler.AddCollection)
	collections.POST("/:collectionName", handler.AddObjectToCollection)
	collections.GET("/:collectionName", handler.GetObjectsFromCollection)
	collections.GET("/:collectionName/:objectID", handler.GetObjectFromCollection)
	collections.PUT("/:collectionName/:objectID", handler.PutObjectInCollection)
	collections.DELETE("/:collectionName/:objectID", handler.DeleteObjectFromCollection)

	// routes for admin http api access to collections

	// admin routes with different authz policy
	mgmt := engine.Group("/mgmt")

	mgmtStats := mgmt.Group("/collectionUsage")
	mgmtStats.Use(middleware.AppUserJwtAuthzMiddleware())
	mgmtStats.Use(middleware.AppUserProjectAuthzMiddleware(datastore))
	mgmtStats.GET("/", handler.ListCollectionUsage)
	mgmtStats.GET("/stats", handler.GetStats)

	mgmtCollections := mgmt.Group("/collections")
	mgmtCollections.Use(middleware.ProjectLoggingMiddleware(datastore))
	mgmtCollections.Use(middleware.AppUserJwtAuthzMiddleware())
	mgmtCollections.Use(middleware.AppUserProjectAuthzMiddleware(datastore))
	mgmtCollections.GET("/", handler.GetCollections)
	mgmtCollections.POST("/", handler.AddCollection)
	mgmtCollections.GET("/:collectionName", handler.GetObjectsFromCollection)
	mgmtCollections.PUT("/:collectionName", handler.UpdateCollection)    // this actually uses collection ID as the parameter, gin does not allow different wildcard names
	mgmtCollections.DELETE("/:collectionName", handler.DeleteCollection) // this actually uses collection ID as the parameter, gin does not allow different wildcard names

	return nil
}
