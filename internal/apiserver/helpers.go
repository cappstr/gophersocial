package apiserver

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Validator interface {
	Validate() error
}

type ApiResponse[T any] struct {
	Data    *T     `json:"data,omitempty"`
	Message string `json:"message,omitempty"`
}

func Encode[T any](v T, w http.ResponseWriter, status int) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		return fmt.Errorf("error encoding json: %w", err)
	}
	return nil
}

func Decode[T Validator](r *http.Request) (T, error) {
	var v T
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		return v, fmt.Errorf("error decoding json: %w", err)
	}
	if err := v.Validate(); err != nil {
		return v, fmt.Errorf("error validating json: %w", err)
	}
	return v, nil
}
