package envs

import (
	"log"
	"os"

	"github.com/user0608/ifdevmode"
)

func FindEnv(key, defaultValue string) string {
	value, defined := os.LookupEnv(key)
	if defined {

		return value
	}
	if ifdevmode.Yes() {
		log.Printf("default value (%s) was load (%s)\n", key, defaultValue)
	}
	return defaultValue
}
