package ssx

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"reflect"
	"strconv"
	"sync"
	"time"
)

type (
	// Interface to help with Mock for Unit testing.
	storageInterface interface {
		setData(payload *ActivityHandlerPayload) *storage
		getData(id string) (*storage, error)
	}

	// For storing data.
	// This is setup with minimum parameters required.
	storage struct {
		ID               string `json:"id"`
		Status           string `json:"status"`
		ResultStatusName string `json:"resultStatusName"`
		ResultStatusType string `json:"resultStatusType"`
		Inputs           []*ioData
		Outputs          []*ioData
	}

	// Aggregation relationship with 'storage' struct.
	ioData struct {
		Label      string `json:"label"`
		EntryIndex int    `json:"entryIndex"`
		Value      string `json:"value"`
	}

	// This implement two storages.
	// One for storing the Job details(store).
	// Another one for storing the state of the data(data).
	Store struct {
		store map[string]*storage
		data  map[string]*storage
		m     *sync.Mutex
	}
)

// Constructor to create a new Store.
func NewStore() *Store {
	return &Store{map[string]*storage{}, map[string]*storage{}, &sync.Mutex{}}
}

// Constructor to create a new storage with default data.
func newDefaultStorage(id string) *storage {
	return &storage{
		ID:               id,
		Status:           "PENDING",
		ResultStatusName: "",
		ResultStatusType: "",
		Inputs:           []*ioData{},
		Outputs:          []*ioData{},
	}
}

// Function to duplicate storage.
func copyStorage(s *storage) *storage {
	sCopy := *s

	return &sCopy
}

// Function to create ioData.
func newIOData(key, value string, index int) *ioData {
	return &ioData{
		Label:      key,
		EntryIndex: index,
		Value:      value,
	}
}

// Function to create return payload based on scenario ID.
func (j *Store) setData(payload *ActivityHandlerPayload) *storage {
	id, refID := j.initStore(payload)

	createID := 1002
	deleteID := 1003
	readID := 1004
	updateID := 1005

	switch payload.ScenarioID {
	case createID:
		return j.create(id, refID, &payload.ScenarioInputsValue)
	case deleteID:
		return j.delete(id, refID)
	case readID:
		return j.read(id, refID)
	case updateID:
		return j.update(id, refID, &payload.ScenarioInputsValue)
	default:
		return &storage{}
	}
}

// Function to Initialize Store.
func (j *Store) initStore(payload *ActivityHandlerPayload) (string, string) {
	refID := payload.ScenarioInputsValue.Reference

	id := generateID()
	jstorage := newDefaultStorage(id)
	dstorage := copyStorage(jstorage)

	// Create Job Store space only if the same job id doesn't exist.
	if _, ok := j.store[id]; !ok {
		j.store[id] = jstorage
	}

	// Create Data store space only if the same Reference id doesn't exist.
	if _, ok := j.data[refID]; !ok {
		j.data[refID] = dstorage
	}

	return id, refID
}

// Function to simulate resource create/Deploy.
func (j *Store) create(id, refID string, hwspec *activityHandlerPayloadScenarioInputsValue) *storage {
	go createHelper(j.store[id], hwspec, 30000, &sync.Mutex{})
	go createHelper(j.data[refID], hwspec, 30000, &sync.Mutex{})
	return j.store[id]
}

// Create helper function to autoupdate the Store after a specific duration.
func createHelper(store *storage, hwspec *activityHandlerPayloadScenarioInputsValue, delay time.Duration, m *sync.Mutex) {
	time.Sleep(delay * time.Millisecond)

	hwJSON, err := json.Marshal(hwspec)
	if err != nil {
		fmt.Println(err)
	}
	output := newIOData("JSON", string(hwJSON), 1)
	m.Lock()
	store.Outputs = []*ioData{output}
	storeDone(store)
	m.Unlock()
}

// Function to simulate resource read.
func (j *Store) read(id, refID string) *storage {
	j.data[refID].ID = id

	go readHelper(j.store[id], j.data[refID], 30000)

	return j.store[id]
}

// Read helper function to autoupdate the Store after a specific duration.
func readHelper(store, data *storage, delay time.Duration) {
	time.Sleep(delay * time.Millisecond)

	tempStore := *data
	*store = tempStore

	storeDone(store)
}

// Function to simulate resource update.
func (j *Store) update(id, refID string, hwspec *activityHandlerPayloadScenarioInputsValue) *storage {
	j.data[refID].ID = id
	go updateHelper(j.store[id], j.data[refID], hwspec, 30000, &sync.Mutex{})
	return j.store[id]
}

// Read helper function to autoupdate the Store after a specific duration.
func updateHelper(store, data *storage, hwspec *activityHandlerPayloadScenarioInputsValue, delay time.Duration, m *sync.Mutex) {
	time.Sleep(delay * time.Millisecond)

	oldInput := stringToStruct(data.Outputs[0].Value)

	// reflect.Value.
	// ValueOf returns a new Value initialized to the concrete value stored in struct.
	// ValueOf(nil) returns the zero Value.
	sourcePointer := reflect.ValueOf(hwspec)
	destinationPointer := reflect.ValueOf(oldInput)

	// Indirect returns the value that sourcePointer points to.
	// If sourcePointer is a nil pointer, Indirect returns a zero Value.
	// If sourcePointer is not a pointer, Indirect returns sourcePointer.
	sourceStruct := reflect.Indirect(sourcePointer)

	// Elem returns the value that the sourceValue contains or that the pointer sourceValue points to.
	// It panics if v's Kind is not Interface or Ptr. It returns the zero Value if sourceValue is nil.
	sourceFields := sourcePointer.Elem()
	destinationFields := destinationPointer.Elem()
	numberOfFields := sourceFields.NumField()

	for i := 0; i < numberOfFields; i++ {
		// Field returns the i'th field of the struct.
		sourceValue := sourceFields.Field(i)
		if !sourceValue.IsZero() {
			sourceKey := sourceStruct.Type().Field(i).Name

			destinationValue := destinationFields.FieldByName(sourceKey)
			if destinationValue.CanSet() {
				destinationValue.Set(sourceValue)
			}
		}
	}

	newInput := structToString(oldInput)

	output := newIOData("JSON", newInput, 1)
	m.Lock()
	store.Outputs = []*ioData{output}
	data.Outputs = []*ioData{output}

	storeDone(store)
	m.Unlock()
}

// Function returns the JSON encoding of struct 'activityHandlerPayloadScenarioInputsValue'.
func structToString(payload *activityHandlerPayloadScenarioInputsValue) (encodedData string) {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		fmt.Println(err)
	}

	encodedData = string(jsonData)

	return
}

// Function parses the JSON-encoded data and return struct 'activityHandlerPayloadScenarioInputsValue'.
func stringToStruct(s string) *activityHandlerPayloadScenarioInputsValue {
	decodedData := &activityHandlerPayloadScenarioInputsValue{}

	err := json.Unmarshal([]byte(s), decodedData)
	if err != nil {
		fmt.Println(err)
	}

	return decodedData
}

// Function to delete a record from Main Store.
// Return the resulting Job store.
func (j *Store) delete(id, refID string) *storage {
	j.m.Lock()
	delete(j.data, refID)
	j.m.Unlock()

	go deleteHelper(j.store[id], 30000, &sync.Mutex{})

	return j.store[id]
}

// Helper function for delete function.
// Actions in this function only takeplace after certain duration.
// Duration need to pass to the function as last parameter.
func deleteHelper(store *storage, delay time.Duration, m *sync.Mutex) {
	time.Sleep(delay * time.Millisecond)
	m.Lock()
	storeDone(store)
	m.Unlock()
}

// This function return status of a certain job which triggered earlier.
// Job ID required to pass to the function as an argument.
// This function will look in to the 'Store' and return the content.
// Return an error incase the 'job ID' not exist.
func (j *Store) getData(id string) (*storage, error) {
	store, ok := j.store[id]
	if ok {
		return store, nil
	}

	return &storage{}, errors.New("Requested ID not exist")
}

// Update the store fields with 'done'.
// This is required to make sure the 'CRUD' return values are consistent across the calls.
func storeDone(store *storage) {
	store.Status = "COMPLETED"
	store.ResultStatusName = "success"
	store.ResultStatusType = "RESOLVED"
}

// To generate a Job ID which is an integer ranging between '1000' and '1999'.
// The integer value will be convert to datatype 'string' before it return.
// Package strconv implements conversions from Integer to ASCII.
func generateID() string {
	idMin := 1000
	idMax := 1999

	return strconv.Itoa(generateRandomID(idMin, idMax))
}

// To generate random ID.
// This is not suitable for security-sensitive works.
// Seed uses the provided seed value to initialize the default Source to a deterministic state.
// The resulting random ID will be a random integer between 'min' and 'max' values.
func generateRandomID(min, max int) int {
	rand.Seed(time.Now().UTC().UnixNano())

	return rand.Intn(max-min) + min
}
