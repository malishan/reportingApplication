package server

import (
	"encoding/json"
	"net/http"
	"os"
	"project/reportingApplication/operation"
	"time"

	"project/reportingApplication/utils"

	jwt "github.com/dgrijalva/jwt-go"
)

// health struct holds response for service health
type health struct {
	Msg string `json:"msg"`
}

// empInfo request struct for particular employee
type empInfo struct {
	EmployeeID string `json:"employeeID"`
	Type       string `json:"type"`
	StartDate  string `json:"startDate"`
	EndDate    string `json:"endDate"`
}

// monthlyReport holds request for monthly report
type monthlyReport struct {
	Unit string `json:"unit"`
	Year string `json:"year"`
}

// Token forms the structure to validate user
type Token struct {
	Name string
	*jwt.StandardClaims
}

// signin validates user login
func signin(w http.ResponseWriter, r *http.Request) {
	var reqLogin operation.Login
	var resp map[string]interface{}

	if err := json.NewDecoder(r.Body).Decode(&reqLogin); err != nil {
		utils.Log(r.RemoteAddr, "Employee Login - invalid payload received :", err.Error())
		s := &struct {
			Msg string `json:"message"`
		}{"Invalid Login payload : " + err.Error()}
		returnResponse(w, http.StatusBadRequest, s)
		return
	}

	var (
		filter    string
		baseDN    = os.Getenv("BaseDN")        //"dc=AppDomain,dc=local"
		adminUser = os.Getenv("userAdmin")     //"uid=admin,ou=system"
		adminPswd = os.Getenv("adminPassword") //"secret"
		ldapHost  = os.Getenv("LdapHost")      //"'localhost:10389"
	)

	if reqLogin.Username != "" {
		filter = "sAMAccountName"
	} else if reqLogin.Email != "" {
		filter = "mail"
	} else {
		s := &struct {
			Msg string `json:"message"`
		}{"ERROR : Neither UserName Nor Mail Found"}
		returnResponse(w, http.StatusBadRequest, s)
		return
	}

	client, err := operation.NewLDAPClient(operation.Config{
		BaseDN: baseDN,
		Filter: filter,
		ROUser: operation.User{Name: adminUser, Pswd: adminPswd},
		Host:   ldapHost,
	})

	if err != nil {
		msg := &struct {
			Msg string `json:"message"`
		}{"Failed to create LDAP Client"}
		returnResponse(w, http.StatusInternalServerError, msg)
		return
	}

	err = client.Authenticate(reqLogin)
	if err != nil {
		msg := &struct {
			Msg string `json:"message"`
		}{"Invalid login credentials. Please try again"}
		returnResponse(w, http.StatusBadRequest, msg)
		return
	}

	expirationTime := time.Now().Add(30 * time.Minute)

	tk := &Token{
		Name: reqLogin.Username,
		StandardClaims: &jwt.StandardClaims{
			IssuedAt:  time.Now().Unix(),
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.GetSigningMethod("HS256"), tk)

	secretKey := os.Getenv("ReportingAppSignKey")
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		msg := &struct {
			Msg string `json:"message"`
		}{"Internal Server Error"}
		returnResponse(w, http.StatusBadRequest, msg)
	}

	http.SetCookie(w, &http.Cookie{
		Name:    "reportingApp-access-token",
		Value:   tokenString,
		Expires: expirationTime,
	})

	reqLogin.Password = ""

	resp = map[string]interface{}{"status": false, "message": "logged in", "token": tk, "user": reqLogin}

	returnResponse(w, http.StatusOK, resp)
}

// employeeReport requets for all info of a particular employee
func employeeReport(w http.ResponseWriter, r *http.Request) {
	var reqEmpInfo empInfo

	if err := json.NewDecoder(r.Body).Decode(&reqEmpInfo); err != nil {
		utils.Log("Employee details - invalid payload received :", err.Error())
		s := &struct {
			Msg string `json:"message"`
		}{"Invalid payload received"}
		returnResponse(w, http.StatusBadRequest, s)
		return
	}

	response, err := operation.FetchResult(reqEmpInfo.EmployeeID, reqEmpInfo.StartDate, reqEmpInfo.EndDate, reqEmpInfo.Type)
	if err != nil {
		msg := &struct {
			Msg string `json:"message"`
		}{err.Error()}
		returnResponse(w, http.StatusInternalServerError, msg)
		return
	}

	if response == nil {
		msg := &struct {
			Msg string `json:"message"`
		}{"ID not found"}
		returnResponse(w, http.StatusBadRequest, msg)
		return
	}

	returnResponse(w, http.StatusOK, response)
}

// cumulativeReport returns the monthly report of specified units
func cumulativeReport(w http.ResponseWriter, r *http.Request) {
	var reqInfo monthlyReport

	if err := json.NewDecoder(r.Body).Decode(&reqInfo); err != nil {
		utils.Log("Unit details - invalid payload received :", err.Error())
		s := &struct {
			Msg string `json:"message"`
		}{"Invalid payload received : " + err.Error()}
		returnResponse(w, http.StatusBadRequest, s)
		return
	}

	response, err := operation.FetchCumulativeResult(reqInfo.Unit, reqInfo.Year)
	if err != nil {
		msg := &struct {
			Msg string `json:"message"`
		}{err.Error()}
		returnResponse(w, http.StatusInternalServerError, msg)
		return
	}
	returnResponse(w, http.StatusOK, response)
}

// detailedReport returns the detailed info of a particular employee
func detailedReport(w http.ResponseWriter, r *http.Request) {
	var reqEmpInfo empInfo
	if err := json.NewDecoder(r.Body).Decode(&reqEmpInfo); err != nil {
		utils.Log("Employee details - invalid payload received :", err.Error())
		s := &struct {
			Msg string `json:"message"`
		}{"Invalid payload received : " + err.Error()}
		returnResponse(w, http.StatusBadRequest, s)
		return
	}

	response, err := operation.FetchDetailedReport(reqEmpInfo.EmployeeID, reqEmpInfo.Type, reqEmpInfo.StartDate, reqEmpInfo.EndDate)
	if err != nil {
		msg := &struct {
			Msg string `json:"message"`
		}{err.Error()}
		returnResponse(w, http.StatusInternalServerError, msg)
		return
	}

	if response == nil {
		msg := &struct {
			Msg string `json:"message"`
		}{"ID not found"}
		returnResponse(w, http.StatusBadRequest, msg)
		return
	}

	returnResponse(w, http.StatusOK, response)
}

// getHealth checks the health status of the service
func getHealth(w http.ResponseWriter, r *http.Request) {
	utils.Log("Health check ok")

	h := health{"Service is up"}

	err := json.NewEncoder(w).Encode(&h)
	if err != nil {
		utils.Log("error in encoding response: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// returnResponse handles the response type of any API
func returnResponse(w http.ResponseWriter, statusCode int, status interface{}) {
	respb, _ := json.Marshal(status)
	w.WriteHeader(statusCode)
	w.Header().Set("Content-Type", "application/json")
	w.Write(respb)
}
