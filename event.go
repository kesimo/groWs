package groWs

import (
	"encoding/json"
	"log"
)

type Event struct {
	// Event identifier
	Identifier string `json:"event"`
	Data       any    `json:"data"`
}

func IsJSON(data []byte) bool {
	if data[0] == '{' && data[len(data)-1] == '}' {
		return true
	}
	return false
}

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
func eventFromJSON(data []byte) (Event, error) {
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
