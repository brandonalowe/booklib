package api

import "net/http"

func InternalServerErrorResponse(w http.ResponseWriter, err error) {
	writeJson(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
}

func OkResponse(w http.ResponseWriter, v any) {
	writeJson(w, http.StatusOK, v)
}

func NotFoundResponse(w http.ResponseWriter, err string) {
	writeJson(w, http.StatusNotFound, map[string]string{"error": err})
}

func BadRequestResponse(w http.ResponseWriter, err string) {
	writeJson(w, http.StatusBadRequest, map[string]string{"error": err})
}

func CreatedResponse(w http.ResponseWriter, v any) {
	writeJson(w, http.StatusCreated, v)
}
