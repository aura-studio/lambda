package httpserver

import (
	"os"
)

func init() {
	// os.Setenv("AWS_REGION", os.Getenv("CONFIG_REGION"))
	// os.Setenv("AWS_ACCESS_KEY_ID", os.Getenv("CONFIG_ACCESS_KEY_ID"))
	// os.Setenv("AWS_SECRET_ACCESS_KEY", os.Getenv("CONFIG_SECRET_ACCESS_KEY"))
	//
	os.Unsetenv("AWS_SESSION_TOKEN")
}
