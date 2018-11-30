package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"time"

	"github.com/gorilla/mux"
	"github.com/marksauter/markus-ninja-images/pkg/myconf"
	"github.com/marksauter/markus-ninja-images/pkg/mylog"
	"github.com/marksauter/markus-ninja-images/pkg/server/middleware"
	"github.com/marksauter/markus-ninja-images/pkg/server/route"
	"github.com/marksauter/markus-ninja-images/pkg/service"
	"github.com/marksauter/markus-ninja-images/pkg/util"
)

func main() {
	branch := util.GetRequiredEnv("BRANCH")
	confFilename := fmt.Sprintf("config.%s", branch)
	conf := myconf.Load(confFilename)

	svcs, err := service.NewServices(conf)
	if err != nil {
		mylog.Log.WithField("error", err).Fatal(util.Trace("unable to start services"))
	}

	r := mux.NewRouter()

	indexHandler := route.IndexHandler{
		Conf:       conf,
		StorageSvc: svcs.Storage,
	}
	index := middleware.CommonMiddleware.Append(
		indexHandler.Cors().Handler,
	).Then(indexHandler)

	if branch == "development.local" || branch == "test" {
		r.PathPrefix("/debug/").Handler(http.DefaultServeMux)
	}
	r.Handle("/{user_id}/{key}", index)

	router := http.TimeoutHandler(r, 5*time.Second, "Timeout!")

	port := util.GetOptionalEnv("PORT", "5050")
	address := ":" + port
	mylog.Log.Infof("Listening on port %s", port)
	mylog.Log.Fatal(http.ListenAndServe(address, router))
}
