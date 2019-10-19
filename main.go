package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"reflect"
	"time"
)

const (
	AppName      = "Jessica Mock Tool"
	AppVersion   = "0.4"
	AppCodename  = "Llamas in Pajamas"
	AppCopyright = "Copyright (c) 2019 iDesoft Systems. All Rights Reserved."
)

// #region Struct

type Request struct {
	Url    string `json:"url"`
	Method string `json:"method"`
	Body   string `json:"body"`
}

type Response struct {
	Status      int    `json:"status"`
	ContentType string `json:"content-type"`
	Content     string `json:"content"`
}

type Stub struct {
	Request  Request  `json:"request"`
	Response Response `json:"response"`
}

type Config struct {
	Version        string `json:"version"`
	Port           string `json:"port"`
	AllowedHeaders string `json:"allowed_headers"`
	AllowedOrigins string `json:"allowed_origins"`
	AllowedMethods string `json:"allowed_methods"`
	Stubs          []Stub `json:"stubs"`
}

// endregion

// #region Helpers

func Respond(writer http.ResponseWriter, data map[string]interface{}) {
	writer.Header().Add("Content-Type", "application/json")
	json.NewEncoder(writer).Encode(data)
}

func Message(message string) map[string]interface{} {
	return map[string]interface{}{"message": message}
}

func findStub(items []Stub, query *http.Request) (Stub, bool) {
	var exist Stub
	var requestRawBody map[string]interface{}
	requestDecoder := json.NewDecoder(query.Body)
	if err := requestDecoder.Decode(&requestRawBody); err != nil && err != io.EOF {
		log.Printf("RequestDecoderError: %v\n", err)
	}
	log.Printf("Parameters: %v", requestRawBody)

	for _, item := range items {
		if query.URL.Path == item.Request.Url && query.Method == item.Request.Method {

			if item.Request.Body == "" {
				return item, true
			}

			itemRawBody, err := getStubRequest(item.Request.Body)
			if err != nil {
				log.Printf("StubBodyDecoderError: %v\n", err)
			}

			equalRaws := reflect.DeepEqual(requestRawBody, itemRawBody)
			if equalRaws {
				return item, true
			}
		}
	}
	return exist, false
}

func getStubRequest(fileName string) (map[string]interface{}, error) {
	var itemRawBody map[string]interface{}
	requestStub, err := os.Open(fmt.Sprintf("static/%v", fileName))
	if err != nil {
		return itemRawBody, err
	}
	defer requestStub.Close()

	byteVale, _ := ioutil.ReadAll(requestStub)
	if err := json.Unmarshal(byteVale, &itemRawBody); err != nil {
		return itemRawBody, err
	}

	return itemRawBody, nil
}

func getConfig() (Config, error) {
	var config Config
	jsonFile, errFile := os.Open("jessica.json")
	if errFile != nil {
		return config, errFile
	}
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)
	if err := json.Unmarshal(byteValue, &config); err != nil {
		return config, err
	}

	return config, nil
}

// endregion

// #region Handlers

func aboutHandler(w http.ResponseWriter, r *http.Request) {
	resp := map[string]interface{}{
		"name":      AppName,
		"version":   AppVersion,
		"codename":  AppCodename,
		"copyright": AppCopyright,
	}
	Respond(w, resp)
}

func corsHandler(handler http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		config, err := getConfig()
		if err != nil {
			log.Printf("ConfigurationError: %v\n\n", err)
			Respond(w, Message(fmt.Sprintf("ConfigurationError: %v", err)))
			return
		}

		w.Header().Set("Access-Control-Allow-Origin", config.AllowedOrigins)
		w.Header().Add("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		if req.Method == "OPTIONS" {
			w.Header().Set("Access-Control-Allow-Methods", config.AllowedMethods)
			w.Header().Set("Access-Control-Allow-Headers", config.AllowedHeaders)
			return
		} else {
			handler.ServeHTTP(w, req)
		}
	}
	return http.HandlerFunc(fn)
}

func staticFilesHandler(fs http.Handler, mux http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		startTime := time.Now()
		log.Printf("Started %v \"%v\" at %v", req.Method, req.URL.Path, startTime.Format(time.RFC3339))

		config, err := getConfig()
		if err != nil {
			log.Printf("ConfigurationError: %v\n", err)
			Respond(w, Message(fmt.Sprintf("ConfigurationError: %v", err)))
			return
		}

		stubs := config.Stubs
		stub, found := findStub(stubs, req)
		if found {
			fsHandler := http.StripPrefix("", fs)
			log.Printf("Processing by %v", stub.Response.Content)

			if stub.Response.Content != "" {
				// rewrite url to static file
				req.URL.Path = stub.Response.Content
			}

			if stub.Response.ContentType != "" {
				w.Header().Add("Content-Type", stub.Response.ContentType)
			}

			var completedStatus int
			if stub.Response.Status != 0 {
				completedStatus = stub.Response.Status
				w.WriteHeader(stub.Response.Status)
			} else {
				completedStatus = 200
			}

			fsHandler.ServeHTTP(w, req)
			elapsed := time.Since(startTime)

			log.Printf("Completed %v %v in %v\n\n", completedStatus, http.StatusText(completedStatus), elapsed)
		} else if mux != nil {
			log.Printf("RoutingError: No stub matches [%v] \"%v\"\n\n", req.Method, req.URL.Path)
			mux.ServeHTTP(w, req)
		} else {
			http.Error(w, "Page Not Found", http.StatusNotFound)
		}
	}

	return http.HandlerFunc(fn)
}

// endregion

func main() {

	config, err := getConfig()
	if err != nil {
		log.Fatalf("ConfigurationError: %v", err)
	}

	log.Printf("=> %v %v application starting\n", AppName, AppVersion)
	log.Printf("* Mock version %v\n", config.Version)

	mux := http.NewServeMux()
	mux.HandleFunc("/jessica", aboutHandler)

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/", corsHandler(staticFilesHandler(fs, mux)))

	port := config.Port
	if port == "" {
		port = "5000"
	}

	log.Printf("* Listening on tcp://0.0.0.0:%v\n", port)
	log.Printf("Use Ctrl-C to stop\n\n")

	log.Fatal(http.ListenAndServe(":"+port, nil))
}
