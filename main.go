package main

import (
	"project/reportingApplication/config"
	"project/reportingApplication/operation"
	"project/reportingApplication/server"
	"project/reportingApplication/utils"
)

const (
	version = "0.0.4"
)

func main() {
	utils.Log("\n\nSTARTING REPORTING APP SERVER, VERSION: ", version)

	printConfiguration()

	go operation.LoadFile()
	go operation.LoadHolidayFile()

	server.Run()
}

func printConfiguration() {
	utils.Log("<< Configuration >>")
	utils.Log("MasterFilePath :", config.Conf.MasterFile, "EmployeeFilePath :", config.Conf.EmployeeFile, "HolidayFilePath :", config.Conf.HolidayFile, "LogPath :", config.Conf.LogPath, "backupMasterFile :", config.Conf.BackUpMaster)
}
