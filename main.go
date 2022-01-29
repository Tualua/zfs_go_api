package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func loggingMiddleware(next http.Handler) http.Handler {
	return handlers.CombinedLoggingHandler(os.Stdout, next)
}

func apiListAll(w http.ResponseWriter, r *http.Request) {
	var (
		res  jsonResponseListAll
		err  error
		data []ZfsEntity
	)
	res.SetAction("listall")

	if data, err = ZfsListAll(); err != nil {
		res.Error(err.Error())
	} else {
		res.Success()
		res.ZfsEntities = data
	}

	res.Write(&w)
}

func apiCreateSnapshot(w http.ResponseWriter, r *http.Request) {
	var (
		res jsonResponseGeneric
	)
	res.SetAction("snapshot")
	if err := ZfsCreateSnapshot(mux.Vars(r)["snapsource"], mux.Vars(r)["snapname"]); err != nil {
		res.Error(err.Error())
	} else {
		res.Success()
	}
	res.Write(&w)
}

func apiGetLastSnapshot(w http.ResponseWriter, r *http.Request) {
	var (
		lastSnapshot string
		res          jsonResponseGeneric
		err          error
	)
	res.SetAction("lastsnapshot")
	if lastSnapshot, err = ZfsGetLastSnapshot(mux.Vars(r)["dataset"]); err != nil {
		res.Error(err.Error())
	} else {
		res.SetVal("lastsnapshot", lastSnapshot)
		res.Success()
	}

	res.Write(&w)
}

func apiGetCloneInfo(w http.ResponseWriter, r *http.Request) {
	var (
		res       jsonResponseGeneric
		cloneinfo map[string]string
		err       error
	)
	res.SetAction("cloneinfo")

	if cloneinfo, err = ZfsGetCloneInfo(mux.Vars(r)["dataset"]); err != nil {
		res.Error(err.Error())
	} else {
		res.SetVal("cloneinfo", cloneinfo)
		res.Success()
	}
	res.Write(&w)
}

func run(cfg *Config) {
	router := mux.NewRouter().StrictSlash(true)
	addrString := cfg.Server.Host + ":" + cfg.Server.Port
	//router.HandleFunc("/listall", apiListAll)
	router.Path("/").Queries("action", "listall").HandlerFunc(apiListAll)
	router.Path("/").Queries("action", "snapshot",
		"snapsource", "{snapsource}",
		"snapname", "{snapname}").HandlerFunc(apiCreateSnapshot)
	router.Path("/").Queries("action", "lastsnapshot",
		"dataset", "{dataset}").HandlerFunc(apiGetLastSnapshot)
	router.Path("/").Queries("action", "cloneinfo",
		"dataset", "{dataset}").HandlerFunc(apiGetCloneInfo)
	router.Use(loggingMiddleware)
	log.Fatal(http.ListenAndServe(addrString, router))
}

func main() {
	cfg, err := NewConfig("config.yaml")
	if err != nil {
		log.Fatal(err)
	}
	run(cfg)

}
