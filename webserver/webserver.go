package webserver

import (
	"../model"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io"
	"log"
	"net/http"
	"time"
)

// WebserverData is wrapper for the data we send to the UI
// For now only contains the docker stacks
type WebserverData struct {
	Servers map[string]model.FeiwuMessageOrigin
}

// UpdateServers replaces the stacks currently held by the webserverdata
func (wd *WebserverData) UpdateServers(server model.FeiwuMessageOrigin) {
	wd.Servers[server.HostName] = server
}

// HandleGetServers handles the /stacks calls
func (wd *WebserverData) HandleGetServers(w http.ResponseWriter, r *http.Request) {
	if len(wd.Servers) == 0 {
		io.WriteString(w, `[]`)
	} else {
		json.NewEncoder(w).Encode(wd.Servers)
	}
}

// Server is a wrapper object for managing the http router and logger
type Server struct {
	logger *log.Logger
	mux    *http.ServeMux
}

// StartServer starts the web server on the port given
// The channel bool is for telling the server to shutdown
func StartServer(port string, data *WebserverData, c chan bool) {
	router := mux.NewRouter()
	router.HandleFunc("/servers", data.HandleGetServers).Methods("GET")
	listenAddress := fmt.Sprintf(":%s", port)
	server := &http.Server{Addr: listenAddress, Handler: router}
	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()
	if b := <-c; b {
		fmt.Printf(" > [Webserver] We got told to quit\n")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		server.Shutdown(ctx)
		cancel()
	}
	c <- true
}
