package operation

import (
	"encoding/csv"
	"os"
	"project/reportingApplication/config"
	"project/reportingApplication/utils"
	"strings"
	"sync"
	"time"
)

// loadMasterData reads the master records from the csv file
func loadMasterData() {

	file, err := os.Open(config.Conf.MasterFile)
	if err != nil {
		utils.Fatalf("Cannot open %s : %s", config.Conf.MasterFile, err.Error())
	}
	defer file.Close()

	csvReader := csv.NewReader(file)

	backupPath := config.Conf.BackUpMaster + time.Now().Format("2006-01-02") + ".csv"

	backupFile, err := os.OpenFile(backupPath, os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		utils.Fatalln("could not open log file, ", "error:", err)
	}
	defer backupFile.Close()

	csvWriter := csv.NewWriter(backupFile)

	rows, err := csvReader.ReadAll()
	if err != nil {
		utils.Fatalf("CSV reading file %s : %s", config.Conf.MasterFile, err.Error())
	}

	err = csvWriter.WriteAll(rows)
	if err != nil {
		utils.Fatalf("CSV writing to file %s : %s", backupPath, err.Error())
	}

	insertMasterRecords(rows)
}

// LoadFile reads the employee details from the csv file
func LoadFile() {

	loadMasterData()

	file, err := os.Open(config.Conf.EmployeeFile)
	if err != nil {
		utils.Fatalf("Cannot open %s : %s", config.Conf.EmployeeFile, err.Error())
	}
	defer file.Close()

	csvReader := csv.NewReader(file)

	rows, err := csvReader.ReadAll()
	if err != nil {
		utils.Fatalf("CSV reading file %s : %s", config.Conf.EmployeeFile, err.Error())
	}

	insertEmployeeRecords(rows)
}

// LoadHolidayFile reades the holidays from the csv file
func LoadHolidayFile() {
	file, err := os.Open(config.Conf.HolidayFile)
	if err != nil {
		utils.Fatalf("Cannot open %s : %s", config.Conf.HolidayFile, err.Error())
	}
	defer file.Close()

	csvReader := csv.NewReader(file)

	rows, err := csvReader.ReadAll()
	if err != nil {
		utils.Fatalf("CSV reading file %s : %s", config.Conf.HolidayFile, err.Error())
	}

	insertInHolidayDB(rows)
}

// insertMasterRecords reads master file and stores info in db
func insertMasterRecords(rows [][]string) {

	for index, eachRow := range rows {
		if index == 0 {
			continue
		}

		if len(eachRow) < 5 {
			utils.Fatalln("Less number of columns in master records")
		}

		var (
			teamLead, manager, srManager, genManager, vicePresident string
		)

		title := eachRow[2]
		tlReprtsTo := eachRow[4]

		switch {

		case strings.Contains(title, VICEPRESIDENT):

		case strings.Contains(title, GENMANAGER):
			vicePresident, _ = getVicePresident(tlReprtsTo, rows)

		case strings.Contains(title, SNMANAGER):
			genManager, tlReprtsTo = getGenManager(tlReprtsTo, rows)

			if genManager != "" {
				vicePresident, _ = getVicePresident(tlReprtsTo, rows)
			}

		case strings.Contains(title, MANAGER):
			srManager, tlReprtsTo = getSrManager(tlReprtsTo, rows)

			if srManager != "" {
				genManager, tlReprtsTo = getGenManager(tlReprtsTo, rows)
			}

			if genManager != "" {
				vicePresident, _ = getVicePresident(tlReprtsTo, rows)
			}
		case strings.Contains(title, TEAMLEAD):
			manager, tlReprtsTo = getManager(tlReprtsTo, rows)

			if manager != "" {
				srManager, tlReprtsTo = getSrManager(tlReprtsTo, rows)
			}

			if srManager != "" {
				genManager, tlReprtsTo = getGenManager(tlReprtsTo, rows)
			}

			if genManager != "" {
				vicePresident, _ = getVicePresident(tlReprtsTo, rows)
			}

		default:
			teamLead, tlReprtsTo = getTeamLead(tlReprtsTo, rows)

			if teamLead != "" {
				manager, tlReprtsTo = getManager(tlReprtsTo, rows)
			}

			if manager != "" {
				srManager, tlReprtsTo = getSrManager(tlReprtsTo, rows)
			}

			if srManager != "" {
				genManager, tlReprtsTo = getGenManager(tlReprtsTo, rows)
			}

			if genManager != "" {
				vicePresident, _ = getVicePresident(tlReprtsTo, rows)
			}
		}

		//employeeId, unit, division, department, section, team
		insertoMasterDB(eachRow[0], vicePresident, genManager, srManager, manager, teamLead)
	}

	utils.Log("Master records saved to database")
}

// insertEmployeeRecords makes employee details entry into the db
func insertEmployeeRecords(rows [][]string) {

	//var wg sync.WaitGroup

	for index, eachRow := range rows {

		if index == 0 {
			continue
		}

		if len(eachRow) < 6 {
			utils.Fatalln("Less number of columns in employee details records")
		}

		str := strings.TrimSuffix(eachRow[5], "; ")
		attendees := strings.Split(str, "; ")

		for _, v := range attendees {
			emp := Employee{
				StartTime:      eachRow[0],
				EndTime:        eachRow[1],
				Duration:       eachRow[2],
				OrganizerID:    eachRow[3],
				EmployeeID:     v,
				MeetingSubject: eachRow[4],
			}

			//wg.Add(1)
			insertToEmployeeDb(emp)
		}
	}

	//wg.Wait()
	utils.Log("Employee records saved to database")
}

//insertInHolidayDB makes holidays info entry into the db
func insertInHolidayDB(rows [][]string) {

	var wg sync.WaitGroup

	for index, v := range rows {
		if index == 0 {
			continue
		}

		wg.Add(1)
		go insertToHolidayDB(&wg, v[0], v[1])
	}

	wg.Wait()

	utils.Log("Holiday records saved to database")
}

func getTeamLead(lead string, rows [][]string) (string, string) {

	reportsTo := ""

	for _, r := range rows {
		if r[0] == lead {
			reportsTo = r[4]

			if strings.Contains(r[2], TEAMLEAD) {
				return lead, reportsTo
			}
		}
	}

	return "", reportsTo
}

func getManager(manager string, rows [][]string) (string, string) {

	reportsTo := ""

	for _, r := range rows {
		if r[0] == manager {
			reportsTo = r[4]

			if strings.Contains(r[2], MANAGER) {
				return manager, reportsTo
			}
		}
	}

	return "", reportsTo
}

func getSrManager(srManager string, rows [][]string) (string, string) {
	reportsTo := ""

	for _, r := range rows {
		if r[0] == srManager {
			reportsTo = r[4]

			if strings.Contains(r[2], SNMANAGER) {
				return srManager, reportsTo
			}
		}
	}

	return "", reportsTo
}

func getGenManager(grManager string, rows [][]string) (string, string) {

	reportsTo := ""

	for _, r := range rows {
		if r[0] == grManager {
			reportsTo = r[4]

			if strings.Contains(r[2], GENMANAGER) {
				return grManager, reportsTo
			}
		}
	}

	return "", reportsTo
}

func getVicePresident(vp string, rows [][]string) (string, string) {
	reportsTo := ""

	for _, r := range rows {
		if r[0] == vp {
			reportsTo = r[4]

			if strings.Contains(r[2], VICEPRESIDENT) {
				return vp, reportsTo
			}
		}
	}

	return "", reportsTo
}
