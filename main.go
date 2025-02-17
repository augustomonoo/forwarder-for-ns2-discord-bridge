package main

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
)

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

func getRoot(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Got request %s\n", r.Method)
	r.ParseForm()
	if r.Form == nil {
		return
	}
	clonedFormData := cloneFormData(r.Form)
	res := sendData("http://localhost:8080/index.php", clonedFormData)
	fmt.Printf("%s", res.Body)
}

func main() {
	http.HandleFunc("/", getRoot)
	err := http.ListenAndServe(":3333", nil)
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("Server closed\n")
	} else if err != nil {
		fmt.Printf("error starting server : %s\n", err)
		os.Exit(1)
	}
}
