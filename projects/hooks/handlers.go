package hooks

import (
	"net/http"

	"github.com/anothrnick/machinable/dsi/interfaces"
	"github.com/anothrnick/machinable/dsi/models"
	"github.com/gin-gonic/gin"
)

// New returns a pointer to a new `APIKeys` struct
func New(db interfaces.ProjectHooksDatastore) *WebHooks {
	return &WebHooks{
		store: db,
	}
}

// WebHooks wraps the datastore and any HTTP handlers for project web hooks
type WebHooks struct {
	store interfaces.ProjectHooksDatastore
}

// UpdateHook updates an existing project webhook by id and and project id
func (w *WebHooks) UpdateHook(c *gin.Context) {
	hookID := c.Param("hookID")
	projectID := c.MustGet("projectId").(string)

	hook := models.WebHook{}
	c.BindJSON(&hook)

	err := w.store.UpdateHook(projectID, hookID, &hook)
	if err != nil {
		c.JSON(err.Code(), gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}

// AddHook creates a new webhook for a project
func (w *WebHooks) AddHook(c *gin.Context) {
	projectID := c.MustGet("projectId").(string)

	hook := models.WebHook{}
	c.BindJSON(&hook)

	err := w.store.AddHook(projectID, &hook)
	if err != nil {
		c.JSON(err.Code(), gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"hook": &hook})
}

// ListHooks lists all webhooks for a project
func (w *WebHooks) ListHooks(c *gin.Context) {
	projectID := c.MustGet("projectId").(string)

	hooks, err := w.store.ListHooks(projectID)
	if err != nil {
		c.JSON(err.Code(), gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"hooks": &hooks})
}

// ListHooks lists all webhooks for a project
func (w *WebHooks) GetHook(c *gin.Context) {
	hookID := c.Param("hookID")
	projectID := c.MustGet("projectId").(string)

	hook, err := w.store.GetHook(projectID, hookID)
	if err != nil {
		c.JSON(err.Code(), gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, &hook)
}

// DeleteHook deletes a webhook by id and project
func (w *WebHooks) DeleteHook(c *gin.Context) {
	hookID := c.Param("hookID")
	projectID := c.MustGet("projectId").(string)

	err := w.store.DeleteHook(projectID, hookID)
	if err != nil {
		c.JSON(err.Code(), gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, gin.H{})
}
