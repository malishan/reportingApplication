package config

import (
	"encoding/json"
	"log"
	"os"
)

const (
	confFile = "./config.json"
)

// Conf struct holds service configuration
var Conf struct {
	MasterFile   string `json:"masterFile"`
	EmployeeFile string `json:"employeeFile"`
	HolidayFile  string `json:"holidayFile"`
	LogPath      string `json:"logPath"`
	BackUpMaster string `json:"backupMasterFile"`
}

func init() {
	loadConfig()
	validateConfig()
}

// loadConfig loads the configuration file
func loadConfig() {
	file, err := os.Open(confFile)
	if err != nil {
		log.Fatalln("unable to open conf file, error:", err)
	}
	defer file.Close()

	err = json.NewDecoder(file).Decode(&Conf)
	if err != nil {
		log.Fatalln("unable to decode conf file, error:", err)
	}

}

func validateConfig() {
	_, err := os.Stat(Conf.LogPath)

	if os.IsNotExist(err) {
		err = os.MkdirAll(Conf.LogPath, 0755)
		if err != nil {
			log.Fatalln("Failed to create Log Directory, err:", err)
		}
	}

	_, err = os.Stat(Conf.BackUpMaster)

	if os.IsNotExist(err) {
		err = os.MkdirAll(Conf.BackUpMaster, 0755)
		if err != nil {
			log.Fatalln("Failed to create Log Directory, err:", err)
		}
	}
}
