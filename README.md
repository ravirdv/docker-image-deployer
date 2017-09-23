# docker-image-deployer
Allows you to deploy images on multiple docker hosts concurrently.


# Configuration
This service can talk to docker-engine via UDS or TCP.

# How to use

- Install dependencies via glide `glide install`
- Build service : `go build main.go types.go client_docker.go`
- Run : `./main'

## Examples
- Deploy Command using CURL:
    - `curl -XPOST  -H "Content-Type: application/json" --data @deploy.json http://localhost:8080/deploy`

- Deploy Status
    - `curl -X GET http://localhost:8080/deploystatus?name=<containername>`

- Stop
    `curl -X GET http://localhost:8080/stop?name=<containername>`

## Add/Remove/List docker hosts

- List Docker Hosts:
    `curl -X GET http://127.0.0.1:8080/listhosts`

- Add Docker Hosts:
    `curl -XGET http://127.0.0.1:8080/addhost?uri="http://127.0.0.1:2375"`

- Remove Docker Hosts:
    `curl -XGET http://127.0.0.1:8080/removehost?uri="http://127.0.0.1:2375"`


