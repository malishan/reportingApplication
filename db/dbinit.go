package db

import (
	"os"
	"project/reportingApplication/utils"
	"time"

	"github.com/gocql/gocql"
)

var (
	Host            string //"127.0.0.1"
	Keyspace        string
	EmployeeTable   string
	HolidayTable    string
	MasterDataTable string
)

func init() {
	loadEnvVariable()
	session, err := ConnectToCassandra()
	utils.CheckError(err)

	defer session.Close()

	createDatabase(session)
}

func loadEnvVariable() {
	Host = os.Getenv("CassandraHost")
	Keyspace = os.Getenv("CassandraKeyspace")
	EmployeeTable = os.Getenv("EmployeeDetails")
	HolidayTable = os.Getenv("HolidayInfo")
	MasterDataTable = os.Getenv("MasterData")

	if Host == "" || Keyspace == "" || EmployeeTable == "" || HolidayTable == "" || MasterDataTable == "" {
		utils.Fatalln("environment variables not set correctly")
	}
}

// ConnectToCassandra creates a new session for cassandra
func ConnectToCassandra() (*gocql.Session, error) {
	cluster := gocql.NewCluster(Host)
	cluster.ConnectTimeout = 1 * time.Minute
	cluster.Timeout = 2 * time.Minute
	return cluster.CreateSession()
}

// createDatabase initializes db and creates required tables
func createDatabase(session *gocql.Session) {

	// create keyspace
	err := session.Query("CREATE KEYSPACE IF NOT EXISTS " + Keyspace + " WITH replication = {'class' : 'SimpleStrategy', 'replication_factor' : 1} AND durable_writes = 'true';").Exec()
	if err != nil {
		utils.Fatalf("failed to create keyspace: %s, err : %s", Keyspace, err.Error())
	}

	// drop master_data table
	err = session.Query("DROP TABLE IF EXISTS " + Keyspace + "." + MasterDataTable).Exec()
	if err != nil {
		utils.Fatalf("failed to drop master data table: %s, err : %s", MasterDataTable, err.Error())
	}

	// create master_data table
	err = session.Query("CREATE TABLE IF NOT EXISTS " + Keyspace + "." + MasterDataTable + " (employeeID text, unit text, divisions text, department text, section text, team text, PRIMARY KEY(employeeID));").Exec()
	if err != nil {
		utils.Fatalf("failed to create master data table: %s, err : %s", MasterDataTable, err.Error())
	}

	// create employee_details table
	err = session.Query("CREATE TABLE IF NOT EXISTS " + Keyspace + "." + EmployeeTable + " (unit text, divisions text, department text, section text, team text, organizerID text, employeeID text, startTime text, endTime text, duration text, meetingSubject text, PRIMARY KEY(employeeID, startTime));").Exec()
	if err != nil {
		utils.Fatalf("failed to create employee info table: %s, err : %s", EmployeeTable, err.Error())
	}

	// create holiday_info table
	err = session.Query("CREATE TABLE IF NOT EXISTS " + Keyspace + "." + HolidayTable + " (serailNo text, dates date, PRIMARY KEY(dates));").Exec()
	if err != nil {
		utils.Fatalf("failed to create holiday info table: %s, err : %s", HolidayTable, err.Error())
	}
}
