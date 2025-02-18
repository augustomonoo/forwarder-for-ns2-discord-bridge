package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
)

type Endpoint struct {
	Endpoint string
	IDs      []string
	Servers  []string
}

type Config struct {
	BindAddress    string
	BindPort       int
	FieldToReplace string
	Endpoints      []Endpoint
}

var (
	CONFIGURATION Config
	config_path   = "config.json"
)

func loadConfig(b []byte, c *Config) {
	if err := json.Unmarshal(b, &c); err != nil {
		log.Fatal(err)
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
		log.Printf("ERROR sending data to %s\n", url)
		log.Printf("%s", err)
	}
	return res
}

func receiveAndForward(r *http.Request, endpoint Endpoint) {
	// Data is received through a POST form
	r.ParseForm()
	if r.Form == nil {
		// TODO maybe return an error if form is empty?
		return
	}
	// Maybe check to only process POST requests
	for _, id := range endpoint.IDs {
		for _, url := range endpoint.Servers {
			formData := cloneFormData(r.Form)
			formData[CONFIGURATION.FieldToReplace] = []string{id}
			log.Printf("[%s] %s => [%s]", endpoint.Endpoint, url, id)
			response := sendData(url, formData)
			log.Printf("[%s] %s => [%s]: %d", endpoint.Endpoint, url, id, response.StatusCode)
		}
	}
}

func handleEndpoint(w http.ResponseWriter, r *http.Request) {
	// Get the /pattern from the URL
	// Call receiveAndForward with the server list that
	// matches the pattern found

	// remove the first slash in the pattern. Eg
	// /endpoint => endpoint
	requested_endpoint := r.URL.Path[1:]

	found := false
	for _, endpoint := range CONFIGURATION.Endpoints {
		if requested_endpoint == endpoint.Endpoint {
			found = true
			log.Printf("[%s] new request %s", endpoint.Endpoint, r.RemoteAddr)
			receiveAndForward(r, endpoint)
		}
	}
	if !found {
		log.Printf("Received request for non existing endpoint '%s'", requested_endpoint)
	}
}

func main() {
	file_content, err := os.ReadFile("config.json")
	if err != nil {
		log.Fatal(err)
	}
	loadConfig(file_content, &CONFIGURATION)

	// Just print the loaded servers
	for _, endpoint := range CONFIGURATION.Endpoints {
		log.Printf("Endpoint: '%s'\n", endpoint.Endpoint)
		log.Println("  Server IDs:")
		for i, id := range endpoint.IDs {
			log.Printf("   [%d]: %s\n", i, id)
		}
		log.Println("  Discord Bridge Servers:")
		for i, url := range endpoint.Servers {
			log.Printf("   [%d]: %s\n", i, url)
		}
	}
	addr := fmt.Sprintf("%s:%d", CONFIGURATION.BindAddress, CONFIGURATION.BindPort)
	log.Printf("Listening on %s", addr)
	http.HandleFunc("/", handleEndpoint)
	err = http.ListenAndServe(addr, nil)
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("Server closed\n")
	} else if err != nil {
		fmt.Printf("error starting server : %s\n", err)
		os.Exit(1)
	}
}
