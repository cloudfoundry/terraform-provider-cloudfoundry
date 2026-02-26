package main

import (
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
)

// Service Broker API structures
type Catalog struct {
	Services []Service `json:"services"`
}

type Service struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Description     string `json:"description"`
	Bindable        bool   `json:"bindable"`
	PlanUpdateable  bool   `json:"plan_updateable"`
	Plans           []Plan `json:"plans"`
}

type Plan struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Free        bool   `json:"free"`
}

type ProvisionRequest struct {
	ServiceID        string                 `json:"service_id"`
	PlanID           string                 `json:"plan_id"`
	OrganizationGUID string                 `json:"organization_guid"`
	SpaceGUID        string                 `json:"space_guid"`
	Parameters       map[string]interface{} `json:"parameters"`
}

type ProvisionResponse struct {
	DashboardURL string `json:"dashboard_url,omitempty"`
	Operation    string `json:"operation,omitempty"`
}

type BindRequest struct {
	ServiceID  string                 `json:"service_id"`
	PlanID     string                 `json:"plan_id"`
	AppGUID    string                 `json:"app_guid"`
	Parameters map[string]interface{} `json:"parameters"`
}

type BindResponse struct {
	Credentials map[string]interface{} `json:"credentials"`
	SyslogDrainURL string `json:"syslog_drain_url,omitempty"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func basicAuth(r *http.Request) bool {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return false
	}

	parts := strings.SplitN(auth, " ", 2)
	if len(parts) != 2 || parts[0] != "Basic" {
		return false
	}

	credentials, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return false
	}

	userPass := strings.SplitN(string(credentials), ":", 2)
	if len(userPass) != 2 {
		return false
	}

	username := os.Getenv("BROKER_USERNAME")
	if username == "" {
		username = "admin"
	}

	password := os.Getenv("BROKER_PASSWORD")
	if password == "" {
		password = "password"
	}

	return userPass[0] == username && userPass[1] == password
}

func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !basicAuth(r) {
			w.Header().Set("WWW-Authenticate", `Basic realm="Service Broker"`)
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "Unauthorized"})
			return
		}
		next(w, r)
	}
}

func catalogHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	catalog := Catalog{
		Services: []Service{
			{
				ID:              "simple-service-id",
				Name:            "simple-service",
				Description:     "A simple test service",
				Bindable:        true,
				PlanUpdateable:  false,
				Plans: []Plan{
					{
						ID:          "simple-plan-id",
						Name:        "simple",
						Description: "Simple plan",
						Free:        true,
					},
					{
						ID:          "easy-plan-id",
						Name:        "easy",
						Description: "Easy plan",
						Free:        true,
					},
				},
			},
		},
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(catalog)
}

func provisionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	response := ProvisionResponse{
		DashboardURL: "http://example.com/dashboard",
		Operation:    "created",
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func deprovisionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	response := map[string]string{
		"operation": "deleted",
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func bindHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	credentials := map[string]interface{}{
		"username": "service_user",
		"password": "service_password_123",
		"host":     "service.example.com",
		"port":     5432,
		"database": "service_db",
		"uri":      "postgresql://service_user:service_password_123@service.example.com:5432/service_db",
	}

	response := BindResponse{
		Credentials: credentials,
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func unbindHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	response := map[string]string{
		"operation": "deleted",
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	response := map[string]string{"status": "ok"}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Public endpoints
	http.HandleFunc("/health", healthHandler)

	// Protected endpoints (require basic auth)
	http.HandleFunc("/v2/catalog", authMiddleware(catalogHandler))
	http.HandleFunc("/v2/service_instances/", func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/service_bindings/") {
			if r.Method == http.MethodPut {
				authMiddleware(bindHandler)(w, r)
			} else if r.Method == http.MethodDelete {
				authMiddleware(unbindHandler)(w, r)
			}
		} else {
			if r.Method == http.MethodPut {
				authMiddleware(provisionHandler)(w, r)
			} else if r.Method == http.MethodDelete {
				authMiddleware(deprovisionHandler)(w, r)
			}
		}
	})

	log.Printf("Simple Service Broker listening on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
