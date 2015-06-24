package main

import (
	"github.com/wayt/happyngine"
	"github.com/wayt/happyngine/env"
	"github.com/wayt/happyngine/log"
	"math/rand"
	"net/http"
	"runtime"
	"time"
)

func main() {

	app := happyngine.NewAPI()

	// Setup seed
	rand.Seed(time.Now().UnixNano())

	// Setup Origin
	if origin := env.Get("ALLOW_ORIGIN"); len(origin) > 0 {
		app.Headers["Access-Control-Allow-Origin"] = origin
	}

	// Register actions
	app.AddRoute("GET", "/:bucket/:file", newGetFileAction)

	// Setup custuom 404 handler
	app.Error404Handler = func(ctx *happyngine.Context, err interface{}) {

		ctx.Send(http.StatusNotFound, `not found 404`)
	}

	// Setup custuom panic handler
	app.PanicHandler = func(ctx *happyngine.Context, err interface{}) {

		ctx.Send(500, `{"error":"internal_error"}`)

		trace := make([]byte, 1024)
		runtime.Stack(trace, true)

		ctx.Criticalln(err, string(trace))
	}

	log.Debugln("Running...")
	if err := app.Run(":8080"); err != nil {
		log.Criticalln("app.Run:", err)
	}
}
