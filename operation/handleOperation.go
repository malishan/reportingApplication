package operation

import (
	"fmt"
	"project/reportingApplication/utils"
	"strconv"
	"strings"
	"time"
)

// evaluate makes the calculation for the response
func evaluate(id string, day, meetingTime float64) *EmpInfoResponse {

	totalTimeInMins := day * 8 * 60
	workingTime := totalTimeInMins - meetingTime

	meetingPercentage := (meetingTime * 100) / totalTimeInMins
	workingPercentage := 100 - meetingPercentage

	workingHour, _ := time.ParseDuration(fmt.Sprintf("%f", workingTime) + "m")
	meetingHour, _ := time.ParseDuration(fmt.Sprintf("%f", meetingTime) + "m")

	resp := &EmpInfoResponse{
		EmployeeID:        id,
		MeetingHours:      meetingHour.Hours(),
		WorkingHours:      workingHour.Hours(),
		MeetingPercentage: meetingPercentage,
		WorkingPercentage: workingPercentage,
	}

	return resp
}

// calTimePeriod it excludes all weekends within the time given period
func calculateTimeDiff(start, end time.Time) int {
	days := 0

	for {
		if start.Equal(end) {
			break
		}

		if start.Weekday() != 5 && start.Weekday() != 6 {
			days++
		}

		start = start.Add(time.Hour * 24)
	}

	return days
}

// FetchResult creates response for the API
func FetchResult(id, startDate, endDate, empTyp string) (*EmpInfoResponse, error) {

	start, err := time.Parse(layoutDate, startDate+" GST") // GST - Gulf Standard Time Zone
	if err != nil {
		utils.Logf("Failed to parse startDate : %s", err.Error())
		return nil, err
	}

	end, err := time.Parse(layoutDate, endDate+" GST")
	if err != nil {
		utils.Logf("Failed to parse endDate : %s", err.Error())
		return nil, err
	}
	end = end.Add(time.Hour * 24) // end date has to be 24 hours more even if the start and end date is same

	holidays, err := fetchHolidayDate(startDate, endDate)
	if err != nil {
		utils.Logf("Failed to fetch holiday count : %s", err.Error())
		return nil, err
	}

	diff := calculateTimeDiff(start, end) - holidays
	if diff <= 0 {
		utils.Log("Start data is greater than End date OR queried date is a holiday")
		return nil, fmt.Errorf("start date greater than end date OR queried date is a holiday")
	}

	empList := fetchAllEmployeeFromDb(id, empTyp) //to-do: make query using start and end date

	if len(empList) == 0 {
		return nil, nil
	}

	var (
		totalMeetingCount    int
		totalMeetingTime     float64
		externalMeetingCount int
		externalMeetingTime  float64
	)

	for _, v := range empList {

		startTime, err := time.Parse(layoutTime, v.StartTime)
		if err != nil {
			utils.Logf("Failed to parse startTime : %s", err.Error())
			return nil, err
		}
		endTime, err := time.Parse(layoutTime, v.EndTime)
		if err != nil {
			utils.Logf("Failed to parse endTime : %s", err.Error())
			return nil, err
		}

		if startTime.After(start) && endTime.Before(end) {
			totalMeetingTime += endTime.Sub(startTime).Minutes()
			totalMeetingCount++

			if strings.Contains(v.OrganizerID, "@") {
				externalMeetingTime += endTime.Sub(startTime).Minutes()
				externalMeetingCount++
			}
		}
	}

	empRsp := evaluate(id, float64(diff), totalMeetingTime)

	interMeetingHours, _ := time.ParseDuration(fmt.Sprintf("%f", totalMeetingTime-externalMeetingTime) + "m")

	empRsp.ExternalMeetingHours = time.Duration(externalMeetingTime).Hours()
	empRsp.InternalMeetingHours = interMeetingHours.Hours()
	empRsp.ExternalMeetingCount = externalMeetingCount
	empRsp.InternalMeetingCount = totalMeetingCount - externalMeetingCount
	empRsp.TotalMeetingCount = totalMeetingCount

	return empRsp, nil
}

// FetchCumulativeResult returns the cumulative meeting hours for given unit
func FetchCumulativeResult(unit, year string) (map[string]float64, error) {
	monthlyReportResponse := make(map[string]float64)

	ch = make(chan cumResposne, 12)
	defer close(ch)

	go fetchMonthlyMeetingHours("January", unit, "1/1/"+year, "1/32/"+year)

	yearInt, err := strconv.Atoi(year)
	if err != nil {
		utils.Logf("Faile to convert to int : %s", err.Error())
		return nil, err
	}

	if (yearInt%100 != 0 && yearInt%4 == 0) || yearInt%400 == 0 {
		go fetchMonthlyMeetingHours("February", unit, "2/1/"+year, "2/30/"+year)
	} else {
		go fetchMonthlyMeetingHours("February", unit, "2/1/"+year, "2/29/"+year)
	}

	go fetchMonthlyMeetingHours("March", unit, "3/1/"+year, "3/32/"+year)
	go fetchMonthlyMeetingHours("April", unit, "4/1/"+year, "4/31/"+year)
	go fetchMonthlyMeetingHours("May", unit, "5/1/"+year, "5/32/"+year)
	go fetchMonthlyMeetingHours("June", unit, "6/1/"+year, "6/31/"+year)
	go fetchMonthlyMeetingHours("July", unit, "7/1/"+year, "7/32/"+year)
	go fetchMonthlyMeetingHours("August", unit, "8/1/"+year, "8/32/"+year)
	go fetchMonthlyMeetingHours("September", unit, "9/1/"+year, "9/31/"+year)
	go fetchMonthlyMeetingHours("October", unit, "10/1/"+year, "10/32/"+year)
	go fetchMonthlyMeetingHours("November", unit, "11/1/"+year, "11/31/"+year)
	go fetchMonthlyMeetingHours("December", unit, "12/1/"+year, "12/32/"+year)

	for i := 0; i < 12; i++ {
		select {
		case rslt := <-ch:
			monthlyReportResponse[rslt.month] = rslt.meetingHours
		}
	}

	return monthlyReportResponse, nil
}

// FetchDetailedReport returns the detailed repost for each department manager
func FetchDetailedReport(id, empType, startDate, endDate string) (*DetailedInfoResponse, error) {

	var result = &DetailedInfoResponse{}

	start, err := time.Parse(layoutDate, startDate+" GST") // GST - Gulf Standard Time Zone
	if err != nil {
		utils.Logf("Failed to parse startDate : %s", err.Error())
		return nil, err
	}

	end, err := time.Parse(layoutDate, endDate+" GST")
	if err != nil {
		utils.Logf("Failed to parse endDate : %s", err.Error())
		return nil, err
	}
	end = end.Add(time.Hour * 24)

	info, err := fetchEmpDetailedRpt(id, empType, startDate, endDate)
	if err != nil {
		utils.Logf("Failed to fetch emp detailed report from db : %s", err.Error())
		return nil, err
	}

	for _, v := range info.Info {
		startTime, err := time.Parse(layoutTime, v.StartTime)
		if err != nil {
			utils.Logf("Failed to parse startTime : %s", err.Error())
			return nil, err
		}
		endTime, err := time.Parse(layoutTime, v.EndTime)
		if err != nil {
			utils.Logf("Failed to parse endTime : %s", err.Error())
			return nil, err
		}

		if startTime.After(start) && endTime.Before(end) {
			result.Info = append(result.Info, v)
		}
	}

	result.EmployeeID = id

	return result, nil
}
