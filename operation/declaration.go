package operation

const (
	layoutDate     = "1/2/2006 MST"
	layoutTime     = "1/2/2006 3:4:5 PM"
	layoutOnlyTime = "03:04:05"
)

const (
	internal = "internal"
	external = "external"
	all      = "all"
)

const (
	TEAMLEAD      = "Team Leader"
	MANAGER       = "Manager"
	SNMANAGER     = "Senior Manager"
	GENMANAGER    = "General Manager"
	VICEPRESIDENT = "Vice President"
)

const (
	TEAM       = "team"
	SECTION    = "section"
	DEPARTMENT = "department"
	DIVISION   = "division"
	UNIT       = "unit"
)

// EmpManager forms the struct to hold an employee's supervisiors
type EmpManager struct {
	employeeID string `bson:"employeeID"`
	unit       string `bson:"unit"`
	division   string `bson:"division"`
	department string `bson:"department"`
	section    string `bson:"section"`
	team       string `bson:"team"`
}

// Employee forms the struct to be stored in DB
type Employee struct {
	StartTime      string `bson:"startTime"`
	EndTime        string `bson:"endTime"`
	Duration       string `bson:"duration"`
	OrganizerID    string `bson:"organizerID"`
	EmployeeID     string `bson:"employeeID"`
	MeetingSubject string `bson:"meetingSubject"`
}

// EmpInfoResponse forms the structure for the API response
type EmpInfoResponse struct {
	EmployeeID           string  `json:"employeeID"`
	MeetingHours         float64 `json:"meetingTime"`
	WorkingHours         float64 `json:"workingHours"`
	MeetingPercentage    float64 `json:"meetingPercentage"`
	WorkingPercentage    float64 `json:"workingPercentage"`
	ExternalMeetingHours float64 `json:"externalMeetingHours"`
	InternalMeetingHours float64 `json:"internalMeetingHours"`
	InternalMeetingCount int     `json:"internalMeetingCount"`
	ExternalMeetingCount int     `json:"externalMeetingCount"`
	TotalMeetingCount    int     `json:"totalMeetingCount"`
}

// PerEmployeeInfo forms the struct to hold its detailed info
type PerEmployeeInfo struct {
	MeetingSubject string `json:"meetingSubject"`
	StartTime      string `json:"startTime"`
	EndTime        string `json:"endTime"`
	OrganizerID    string `bjson:"organizerID"`
	AttendeeID     string `json:"attendeeID"`
	Duration       string `bson:"duration"`
}

// DetailedInfoResponse holds info per employee
type DetailedInfoResponse struct {
	EmployeeID string            `json:"employeeID"`
	Info       []PerEmployeeInfo `json:"info"`
}

// cumResposne holds the info per month basis
type cumResposne struct {
	month        string
	meetingHours float64
}

// ch channel enables communication for receiving monthly responses
var ch chan cumResposne
