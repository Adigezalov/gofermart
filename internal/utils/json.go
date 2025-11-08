package utils

import (
	"encoding/json"
	"net/http"
)

// WriteJSON записывает данные в формате JSON в HTTP response
func WriteJSON(w http.ResponseWriter, data interface{}) error {
	encoder := json.NewEncoder(w)
	return encoder.Encode(data)
}
