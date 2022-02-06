package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

var (
	pathConfigProd = flag.String("config", "/etc/zfsapi/config.yaml", "Path to config.yaml")
	pathConfigDev  = "config.yaml"
	pathConfig     string
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
		// res.SetVal("cloneinfo", cloneinfo)
		res.Data = make(map[string]interface{})
		for k, v := range cloneinfo {
			res.Data[k] = v
		}
		res.Success()
	}
	res.Write(&w)
}

func apiDestroyDataset(w http.ResponseWriter, r *http.Request) {
	var (
		res jsonResponseGeneric
		err error
	)
	res.SetAction("destroy")
	if err = ZfsDestroyDataset(mux.Vars(r)["dataset"]); err != nil {
		res.Error(err.Error())
	} else {
		res.Success()
	}

	res.Write(&w)
}

func apiClone(w http.ResponseWriter, r *http.Request) {
	var (
		res jsonResponseGeneric
		err error
	)
	res.SetAction("clone")
	if err = ZfsClone(mux.Vars(r)["origin"], mux.Vars(r)["dataset"]); err != nil {
		res.Error(err.Error())
	} else {
		res.Success()
	}

	res.Write(&w)
}

func apiCloneLast(w http.ResponseWriter, r *http.Request) {
	var (
		res jsonResponseGeneric
		err error
	)
	res.SetAction("clonelast")
	if err = ZfsCloneLast(mux.Vars(r)["origin"], mux.Vars(r)["dataset"]); err != nil {
		res.Error(err.Error())
	} else {
		res.Success()
	}

	res.Write(&w)
}

func apiRollback(w http.ResponseWriter, r *http.Request) {
	var (
		res jsonResponseGeneric
		err error
	)
	res.SetAction("rollback")
	if err = ZfsRollback(mux.Vars(r)["snapshot"]); err != nil {
		res.Error(err.Error())
	} else {
		res.Success()
	}

	res.Write(&w)
}

func apiCheckZvol(w http.ResponseWriter, r *http.Request) {
	var (
		res jsonResponseGeneric
		err error
	)
	if err = ZfsCheckZvol(mux.Vars(r)["dataset"]); err != nil {
		res.Error(err.Error())
	} else {
		res.Success()
	}
	res.SetAction("checkzvol")
	res.Write(&w)
}

func run(cfg *Config) {
	router := mux.NewRouter().StrictSlash(true)
	addrString := cfg.Server.Host + ":" + cfg.Server.Port
	router.Path("/").Queries(
		"action", "listall").HandlerFunc(apiListAll)
	router.Path("/").Queries( //Create snapshot
		"action", "snapshot",
		"snapsource", "{snapsource}",
		"snapname", "{snapname}").HandlerFunc(apiCreateSnapshot)
	router.Path("/").Queries( //Find last snapshot of dataset
		"action", "lastsnapshot",
		"dataset", "{dataset}").HandlerFunc(apiGetLastSnapshot)
	router.Path("/").Queries( //Get clone information
		"action", "cloneinfo",
		"dataset", "{dataset}").HandlerFunc(apiGetCloneInfo)
	router.Path("/").Queries( //Destroy dataset
		"action", "destroy",
		"dataset", "{dataset}").HandlerFunc(apiDestroyDataset)
	router.Path("/").Queries( //Clone
		"action", "clone",
		"origin", "{origin}",
		"dataset", "{dataset}").HandlerFunc(apiClone)
	router.Path("/").Queries( //Clone last snapshot of origin
		"action", "clonelast",
		"origin", "{origin}",
		"dataset", "{dataset}").HandlerFunc(apiCloneLast)
	router.Path("/").Queries( //Rollback dataset to given snapshot
		"action", "rollback",
		"snapshot", "{snapshot}").HandlerFunc(apiRollback)
	router.Path("/").Queries( //Check if zvol symlink exists in /dev
		"action", "checkzvol",
		"dataset", "{dataset}").HandlerFunc(apiCheckZvol)
	router.Use(loggingMiddleware)
	log.Fatal(http.ListenAndServe(addrString, router))
}

func main() {
	if os.Getenv("APP_ENV") == "dev" {
		pathConfig = pathConfigDev
		log.Println("Running in development environment")
	} else {
		flag.Parse()
		pathConfig = *pathConfigProd
	}

	if cfg, err := NewConfig(pathConfig); err != nil {
		log.Fatal(err)
	} else {
		log.Printf("Using config file %s", pathConfig)
		if os.Getenv("APP_ENV") != "dev" && cfg.Server.Host != "127.0.0.1" {
			log.Printf("ZFS Api will be listening on %s. Using other than 127.0.0.1 address is NOT RECOMMENDED for production evironment!", cfg.Server.Host)
		}
		run(cfg)
	}
}
