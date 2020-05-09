package operation

import (
	"project/reportingApplication/db"
	"project/reportingApplication/utils"
	"strings"
	"sync"
	"time"

	"github.com/gocql/gocql"
)

// insertoMasterDB inserts records into the master_data table
func insertoMasterDB(employee, unit, division, department, section, team string) {

	session, err := db.ConnectToCassandra()
	utils.CheckError(err)

	defer session.Close()

	query := "INSERT INTO " + db.Keyspace + "." + db.MasterDataTable + " (employeeID,unit,divisions,department,section,team) VALUES (?, ?, ?, ?,?,?) IF NOT EXISTS"

	err = session.Query(query, employee, unit, division, department, section, team).Exec()
	if err != nil {
		utils.Fatalf("Failed to add employee info to masterDB : %s", err.Error())
	}
}

// insertToEmployeeDb inserts records into the employee_details table
func insertToEmployeeDb(emp Employee) {

	session, err := db.ConnectToCassandra()
	utils.CheckError(err)

	defer session.Close()

	manager := fetchEmployeeSupervisior(emp.EmployeeID)
	query := "INSERT INTO " + db.Keyspace + "." + db.EmployeeTable + " (unit,divisions,department,section,team,organizerID,employeeID,startTime,endTime,duration,meetingSubject) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) IF NOT EXISTS"
	err = session.Query(query, manager.unit, manager.division, manager.department, manager.section, manager.team, emp.OrganizerID, emp.EmployeeID, emp.StartTime, emp.EndTime, emp.Duration, emp.Duration).Exec()
	if err != nil {
		utils.Fatalf("Failed to add employee info to db : %s", err.Error())
	}

	//wg.Done()
}

// insertToHolidayDB inserts records into the holiday_info table
func insertToHolidayDB(wg *sync.WaitGroup, serailNo, date string) {

	session, err := db.ConnectToCassandra()
	utils.CheckError(err)

	defer session.Close()

	//WHERE date = '2017-01-25'

	query := "INSERT INTO " + db.Keyspace + "." + db.HolidayTable + " (serailNo,dates) VALUES (?, ?) IF NOT EXISTS"
	err = session.Query(query, serailNo, date).Exec()
	if err != nil {
		utils.Fatalf("Failed to add holiday info to db : %s", err.Error())
	}

	wg.Done()
}

// fetchEmployeeSupervisior retrieves employee's supervisiors info from db
func fetchEmployeeSupervisior(employeeID string) *EmpManager {

	session, err := db.ConnectToCassandra()
	utils.CheckError(err)

	defer session.Close()

	var head *EmpManager

	query := "SELECT * FROM " + db.Keyspace + "." + db.MasterDataTable + " WHERE employeeID=?"
	iterable := session.Query(query, employeeID).Iter()

	for {

		row := make(map[string]interface{})
		if !iterable.MapScan(row) {
			break
		}

		head = &EmpManager{
			employeeID: employeeID,
			unit:       row["unit"].(string),
			division:   row["divisions"].(string),
			department: row["department"].(string),
			section:    row["section"].(string),
			team:       row["team"].(string),
		}

		break
	}

	return head
}

// fetchAllEmployeeFromDb returns all instances of an employee from db
func fetchAllEmployeeFromDb(id, empType string) []Employee {

	var empList []Employee

	session, err := db.ConnectToCassandra()
	utils.CheckError(err)

	defer session.Close()

	var query string

	if empType == TEAM {
		query = "SELECT * FROM " + db.Keyspace + "." + db.EmployeeTable + " WHERE team=? ALLOW FILTERING"
	} else if empType == SECTION {
		query = "SELECT * FROM " + db.Keyspace + "." + db.EmployeeTable + " WHERE section=? ALLOW FILTERING"
	} else if empType == DEPARTMENT {
		query = "SELECT * FROM " + db.Keyspace + "." + db.EmployeeTable + " WHERE department=? ALLOW FILTERING"
	} else if empType == DIVISION {
		query = "SELECT * FROM " + db.Keyspace + "." + db.EmployeeTable + " WHERE divisions=? ALLOW FILTERING"
	} else if empType == UNIT {
		query = "SELECT * FROM " + db.Keyspace + "." + db.EmployeeTable + " WHERE unit=? ALLOW FILTERING"
	} else {
		query = "SELECT * FROM " + db.Keyspace + "." + db.EmployeeTable + " WHERE employeeID=?"
	}

	iterable := session.Query(query, id).Iter()

	utils.Logf("Number of matched rows : %v", iterable.NumRows())

	for {

		row := make(map[string]interface{})
		if !iterable.MapScan(row) {
			break
		}

		start, _ := row["starttime"].(string)
		end, _ := row["endtime"].(string)
		organizer, _ := row["organizerid"].(string)

		empList = append(empList, Employee{
			StartTime:   start,
			EndTime:     end,
			OrganizerID: organizer,
		})
	}

	return empList
}

// fetchHolidayDate retrieves the holidays from the holiday_info table excluding weekends
func fetchHolidayDate(start, end string) (int, error) {
	count := 0

	session, err := db.ConnectToCassandra()
	utils.CheckError(err)
	defer session.Close()

	query := "SELECT dates FROM " + db.Keyspace + "." + db.HolidayTable + " WHERE dates >= ? AND dates <= ? ALLOW FILTERING"
	iterable := session.Query(query, start, end).Iter()

	for {
		row := make(map[string]interface{})
		if !iterable.MapScan(row) {
			break
		}

		date := row["dates"].(time.Time)

		if date.Weekday() == 5 || date.Weekday() == 6 {
			continue
		}

		count++
	}

	return count, nil
}

// fetchMonthlyMeetingHours returns the monthly meeting duration from db
func fetchMonthlyMeetingHours(month, unit, start, end string) {
	var (
		query        string
		iterable     *gocql.Iter
		meetingHours float64
	)

	session, err := db.ConnectToCassandra()
	utils.CheckError(err)

	defer session.Close()

	if unit == all {
		query = "SELECT duration FROM " + db.Keyspace + "." + db.EmployeeTable + " WHERE starttime >= ? AND endtime < ? ALLOW FILTERING"
		iterable = session.Query(query, start, end).Iter()
	} else {
		query = "SELECT duration FROM " + db.Keyspace + "." + db.EmployeeTable + " WHERE unit = ? AND starttime >= ? AND endtime < ? ALLOW FILTERING"
		iterable = session.Query(query, unit, start, end).Iter()
	}

	for {
		row := make(map[string]interface{})
		if !iterable.MapScan(row) {
			break
		}
		duration, _ := row["duration"].(string)

		tArr := strings.Split(duration, ":")

		format := tArr[0] + "h" + tArr[1] + "m" + tArr[2] + "s"

		dur, _ := time.ParseDuration(format)

		meetingHours += dur.Hours()
	}

	rslt := cumResposne{
		month:        month,
		meetingHours: meetingHours,
	}

	ch <- rslt
}

// fetchEmpDetailedRpt retrieves the detailed report of the employee from db
func fetchEmpDetailedRpt(id, empType, startDate, endDate string) (*DetailedInfoResponse, error) {

	var (
		details = &DetailedInfoResponse{}
	)

	session, err := db.ConnectToCassandra()
	utils.CheckError(err)

	defer session.Close()

	var query string

	if empType == TEAM {
		query = "SELECT * FROM " + db.Keyspace + "." + db.EmployeeTable + " WHERE team=? ALLOW FILTERING"
	} else if empType == SECTION {
		query = "SELECT * FROM " + db.Keyspace + "." + db.EmployeeTable + " WHERE section=? ALLOW FILTERING"
	} else if empType == DEPARTMENT {
		query = "SELECT * FROM " + db.Keyspace + "." + db.EmployeeTable + " WHERE department=? ALLOW FILTERING"
	} else if empType == DIVISION {
		query = "SELECT * FROM " + db.Keyspace + "." + db.EmployeeTable + " WHERE divisions=? ALLOW FILTERING"
	} else if empType == UNIT {
		query = "SELECT * FROM " + db.Keyspace + "." + db.EmployeeTable + " WHERE unit=? ALLOW FILTERING"
	} else {
		query = "SELECT * FROM " + db.Keyspace + "." + db.EmployeeTable + " WHERE employeeID=?"
	}

	iterable := session.Query(query, id).Iter()

	for {
		row := make(map[string]interface{})
		if !iterable.MapScan(row) {
			break
		}

		meetingSubject, _ := row["meetingSubject"].(string)
		startDate, _ := row["starttime"].(string)
		endDate, _ := row["endtime"].(string)
		organizer, _ := row["organizerid"].(string)
		attendees, _ := row["employeeid"].(string)
		duration, _ := row["duration"].(string)

		perEmp := PerEmployeeInfo{
			MeetingSubject: meetingSubject,
			StartTime:      startDate,
			EndTime:        endDate,
			OrganizerID:    organizer,
			AttendeeID:     attendees,
			Duration:       duration,
		}

		details.EmployeeID = id
		details.Info = append(details.Info, perEmp)
	}

	return details, nil
}
