# reportingApplication
A backend webApp that serves the purpose of filing employee reports in an organization

NOTE: The following project was designed on MACOS

PREREQUISITES
install cassandra
install go
Input path for file is := ./static/sample_task.csv

BUILD
cd ~/go/src/reportingApp
env GOOS=windows GOARCH=amd64 (386) go build          ->  For Windows
env GOOS=linux go build                         ->  For Linux
env GOOS=linux GOARCH=arm GOARM=7 go build      ->  For RaspberryPI

FUNCTIONING STEPS
cassandra -f                                    -> To start the Cassandra server locally


RUN
cd ~/go/src/reportingApp
./reportingApp_mac                                   -> For MACOS
./reportingApp_linux                                 -> For LINUX


API
http://localhost:4040/reportingApp/health
http://localhost:4040/reportingApp/workInfo


ASSUMPTION IN THE PROGRAM
1. An organizer is also part of the attendees (assumption made as per the sample file)
2. Meeting times cannot start on one date and end on another
3. The maximum working hours per day is 8 hrs.


ENVIRONMENT VARIABLE

export CassandraHost=127.0.0.1
export CassandraKeyspace="reportingAppService"
export EmployeeDetails="employee_details"
export HolidayInfo="holiday_table"
export MasterData="master_table"
export ReportingAppSignKey="reportingApp_secret_key"
export userAdmin="uid=admin,ou=system"
export adminPassword="secret"
export BaseDN="dc=AppDomain,dc=local"