package main

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
)

var serverList map[string][]ForwardServer

// TODO check if the field that needs replacing is 'server'
var formFieldToReplace = "server"

type ForwardServer struct {
	url       string
	server_id string
}

func getForwardServers() map[string][]ForwardServer {
	return map[string][]ForwardServer{
		"server1": {{"http://localhost:8000", "banana"}, {"http://localhost:8000", "phone"}},
		"server2": {{"http://localhost:8000", "discord_server_1"}, {"http://localhost:8000/discbridge", "discord_server_2"}},
	}
}

func cloneFormData(data url.Values) url.Values {
	new_data := url.Values{}
	for key, value := range data {
		new_data[key] = value
	}
	return new_data
}

func sendData(url string, data url.Values) *http.Response {
	res, err := http.PostForm(url, data)
	if err != nil {
		fmt.Printf("ERROR sending data to %s\n", url)
		fmt.Printf("%s", err)
	}
	fmt.Printf("%d\n", res.StatusCode)
	return res
}

func receiveAndForward(r *http.Request, forwardServers []ForwardServer) {
	// Data is received through a POST form
	r.ParseForm()
	if r.Form == nil {
		// TODO maybe return an error if form is empty?
		return
	}
	// Maybe check to only process POST requests
	for _, forwardServer := range forwardServers {
		formData := cloneFormData(r.Form)
		formData[formFieldToReplace] = []string{forwardServer.server_id}
		sendData(forwardServer.url, formData)
	}
}

func handleEndpoint(w http.ResponseWriter, r *http.Request) {
	// Get the /pattern from the URL
	// Call receiveAndForward with the server list that
	// matches the pattern found

	// remove the first slash in the pattern. Eg
	// /endpoint => endpoint
	requested_endpoint := r.URL.Path[1:]

	forwardServerList, ok := serverList[requested_endpoint]
	if ok {
		receiveAndForward(r, forwardServerList)
	} else {
		fmt.Printf("Received request for undefined endpoint: '%s'\n", requested_endpoint)
	}
}

func main() {
	serverList = getForwardServers()
	// Just print the loaded servers
	for endpoint, server := range serverList {
		fmt.Printf("Servers for endpoint '%s':\n", endpoint)
		for i, forward_server := range server {
			fmt.Printf("  [%d] url: '%s', server_id: '%s'\n", i, forward_server.url, forward_server.server_id)
		}
	}

	http.HandleFunc("/", handleEndpoint)
	err := http.ListenAndServe(":3333", nil)
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("Server closed\n")
	} else if err != nil {
		fmt.Printf("error starting server : %s\n", err)
		os.Exit(1)
	}
}
