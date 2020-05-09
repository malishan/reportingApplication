package utils

import (
	"log"
	"os"
	"project/reportingApplication/config"
	"time"
)

var (
	// Logger used for logging
	Logger *log.Logger
)

func init() {
	date := time.Now().Format("2006-01-02")
	logPath := config.Conf.LogPath + date + ".log"

	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0755)
	if err != nil {
		log.Fatalln("could not open log file, ", "error:", err)
	}
	Logger = log.New(f, "[reportingAppService] ", log.LstdFlags)
}

// Log logs to log-file
func Log(v ...interface{}) {
	Logger.Println(v...)
}

// Logf logs to log-file
func Logf(format string, v ...interface{}) {
	Logger.Printf(format+"\n", v...)
}

// Fatalln logs to log-file and then exits
func Fatalln(v ...interface{}) {
	Logger.Fatalln(v...)
}

// Fatalf logs to log-file and then exits
func Fatalf(format string, v ...interface{}) {
	Logger.Fatalf(format+"\n", v...)
}

// CheckError checks if there is error and then panics
func CheckError(err error) {
	if err != nil {
		Log("Panic :", err)
		panic(err)
	}
}
