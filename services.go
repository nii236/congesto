package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/jmoiron/sqlx"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	cache "github.com/victorspringer/http-cache"
	"github.com/victorspringer/http-cache/adapter/memory"
)

func StartBot(conn *sqlx.DB, botToken string, checkInterval time.Duration) error {
	b, err := NewBot(conn, botToken, checkInterval)
	if err != nil {
		return err
	}
	go b.Run()
	go b.RunChecker()
	return nil
}

func StartAPI(conn *sqlx.DB, addr string) {
	memcached, err := memory.NewAdapter(
		memory.AdapterWithAlgorithm(memory.LRU),
		memory.AdapterWithCapacity(100000),
	)
	if err != nil {
		logger.Error(err)
		os.Exit(1)
	}

	cacheClient, err := cache.NewClient(
		cache.ClientWithAdapter(memcached),
		cache.ClientWithTTL(1*time.Second),
		cache.ClientWithRefreshKey("opn"),
	)
	if err != nil {
		logger.Error(err)
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

	logger.Info(fmt.Sprintf("Start server on %s", addr))
	go func(r chi.Router) { log.Fatalln(http.ListenAndServe(addr, r)) }(r)
}
