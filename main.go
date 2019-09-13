package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

// #region Helpers

func Respond(writer http.ResponseWriter, data map[string]interface{}) {
	writer.Header().Add("Content-Type", "application/json")
	json.NewEncoder(writer).Encode(data)
}

func findRoute(items []Route, query string) (Route, bool) {
	var exist Route
	for _, item := range items {
		if query == item.Path {
			return item, true
		}
	}
	return exist, false
}

// endregion

// #region Struct

type Route struct {
	Method string `json:"method"`
	Path   string `json:"path"`
	Data   string `json:"data"`
}

type Config struct {
	Version string  `json:"version"`
	Port    string  `json:"port"`
	Routes  []Route `json:"routes"`
}

// endregion

// #region Handlers

func aboutHandler(w http.ResponseWriter, r *http.Request) {
	resp := map[string]interface{}{
		"name":      "jessica",
		"version":   "0.1",
		"codename":  "Llamas in Pajamas",
		"copyright": "Copyright (c) 2019 iDesoft Systems. All Rights Reserved.",
	}
	Respond(w, resp)
}

func handlerByStaticFile(fs http.Handler, mux http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		jsonFile, errFile := os.Open("jessica.json")
		if errFile != nil {
			log.Fatalf("ConfigurationError: %v", errFile)
		}
		defer jsonFile.Close()

		byteValue, _ := ioutil.ReadAll(jsonFile)
		var config Config
		if err := json.Unmarshal(byteValue, &config); err != nil {
			log.Printf("ConfigurationError: %v", err)
			http.Error(w, "Page Not Found", http.StatusNotFound)
			return
		}

		routes := config.Routes
		route, found := findRoute(routes, req.URL.Path)
		if found {
			log.Printf("Started %v \"%v\"", route.Method, req.URL.Path)

			fsHandler := http.StripPrefix("", fs)
			log.Printf("Processing by %v", route.Data)

			// rewrite url to static file
			req.URL.Path = route.Data

			fsHandler.ServeHTTP(w, req)
			log.Printf("Completed")
		} else if mux != nil {
			log.Printf("RoutingError: No route matches [%v] \"%v\"", req.Method, req.URL.Path)
			mux.ServeHTTP(w, req)
		} else {
			http.Error(w, "Page Not Found", http.StatusNotFound)
		}
	}

	return http.HandlerFunc(fn)
}

// endregion

func main() {

	jsonFile, errFile := os.Open("jessica.json")
	if errFile != nil {
		log.Fatalf("ConfigurationError: %v", errFile)
	}

	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)
	var config Config
	if err := json.Unmarshal(byteValue, &config); err != nil {
		log.Fatalf("ConfigurationError: %v", err)
	}

	log.Println("=> Jessica 0.1 application starting")
	log.Printf("* Mock version %v\n", config.Version)

	mux := http.NewServeMux()
	mux.HandleFunc("/jessica", aboutHandler)

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/", handlerByStaticFile(fs, mux))

	port := config.Port
	if port == "" {
		port = "5000"
	}

	log.Printf("* Listening on tcp://0.0.0.0:%v\n", port)
	log.Printf("Use Ctrl-C to stop\n")

	log.Fatal(http.ListenAndServe(":"+port, nil))
}
