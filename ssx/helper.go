package ssx

import (
	"encoding/json"
	"fmt"
)

// Constructor for creating payload for trigger a criteria.
// This is helpful for consumer to create a payload at first.
func CreateActivityPayload(scenarioID int, input map[string]string, triggername string) *ActivityHandlerPayload {
	jsonInput, err := json.Marshal(input)
	if err != nil {
		fmt.Println(err)
	}

	inputStruct := activityHandlerPayloadScenarioInputsValue{}

	if err := json.Unmarshal(jsonInput, &inputStruct); err != nil {
		fmt.Println(err)
	}

	payload := ActivityHandlerPayload{
		ScenarioID:          scenarioID,
		ScenarioInputsValue: inputStruct,
		TriggerName:         triggername,
	}

	return &payload
}
