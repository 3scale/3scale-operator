package mocks

import (
	"encoding/json"
	"net/http"
)

func errorResponse(w http.ResponseWriter, key string, value []string, status int) {
	errObj := struct {
		Errors map[string][]string `json:"errors"`
	}{
		Errors: map[string][]string{key: value},
	}

	data, err := json.Marshal(errObj)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	http.Error(w, string(data), status)
}
