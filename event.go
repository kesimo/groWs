package groWs

import "encoding/json"

type Event struct {
	// Event identifier
	Identifier string `json:"event"`
	Data       []byte `json:"data"`
}

func isJSON(data []byte) bool {
	if data[0] == '{' && data[len(data)-1] == '}' {
		return true
	}
	return false
}

func isEvent(data []byte) bool {
	var e Event
	err := json.Unmarshal(data, &e)
	if err != nil {
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
