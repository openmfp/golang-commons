package contentconfiguration

import (
	"encoding/json"
	"github.com/xeipuuv/gojsonschema"
	"gopkg.in/yaml.v3"
)

type contentConfiguration struct{}

func NewContentConfiguration() ContentConfigurationInterface {
	return &contentConfiguration{}
}

func (cC *contentConfiguration) ValidateJSON(input []byte) (bool, []string) {
	var config ContentConfiguration
	if err := json.Unmarshal(input, &config); err != nil {
		return false, []string{err.Error()}
	}
	return validateConfig(config)
}

func (cC *contentConfiguration) ValidateYAML(input []byte) (bool, []string) {
	var config ContentConfiguration
	if err := yaml.Unmarshal(input, &config); err != nil {
		return false, []string{err.Error()}
	}
	return validateConfig(config)
}

func validateConfig(config ContentConfiguration) (bool, []string) {
	var errors []string

	if config.Name == "" {
		errors = append(errors, "Name is a mandatory parameter and should be specified")
	}

	for _, fragment := range config.LuigiConfigFragment {
		for _, node := range fragment.Data.Nodes {
			if node.EntityType == "" {
				errors = append(errors, "EntityType is a mandatory parameter and should be specified")
			}
			if node.PathSegment == "" {
				errors = append(errors, "PathSegment is a mandatory parameter and should be specified")
			}
			if node.Label == "" {
				errors = append(errors, "Label is a mandatory parameter and should be specified")
			}
			if node.Icon == "" {
				errors = append(errors, "Icon is a mandatory parameter and should be specified")
			}
		}
	}

	return len(errors) == 0, errors
}

func (cC *contentConfiguration) ValidateSchema(input []byte, schema string) (bool, []string) {
	schemaLoader := gojsonschema.NewStringLoader(schema)
	documentLoader := gojsonschema.NewBytesLoader(input)

	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return false, []string{err.Error()}
	}

	if !result.Valid() {
		var errors []string
		for _, desc := range result.Errors() {
			errors = append(errors, desc.String())
		}
		return false, errors
	}

	return true, nil
}
