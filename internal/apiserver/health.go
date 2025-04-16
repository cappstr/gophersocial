package apiserver

import (
	"net/http"
)

func (s *ApiServer) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}
