package main

import (
	"fmt"
	"go_backend/routes/index"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"go_backend/util"

	"github.com/gin-gonic/gin"
)

var quitServer chan os.Signal // by os.Signal Interrupt
var quitSignal chan struct{}  // manually

var serverEngine *gin.Engine
var rootServer *http.Server

// InitializeServerComponents Initializes Non blocking Close Channels
func InitializeServerComponents() {
	quitServer = make(chan os.Signal, 1)
	quitSignal = make(chan struct{}, 1)
}

func CreateServer(pLog *log.Logger, eLogger *log.Logger) {
	serverEngine = gin.Default()

	pLog.Println("Server Created")

	rootServer = &http.Server{
		Addr:         fmt.Sprintf("localhost:%d", util.GetPort()),
		ErrorLog:     eLogger,
		Handler:      serverEngine,
		WriteTimeout: time.Duration(0),
	}

	rootServer.RegisterOnShutdown(OnShutDown)

	signal.Notify(quitServer, os.Interrupt)
	signal.Notify(quitServer, os.Kill)
}

func InitializeGinEngine() {
	serverEngine.GET("/", index.Get)
}

func StartServer(pLog *log.Logger, eLog *log.Logger) {
	pLog.Println("Server Started")
	err := rootServer.ListenAndServe()
	util.CheckError(err, eLog)
}

func StopServer(pLog *log.Logger, eLog *log.Logger) {
	err := rootServer.Close()
	pLog.Println("Closing all channels")

	signal.Stop(quitServer)

	util.CheckError(err, eLog)
}

func OnShutDown() {
	fmt.Println("Oops Closed Server From Server")
}

func WatchCloseManually(pLog *log.Logger, eLog *log.Logger) {
	<-quitSignal
	pLog.Printf("Server Manually Closed")
	StopServer(pLog, eLog)
}

func WatchCloseOS(pLog *log.Logger, eLog *log.Logger) {
	<-quitServer
	pLog.Printf("Server Closed due to Interrupt")
	StopServer(pLog, eLog)
}

func main() {
	pLogger := util.GetLogger()
	eLogger := util.GetErrorLogger()

	InitializeServerComponents()   // Creates Close Channels
	CreateServer(pLogger, eLogger) // Creates server , server gin engine , initializes close channels
	InitializeGinEngine()          // Attaches routes to server gin engine
	go StartServer(pLogger, eLogger)
	go WatchCloseManually(pLogger, eLogger)
	go WatchCloseManually(pLogger, eLogger)

	time.Sleep(time.Duration(util.GetServerUpTime()) * time.Second)
	quitSignal <- struct{}{}

	time.Sleep(time.Duration(1) * time.Second)
}
