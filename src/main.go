package main

import (
	"flag"
	"fmt"
	"github.com/Power-LAB/PeerVault/communication/control"
	"github.com/Power-LAB/PeerVault/communication/event"
	"github.com/Power-LAB/PeerVault/communication/peer"
	"github.com/Power-LAB/PeerVault/crypto"
	"github.com/Power-LAB/PeerVault/database"
	"github.com/op/go-logging"
	"os"
)

var (
	log = logging.MustGetLogger("peerVaultLogger")
)

// Example format string. Everything except the message has a custom color
// which is dependent on the log level. Many fields have a custom output
// formatting too, eg. the time returns the hour down to the milli second.
var format = logging.MustStringFormatter(
	"%{color}%{time:15:04:05.000} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}",
)

// Password is just an example type implementing the Redactor interface. Any
// time this is logged, the Redacted() function will be called.
type Password string

func (p Password) Redacted() interface{} {
	return logging.Redact(string(p))
}

func main() {
	debugMode := flag.Bool("dev", false, "Enable dev mode")
	apiAddress := flag.String("apiAddr", "localhost:4444", "http api service address")
	wsAddress := flag.String("wsAddr", "localhost:5555", "WebSocket event service address")
	relayHost := flag.String("relay", "", "Relay Host URL")
	logLevel := flag.Int("log", 9, "Log level, 3=error,warning 6=notice,info, 9=debug")
	dbFilePath := flag.String("bbolt", "", "Location of bbolt DB file")
	logFilePath := flag.String("logfile", "", "Location of log file")
	flag.Parse()

	configureLogger(*logFilePath, *logLevel)

	if *relayHost == "" {
		log.Fatal("Please provide relay host with --relay option")
	}

	if *dbFilePath == "" {
		log.Fatal("Please provide db file path with --bbolt option")
	}

	if *debugMode {
		crypto.EnableDevMode()
	}

	database.SetDbPath(*dbFilePath)

	run(wsAddress, apiAddress, relayHost)
}

func configureLogger(logFilePath string, logLevel int) {
	var logfile *os.File
	var err error
	if logFilePath != "" {
		logfile, err = os.OpenFile(
			logFilePath,
			os.O_RDWR|os.O_APPEND|os.O_CREATE,
			0600,
		)
		if err != nil {
			panic("Cannot open log file")
		}
	} else {
		logfile = os.Stderr
	}

	logging.SetFormatter(format)

	backendLog := logging.NewLogBackend(logfile, "", 0)
	// Only errors and more severe messages should be sent to backend1
	backendLogLeveled := logging.AddModuleLevel(backendLog)

	if logLevel >= 3 {
		backendLogLeveled.SetLevel(logging.ERROR, "")
	}
	if logLevel >= 6 {
		backendLogLeveled.SetLevel(logging.INFO, "")
	}
	if logLevel >= 9 {
		backendLogLeveled.SetLevel(logging.DEBUG, "")
		fmt.Printf(
			"\033[%dm%s\033[0m",
			int(logging.ColorRed),
			"!!! ATTENTION !!!\nDEBUG LOGGING MAY CONTAIN SENSIBLE INFORMATION SUCH AS CLEAR PRIVATE " +
				"KEY OR ANY DATA.\nIT SHOULD ONLY BE USED IN DEVELOPPER MODE\n\n")
	}
	logging.SetBackend(backendLogLeveled)
}

func run(wsAddress *string, apiAddress *string, relayHost *string) {
	// Start websocket events
	go event.Listen(wsAddress)

	// Start API
	go control.Listen(apiAddress)

	// Start peer
	go peer.Listen(relayHost)

	select {}
}