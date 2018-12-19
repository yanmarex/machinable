package interfaces

import "bitbucket.org/nsjostrom/machinable/dsi/models"

// ProjectAPIKeysDatastore exposes functions to manage project api keys
type ProjectAPIKeysDatastore interface {
	GetAPIKeyByKey(project, hash string) (*models.ProjectAPIKey, error)
	CreateAPIKey(project, hash, description string, read, write bool) (*models.ProjectAPIKey, error)
	ListAPIKeys(project string) ([]*models.ProjectAPIKey, error)
	DeleteAPIKey(project, keyID string) error
}