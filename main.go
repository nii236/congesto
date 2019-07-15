package main

import (
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/victorspringer/http-cache"
	"github.com/victorspringer/http-cache/adapter/memory"
)

var previousStatus *Status
var nextStatus *Status

func init() {
	nextStatus = &Status{RWMutex: &sync.RWMutex{}}
	previousStatus = &Status{RWMutex: &sync.RWMutex{}}
}
func main() {
	conn, err := initDB(false)
	if err != nil {
		fmt.Println(err)
		return
	}
	memcached, err := memory.NewAdapter(
		memory.AdapterWithAlgorithm(memory.LRU),
		memory.AdapterWithCapacity(10000000),
	)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	cacheClient, err := cache.NewClient(
		cache.ClientWithAdapter(memcached),
		cache.ClientWithTTL(1*time.Second),
		cache.ClientWithRefreshKey("opn"),
	)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	handleStatus := cacheClient.Middleware(http.HandlerFunc(handleStatus))
	r.Get("/", handleStatus.ServeHTTP)
	r.Post("/subscribe", handleSubscribe(conn))
	r.Get("/metrics", promhttp.Handler().ServeHTTP)

	fmt.Println("Start server on :8080")
	http.ListenAndServe(":8080", r)
}

// Status is the full app status
type Status struct {
	*sync.RWMutex
	Regions RegionSlice
}

// RegionSlice state
type RegionSlice []*Region

// Region state
type Region struct {
	Name        string        `json:"name"`
	DataCentres []*DataCentre `json:"data_centres"`
}

// DataCentre state
type DataCentre struct {
	Name    string    `json:"name"`
	Servers []*Server `json:"servers"`
}

// Server state
type Server struct {
	Name                     string   `json:"name"`
	Category                 Category `json:"category"`
	CreateCharacterAvailable bool     `json:"create_character_available"`
}
