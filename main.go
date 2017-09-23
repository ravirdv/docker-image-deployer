package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"encoding/json"
)

// list of docker hosts, can be managed by API.
var dockerHosts = []string{"http://127.0.0.1:2375"}

// do we want to enable concurrency while creating deployments
var enableConcurrency = true

func main() {
	// initialize dockerclients
	Initialize()
	// let's setup our routers
	router := mux.NewRouter()
	router.HandleFunc("/deploy", Deploy).Methods("POST").HeadersRegexp("Content-Type", "application/json")
	router.HandleFunc("/deploystatus", Status).Methods("GET")
	router.HandleFunc("/stop", Stop).Methods("GET")

	// to dynamically add/remove docker hosts
	router.HandleFunc("/addhost", AddDockerHost)
	router.HandleFunc("/removehost", RemoveDockerHost)
	router.HandleFunc("/listhosts", ListDockerHosts)
	// let's start our server
	log.Print("Starting server on port 8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}

// entry point for our deploy endpoint
func Deploy(w http.ResponseWriter, r *http.Request) {
	// this will hold our decoded request json
    var requestParams deployStruct
	// let's decode request body
    decoder := json.NewDecoder(r.Body)
    err := decoder.Decode(&requestParams)
	// is it a bad request?
    if err != nil {
        log.Print("ERROR: failed to decode request JSON")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
    }

	// let's forward request to docker hosts
	fmt.Fprintln(w, DeployAndRunContainer(&requestParams))
}

// gets container status based on givne name from all hosts
func Status(w http.ResponseWriter, r *http.Request) {
	  // name is expected
	  name := r.URL.Query().Get("name")
	  if name != "" {
		 fmt.Fprintln(w, GetContainerStatus(name))
	  } else {
		  // error
		http.Error(w, "Bad request : name cannot be blank", http.StatusBadRequest)
	  }
}

// stops container with given name on all hosts
func Stop(w http.ResponseWriter, r *http.Request) {
	  // name is expected
	  name := r.URL.Query().Get("name")
	  if name != "" {
		  // lets trigger stop command on all docker hosts
		 fmt.Fprintln(w, StopContainer(name))
	  } else {
		  // name not given
		http.Error(w, "Bad request : name cannot be blank", http.StatusBadRequest)
	  }
}

// register a new docker host
func AddDockerHost(w http.ResponseWriter, r *http.Request) {
	  if AddHost(r.URL.Query().Get("uri")) {
		http.Error(w, http.StatusText(http.StatusOK), http.StatusOK)
	  } else {
		  // uri not given
		http.Error(w, "Bad request : uri cannot be blank", http.StatusBadRequest)
	  }
}

// remove docker host from our list
func RemoveDockerHost(w http.ResponseWriter, r *http.Request) {
	  // host uri is expected
	  if RemoveHost(r.URL.Query().Get("uri")) {
		http.Error(w, http.StatusText(http.StatusOK), http.StatusOK)
	  } else {
		  // uri not given
		http.Error(w, "Bad request : uri cannot be blank", http.StatusBadRequest)
	  }
}

// lists all registered docker hosts
func ListDockerHosts(w http.ResponseWriter, r *http.Request) {
	// let's get all host uri from map.
	var hosts []string
	for host, _ := range clientMap {
		hosts = append(hosts, host)
	}
	// map with list of hosts
	response := make(map[string][]string)
	response["hosts"] = hosts
	// convert map to json
	jsonString, err := json.Marshal(response)
	if err != nil {
		fmt.Fprintln(w,"{ \"error\" : \"Internal server error\" }")
	}
	fmt.Fprintln(w,string(jsonString))
}
