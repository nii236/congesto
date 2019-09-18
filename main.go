package main

import (
	"flag"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

var previousStatus *Status
var nextStatus *Status
var logger *zap.SugaredLogger
var (
	// FlagAddr for the server Addr
	FlagAddr string = ":8080"
	// FlagBotToken for the server BotToken
	FlagBotToken string
	// FlagDropDB to drop the database on start
	FlagDropDB bool = false

	// FlagCheckIntervalMinutes to check how often to check and notify subscribers
	FlagCheckIntervalMinutes int = 5
)

func init() {
	flag.StringVar(&FlagAddr, "addr", lookupEnvOrString("ADDR", FlagAddr), "Address to host from")
	flag.StringVar(&FlagBotToken, "bot-token", lookupEnvOrString("BOT_TOKEN", FlagBotToken), "Telegram token")
	flag.BoolVar(&FlagDropDB, "drop-db", lookupEnvOrBool("DROP_DB", FlagDropDB), "Drop the database")
	flag.IntVar(&FlagCheckIntervalMinutes, "check-interval-minutes", lookupEnvOrInt("CHECK_INTERVAL_MINUTES", FlagCheckIntervalMinutes), "How often to check the server status")

	flag.Parse()

	zlogger, _ := zap.NewDevelopment()
	defer zlogger.Sync() // flushes buffer, if any
	logger = zlogger.Sugar()

	nextStatus = &Status{RWMutex: &sync.RWMutex{}}
	previousStatus = &Status{RWMutex: &sync.RWMutex{}}
}
func main() {
	conn, err := initDB(FlagDropDB)
	if err != nil {
		fmt.Println(err)
		return
	}
	StartAPI(conn, FlagAddr)
	StartBot(conn, FlagBotToken, time.Duration(FlagCheckIntervalMinutes)*time.Minute)
	select {}
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
