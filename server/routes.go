package server

import (
	"net/http"
	"os"
	"project/reportingApplication/utils"
	"time"

	"github.com/dgrijalva/jwt-go"

	"github.com/gorilla/context"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

const servicePort = "4040"

// Run - start the service console app
func Run() {
	r := mux.NewRouter()
	r = r.PathPrefix("/reportingApp").Subrouter()

	r.HandleFunc("/signin", signin).Methods(http.MethodPost)

	r.Use(jwtVerify)

	r.HandleFunc("/employeeReport", employeeReport).Methods(http.MethodPost) //to-do: should not panic for API cal
	r.HandleFunc("/cumulativeReport", cumulativeReport).Methods(http.MethodPost)
	r.HandleFunc("/detailedReport", detailedReport).Methods(http.MethodPost)

	r.HandleFunc("/health", getHealth).Methods(http.MethodGet)

	start(r)
}

func jwtVerify(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		utils.Log(r.RemoteAddr, r.RequestURI)

		notAuth := []string{"/reportingApp/signin"}

		requestPath := r.URL.Path

		for _, val := range notAuth {
			if val == requestPath {
				next.ServeHTTP(w, r)
				return
			}
		}

		// header := r.Header.Get("reportingApp-access-token")
		// header = strings.TrimSpace(header)

		ck, err := r.Cookie("reportingApp-access-token")
		if err != nil {
			// if err == http.ErrNoCookie {
			// 	w.WriteHeader(http.StatusUnauthorized)
			// }

			returnResponse(w, http.StatusForbidden, &struct {
				Msg string `json:"message"`
			}{"Missing auth token"})
			utils.Log("Missing auth token, err :", err.Error())
			return
		}

		// if header == "" {
		// 	returnResponse(w, http.StatusForbidden, &struct {
		// 		Msg string `json:"message"`
		// 	}{"Missing auth token"})
		// 	return
		// }

		header := ck.Value

		tk := &Token{}

		secretKey := os.Getenv("ReportingAppSignKey")
		token, err := jwt.ParseWithClaims(header, tk, func(token *jwt.Token) (interface{}, error) {
			return []byte(secretKey), nil
		})

		if err != nil {
			returnResponse(w, http.StatusForbidden, &struct {
				Msg string `json:"message"`
			}{"Malformed authentication token : " + err.Error()})
			utils.Log("Malformed authentication token, err :", err.Error())
			return
		}

		if !token.Valid {
			returnResponse(w, http.StatusForbidden, &struct {
				Msg string `json:"message"`
			}{"Token is not valid"})
			utils.Log("Token is not valid, err :", err.Error())
			return
		}

		// ct := ctx.WithValue(r.Context(), "reportingApp-access-token", tk.Id)
		// r = r.WithContext(ct)
		next.ServeHTTP(w, r)
	})
}

// start initializes and starts http service
func start(r *mux.Router) {
	allowedOrigins := handlers.AllowedOrigins([]string{"*"}) // Allowing all origin as of now

	allowedHeaders := handlers.AllowedHeaders([]string{
		"X-Requested-With",
		"X-CSRF-Token",
		"X-Auth-Token",
		"Content-Type",
		"processData",
		"contentType",
		"Origin",
		"Authorization",
		"Accept",
		"Client-Security-Token",
		"Accept-Encoding",
		"timezone",
		"locale",
	})

	allowedMethods := handlers.AllowedMethods([]string{
		"POST",
		"GET",
		"DELETE",
		"PUT",
		"PATCH",
		"OPTIONS"})

	allowCredential := handlers.AllowCredentials()

	s := &http.Server{
		Addr: ":" + servicePort,
		//	Handler: r,
		Handler: handlers.CORS(
			allowedHeaders,
			allowedMethods,
			allowedOrigins,
			allowCredential)(
			context.ClearHandler(
				r,
			)),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		// MaxHeaderBytes: 1 << 20,
	}

	utils.Fatalln(s.ListenAndServe())
}
