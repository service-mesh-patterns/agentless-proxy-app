package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/connect"
	"github.com/hashicorp/go-hclog"
	"github.com/nicholasjackson/env"
)

var name = env.String("SERVICE_NAME", false, "app", "Name of the service")
var bindAddress = env.String("BIND_ADDRESS", false, "0.0.0.0", "Bind address of the service")
var bindPort = env.Int("BIND_PORT", false, 9090, "Bind port of the service")
var upstream = env.String("UPSTREAM", false, "", "Upstream service to call when the service receives an inbound request")
var public = env.Bool("PUBLIC", false, false, "Does the service have a public API, if false the server is configured to use mTLS")
var address = env.String("IP_ADDRESS", false, "localhost", "IP address of the service")

func main() {
	log := hclog.Default()

	err := env.Parse()
	if err != nil {
		log.Error("Unable to parse environment", "error", err)
		os.Exit(1)
	}

	//	// Create a Consul API client
	client, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		log.Error("Unable to create Consul client", "error", err)
		os.Exit(1)
	}

	// register the service
	serviceID := fmt.Sprintf("%s-%d", *name, time.Now().UnixNano())
	client.Agent().ServiceRegister(&api.AgentServiceRegistration{
		Name:    *name,
		ID:      serviceID,
		Address: *address,
		Port:    *bindPort,
		Connect: &api.AgentServiceConnect{
			Native: true,
		},
	})

	// Create an instance representing this service. "my-service" is the
	// name of _this_ service. The service should be cleaned up via Close.
	svc, err := connect.NewService(*name, client)
	if err != nil {
		log.Error("Unable to register service", "error", err)
		os.Exit(1)
	}
	defer svc.Close()

	// register the handler
	http.HandleFunc("/", handleRoot(svc, *upstream, log))

	log.Info("Starting server", "address", *bindAddress, "port", *bindPort)
	if !*public {
		// The service is not public so create a HTTP server that serves via Connect
		server := &http.Server{
			Addr:      fmt.Sprintf("%s:%d", *bindAddress, *bindPort),
			TLSConfig: svc.ServerTLSConfig(),
			Handler:   http.DefaultServeMux,
			// ... other standard fields
		}

		go server.ListenAndServeTLS("", "")
	} else {
		// Start a public server
		go http.ListenAndServe(fmt.Sprintf("%s:%d", *bindAddress, *bindPort), nil)
	}

	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigs
		fmt.Println()
		fmt.Println(sig)
		done <- true
	}()

	fmt.Println("awaiting signal")
	<-done
	fmt.Println("exiting")

	client.Agent().ServiceDeregister(serviceID)

}

type Response struct {
	Service      string    `json:"service"`
	UpstreamCall *Response `json:"upstream_call,omitempty"`
}

func handleRoot(svc *connect.Service, upstream string, log hclog.Logger) func(rw http.ResponseWriter, r *http.Request) {
	return func(rw http.ResponseWriter, r *http.Request) {
		resp := Response{Service: svc.Name()}

		if upstream != "" {
			r, err := svc.HTTPClient().Get(upstream)
			if err != nil || r == nil || r.StatusCode != http.StatusOK {
				http.Error(rw, fmt.Sprintf("Unable to contact upstream, error: %s", err), http.StatusInternalServerError)
				return
			}

			upstreamResponse := &Response{}
			err = json.NewDecoder(r.Body).Decode(upstreamResponse)
			if err != nil {
				http.Error(rw, "Unable to decode upstream response", http.StatusInternalServerError)
				return
			}

			resp.UpstreamCall = upstreamResponse
		}

		err := json.NewEncoder(rw).Encode(resp)
		if err != nil {
			log.Error("Unable to write response", "error", err)
		}
	}
}
