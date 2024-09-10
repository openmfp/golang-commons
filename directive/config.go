package directive

type configuration struct {
	DirectivesAuthorizationEnabled bool `envconfig:"default=false"`
}

var directiveConfiguration = configuration{}
