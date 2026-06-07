package env

import "os"

func IsDevMode() bool {
	return os.Getenv("ED_DEV_MODE") != ""
}
