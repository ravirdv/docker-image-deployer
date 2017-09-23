package main

import (
	"context"
	"encoding/json"
	"log"
	"net/url"
)
import (
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"fmt"
)

// this holds docker clients
var clientMap = make(map[string]*client.Client)

// bootstrap our app with predefined docker hosts
func Initialize() {
	for _, host := range dockerHosts {
		AddHost(host)
	}
}

// triggers stop command on docker-engine
func StopContainer(name string) string {
	// stores host wise operation result.
	statusMap := make(map[string]string)
	// let's make sure name is not blank
	if name != "" {
		// we'll fire stop container command to each host
		for host, c := range clientMap {
			if err := c.ContainerStop(context.Background(), name, nil); err != nil {
				log.Print("WARN: failed to stop container : ", name, " on host : ", host)
				statusMap[host] = fmt.Sprintln(err)
			} else {
				statusMap[host] = "operation successfully completed"
			}
		}
	}
	// convert our map to json
	jsonString, err := json.Marshal(statusMap)
	if err != nil {
		return "{ \"error\" : \"Internal server error\" }"
	}
	return string(jsonString)
}

// triggers ListContainer command on docker-engine
func GetContainerStatus(name string) string {
	// stores host wise operation result.
	statusMap := make(map[string]interface{})
	// let's make sure name is not blank
	if name != "" {
		// we'll trigger ListContainer command on each docker host
		for host, c := range clientMap {
			log.Print("Trying to get container status from host ", host, " with name ", name)
			// let's prepare our name filter.
			filters := filters.NewArgs()
			filters.Add("name", name)
			// let's get container list from dockerengine
			containers, err := c.ContainerList(context.Background(), types.ContainerListOptions{
				All:     true, // we want to get status for all containers
				Filters: filters, // we'll filter it by name
			})
			// let's check for error
			if err != nil {
				log.Print("WARN: failed to fetch container list from host : ", host)
				log.Print(err)
				statusMap[host] = errorMessageStruct{ ErrorMessage: "failed to fetch container list" }
			} else {
				// looks good, lets store respose
				statusMap[host] = containers
			}
		}
	}
	// convert our status map to json
	jsonString, err := json.Marshal(statusMap)
	if err != nil {
		return "{ \"error\" : \"Internal server error\" }"
	}
	return string(jsonString)
}

// this will download image, create and run container based on given params.
func DeployAndRunContainer(params *deployStruct) string {
	statusMap := make(map[string]string)
	// prepared exposed ports

	for host, _ := range clientMap {
		if enableConcurrency {
			go runContainer(params.Image, params.Name, params.Cmd, params.Env, params.Volumes, host)
			statusMap["message"] = "Operation request successful"
		}else{
		    statusMap[host] = runContainer(params.Image, params.Name, params.Cmd, params.Env, params.Volumes, host)
		}
	}
	// prepare our JSON string
	jsonString, err := json.Marshal(statusMap)
	// couldn't prepare JSON string
	if err != nil {
		return "{ \"error\" : \"Internal server error\" }"
	}
	return string(jsonString)
}

func AddHost(uri string) bool {
	// validate uri
	_, err := url.ParseRequestURI(uri)
	if err != nil{
		log.Print("WARN: AddHost: invalid uri : ", uri)
		return false
	}
	// check if host already exists
	if _, ok := clientMap[uri]; ok {
		log.Print("WARN: AddHost: already exists with uri : ", uri)
		return true
	}

	log.Print("registering docker host : ", uri)
	var c, error = client.NewClient(uri, "", nil, nil)
	if err != nil {
		log.Print("WARN: failed to initialize docker client for host : ", uri)
		log.Print(error)
	}
	clientMap[uri] = c
	return true
}

func RemoveHost(uri string) bool {
	// validate uri
	_, err := url.ParseRequestURI(uri)
	if err != nil {
		log.Print("WARN: RemoveHost: invalid uri : ", uri)
		return false
	}

	if _, ok := clientMap[uri]; ok {
		clientMap[uri].Close()
		delete(clientMap, uri)
		log.Print("Removed Host: ", uri)
	}
	return true
}

func runContainer(image, name string, commands, env_vars []string, volumes []string, host string) string {
		c := clientMap[host]
		// download image
		if image != "" {
			log.Print("Trying to pull image ", image, " on host ", host)
			_, err := c.ImagePull(context.Background(), image, types.ImagePullOptions{})
			if err != nil {
				log.Print(err)
				return "Failed to download image."
			}
			log.Print("Trying to create container ", name, " on host ", host)
			// now that we got image, let's create container
			resp, err := c.ContainerCreate(context.Background(), &container.Config{
				Image: image,
				Cmd:   commands,
				Env: env_vars,
				// TODO  fix deps issue https://github.com/moby/moby/issues/29362
				//ExposedPorts: nat.PortSet{
				//},
			}, &container.HostConfig{
				Binds: volumes,
			}, nil, name)
			if err != nil {
				log.Print(err)
				return "Failed to create container."
			}
			log.Print("Trying to start container ", resp.ID, " on host ", host)

			// we've created container, now let's start it.
			if err := c.ContainerStart(context.Background(), resp.ID, types.ContainerStartOptions{}); err != nil {
				log.Print(err)
				return "Failed to start container."
			}
			// operation successfully completed
			return "Container started successfully"
		} else {
			// cant' have blank image name
			return "Invalid image name"
		}
}
