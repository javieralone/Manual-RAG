package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"api-go/internal/adapters/http/response"
)

func decodeJSON(r *http.Request, destination any) error {
	return json.NewDecoder(r.Body).Decode(destination)
}

func writeDecodeError(w http.ResponseWriter, err error) {
	var maxBytesError *http.MaxBytesError
	if errors.As(err, &maxBytesError) {
		response.WriteError(w, http.StatusRequestEntityTooLarge, "cuerpo de solicitud demasiado grande")
		return
	}
	response.WriteError(w, http.StatusBadRequest, "formato JSON inválido")
}