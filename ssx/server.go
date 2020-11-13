package ssx

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

const (
	jsonContentType = "application/json"
)

type (
	// For decoding Reference ID from HTTP request.
	activityHandlerPayloadScenarioInputsValue struct {
		CUSTOMER          string `json:"CUSTOMER"`
		SRAEnvironment    string `json:"SRAEnvironment"`
		DomainNumber      string `json:"domainNumber"`
		OSName            string `json:"OSName"`
		HostEnvironment   string `json:"HostEnvironment"`
		SiteID            string `json:"SiteID"`
		CPUcount          string `json:"CPUcount"`
		MemoryGB          string `json:"MemoryGB"`
		CISupportTier     string `json:"CISupportTier"`
		CIImpact          string `json:"CIImpact"`
		Function          string `json:"Function"`
		ServerCount       string `json:"serverCount"`
		BackupType        string `json:"BackupType"`
		Disk0SizeGB       string `json:"disk0SizeGB"`
		Disk1SizeGB       string `json:"disk1SizeGB"`
		DisksAllowed      string `json:"disksAllowed"`
		OutageDay         string `json:"OutageDay"`
		OutageWindow      string `json:"OutageWindow"`
		ManagedBy         string `json:"managedBy"`
		OfferingName      string `json:"offeringName"`
		SubscriptionEmail string `json:"subscriptionEmail"`
		SubscriptionName  string `json:"subscriptionName"`
		RequestorUser     string `json:"requestorUser"`
		RequestorEmail    string `json:"requestorEmail"`
		RequestorName     string `json:"requestorName"`
		RequestorGroup    string `json:"requestorGroup"`
		Reference         string `json:"Reference"`
		SubscriptionID    string `json:"SubscriptionID"`
		ServiceRequestID  string `json:"serviceRequestId"`
	}

	// For Decoding scenarioId and triggerName from HTTP Request.
	ActivityHandlerPayload struct {
		ScenarioID          int                                       `json:"scenarioId"`
		ScenarioInputsValue activityHandlerPayloadScenarioInputsValue `json:"scenarioInputsValue"`
		TriggerName         string                                    `json:"triggerName"`
	}

	// Handler with router initialization.
	Server struct {
		store storageInterface
		http.Handler
	}
)

// Constructor to create Server handler.
func NewServer(store storageInterface) *Server {
	ssx := new(Server)
	ssx.store = store

	router := mux.NewRouter()
	router.Handle("/rest/v0/activities", http.HandlerFunc(ssx.activityHandler)).Methods(http.MethodPost)
	router.Handle("/rest/v0/activities/{id}", http.HandlerFunc(ssx.statusHandler)).Methods(http.MethodGet)

	ssx.Handler = router

	return ssx
}

// HandlerFunction for POST requests.
func (s *Server) activityHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", jsonContentType)
	w.WriteHeader(http.StatusCreated)

	payload := ActivityHandlerPayload{}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		log.Printf("Unable to parse response from server %v, '%v'", r.Body, err)
	}

	if err := json.NewEncoder(w).Encode(s.store.setData(&payload)); err != nil {
		log.Printf("Unable to encode return payload from server '%v'", err)
	}
}

// HandleFunction for GET requests.
func (s *Server) statusHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", jsonContentType)

	vars := mux.Vars(r)

	data, err := s.store.getData(vars["id"])
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
	}

	if err := json.NewEncoder(w).Encode(data); err != nil {
		fmt.Println(err)
	}
}
