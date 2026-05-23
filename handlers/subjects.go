package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"go-event-analyser/internal/repository"
	"go-event-analyser/internal/services"
	"log"
	"net/http"
	"strconv"
)

type SubjectsHandler struct {
	service services.SubjectsServiceBase
}

func NewSubjectsHandler(service services.SubjectsServiceBase) SubjectsHandler {
	return SubjectsHandler{
		service: service,
	}
}

type CreateSubjectRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type SubjectResponse struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func createResponse(s repository.Subject) SubjectResponse {
	return SubjectResponse{
		ID:          s.ID,
		Name:        s.Name,
		Description: s.Description,
	}
}

func createSliceResponse(subjects []repository.Subject) []SubjectResponse {
	var subjectsReponse []SubjectResponse
	for _, s := range subjects {
		r := SubjectResponse{
			ID:          s.ID,
			Name:        s.Name,
			Description: s.Description,
		}
		subjectsReponse = append(subjectsReponse, r)
	}
	return subjectsReponse
}

func (s *SubjectsHandler) CreateSubject(w http.ResponseWriter, r *http.Request) {
	request := CreateSubjectRequest{}
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		fmt.Println("CreateSubject - error decoding request body: ", err, " Body:", r.Body)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	log.Printf("CreateSubject - Received request body: %v\n", request)

	newSubject := repository.Subject{
		Name:        request.Name,
		Description: request.Description,
	}
	newSubject, err = s.service.Create(newSubject)
	if err != nil {
		fmt.Println("CreateSubject - error inserting subject: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	response := createResponse(newSubject)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (s *SubjectsHandler) GetSubject(w http.ResponseWriter, r *http.Request) {
	queryID := r.URL.Query().Get("id")
	if queryID == "" {
		// If no id was sent, return a slice with all subjects
		subjects := s.service.GetAll()
		response := createSliceResponse(subjects)
		json.NewEncoder(w).Encode(response)
		return
	}

	id, err := strconv.ParseInt(queryID, 10, 64)
	if err != nil {
		fmt.Println("GetSubject - error getting query params: ", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	log.Printf("GetSubject - Received request for id: %d\n", id)

	subject, err := s.service.Get(id)
	if err != nil {
		if errors.As(err, &repository.ErrorSubjectIDNotFound{}) {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		fmt.Println("GetSubject - error getting subject: ", err)
		return
	}

	response := createResponse(subject)

	json.NewEncoder(w).Encode(response)
}
