package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-openapi/spec"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/validate"
)

// ResourceObject is a custom type which wraps a map[string]interface
type ResourceObject map[string]interface{}

// Validate validates that the object matches the schema
func (obj *ResourceObject) Validate(definition *ResourceDefinition) error {
	schema := new(spec.Schema)
	properties := fmt.Sprintf(`{"properties": %s }`, definition.Properties)

	err := json.Unmarshal([]byte(properties), schema)
	if err != nil {
		return err
	}

	data := map[string]interface{}{}
	for key, val := range *obj {
		data[key] = val
	}

	// validate data against schema
	res := validate.NewSchemaValidator(schema, nil, "", strfmt.Default).Validate(data)
	if res.HasErrors() {
		errs := []string{}
		for _, e := range res.Errors {
			errs = append(errs, e.Error())
		}
		return errors.New(strings.Join(errs, ","))
	}
	return nil
}

// ResourceDefinition defines an API resource
type ResourceDefinition struct {
	ID            string    `json:"id"`        // ID is the unique identifier for this resource definition
	Title         string    `json:"title"`     // Title of this resource
	PathName      string    `json:"path_name"` // PathName is the name that will appear in the URL path
	ParallelRead  bool      `json:"parallel_read"`
	ParallelWrite bool      `json:"parallel_write"`
	Created       time.Time `json:"created"`    // Created is the timestamp the resource was created
	Properties    string    `json:"properties"` // Properties is the string representation of the JSON schema properties
}

// MarshalJSON custom marshaller to marshall properties to json
func (def *ResourceDefinition) MarshalJSON() ([]byte, error) {
	properties := map[string]interface{}{}
	err := json.Unmarshal([]byte(def.Properties), &properties)
	if err != nil {
		panic(err)
	}

	return json.Marshal(&struct {
		ID            string                 `json:"id"`        // ID is the unique identifier for this resource definition
		Title         string                 `json:"title"`     // Title of this resource
		PathName      string                 `json:"path_name"` // PathName is the name that will appear in the URL path
		ParallelRead  bool                   `json:"parallel_read"`
		ParallelWrite bool                   `json:"parallel_write"`
		Created       time.Time              `json:"created"`    // Created is the timestamp the resource was created
		Properties    map[string]interface{} `json:"properties"` // Properties is the string representation of the JSON schema properties
	}{
		ID:            def.ID,
		Title:         def.Title,
		PathName:      def.PathName,
		ParallelRead:  def.ParallelRead,
		ParallelWrite: def.ParallelWrite,
		Created:       def.Created,
		Properties:    properties,
	})
}

// UnmarshalJSON is a custom unmarshaller
func (def *ResourceDefinition) UnmarshalJSON(b []byte) error {
	payload := struct {
		Title         string          `json:"title"`      // Title of this resource
		PathName      string          `json:"path_name"`  // PathName is the name that will appear in the URL path
		Properties    json.RawMessage `json:"properties"` // Properties is the string representation of the JSON schema properties
		ParallelRead  bool            `json:"parallel_read"`
		ParallelWrite bool            `json:"parallel_write"`
	}{}

	err := json.Unmarshal(b, &payload)

	if err != nil {
		panic(err)
	}

	def.Title = payload.Title
	def.PathName = payload.PathName
	def.Properties = string(payload.Properties)
	def.ParallelRead = payload.ParallelRead
	def.ParallelWrite = payload.ParallelWrite

	return nil
}

// Validate validates the fields of a resource definition.
func (def *ResourceDefinition) Validate() error {
	if def.Title == "" {
		return errors.New("resource title cannot be empty")
	} else if def.PathName == "" {
		return errors.New("resource path_name cannot be empty")
	} else if def.Properties == "" {
		return errors.New("resource properties cannot be empty")
	}

	schema := new(spec.Schema)

	err := json.Unmarshal([]byte(def.Properties), schema)

	return err
}
