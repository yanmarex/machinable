package logs

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/machinable/machinable/dsi"
	"github.com/machinable/machinable/dsi/interfaces"
	"github.com/machinable/machinable/dsi/models"
	"github.com/machinable/machinable/query"
)

// New returns a pointer to a new instance of the Logs handler
func New(db interfaces.ProjectLogsDatastore) *Logs {
	return &Logs{
		store: db,
	}
}

// Logs wraps handler access to project logs
type Logs struct {
	store interfaces.ProjectLogsDatastore
}

// ListProjectLogs returns the list of activity logs for a project
func (l *Logs) ListProjectLogs(c *gin.Context) {
	projectID := c.MustGet("projectId").(string)

	// Get pagination parameters
	values := c.Request.URL.Query()

	iLimit, err := query.GetLimit(&values)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iOffset, err := query.GetOffset(&values)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// filter anything within x hours
	old := time.Now().Add(-time.Hour * time.Duration(24))
	filter := &models.Filters{
		"created": models.Value{
			models.GTE: old,
		},
	}
	sort := make(map[string]int)
	for k, v := range values {
		if k == dsi.LimitKey || k == dsi.OffsetKey {
			continue
		}

		// check for the order of the sort
		if k == dsi.SortKey {
			sortField := v[0]
			firstChar := string(sortField[0])
			order := 1
			if firstChar == "-" {
				order = -1
				sortField = sortField[1:]
			}
			sort[sortField] = order
			continue
		}

		// validate field exists
		if models.IsValidLogField(k) {
			value, err := models.FieldAsTypedInterface(k, v[0])
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid value for field '%s'", k)})
				return
			}
			filter.AddFilter(k, models.Value{
				models.EQ: value,
			})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("'%s' is not a valid field", k)})
			return
		}
	}

	// get count for pagination
	logCount, err := l.store.CountProjectLogs(projectID, filter)

	pageMax := (logCount % iLimit) + logCount
	if (iLimit+iOffset) > pageMax && iOffset >= logCount && logCount != 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid page"})
		return
	}

	logs, err := l.store.ListProjectLogs(projectID, iLimit, iOffset, filter, sort)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	links := query.NewLinks(c.Request, iLimit, iOffset, logCount)

	c.PureJSON(http.StatusOK, gin.H{"items": logs, "links": links, "count": logCount})
}
