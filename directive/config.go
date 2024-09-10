package directive

import (
	"log"

	"github.com/vrischmann/envconfig"
)

type configuration struct {
	DirectivesAuthorizationEnabled bool `envconfig:"default=false"`
}

var directiveConfiguration = configuration{}

func init() {
	err := envconfig.Init(&directiveConfiguration)
	if err != nil {
		log.Fatalln(err)
	}
}
