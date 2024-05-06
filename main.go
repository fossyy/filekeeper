package main

import (
	"github.com/fossyy/filekeeper/app"
)

func main() {
	//TODO HANDLE OAUTH ERROR : INFO: 2024/05/06 11:07:24 127.0.0.1 GET /auth/google/callback?error=access_denied&state=******** 303
	app.Start()
}
