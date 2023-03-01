package groWs

import (
	"encoding/json"
	"log"
)

type Event struct {
	// Event identifier used to identify the event on the client and server side
	// The ClientHandler.OnEvent() method uses this identifier to match the event
	Identifier string `json:"event"`
	// Data is the data that is sent with the event and can be of any type
	// On send and receive the data is converted from JSON to any type
	Data any `json:"data"`
}

// IsJSONObject checks if the data is JSON
// This uses the first and last character of the data to check if it is JSON
// and is not a 100% accurate way to check if the data is JSON (e.g. for an array) but is faster
func IsJSONObject(data []byte) bool {
	if data[0] == '{' && data[len(data)-1] == '}' {
		return true
	}
	return false
}

// IsEvent checks if the data is an Event by trying to unmarshal a JSON to an Event
func IsEvent(data []byte) bool {
	var e Event
	err := json.Unmarshal(data, &e)
	if err != nil {
		log.Println(err)
		return false
	}
	return true
}

// FromJSON converts JSON data to an Event
func FromJSON(data []byte) (Event, error) {
	var e Event
	err := json.Unmarshal(data, &e)
	if err != nil {
		return e, err
	}
	return e, nil
}

// ToJSON converts an Event to JSON data
func (e Event) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}
