package handler

import (
	"encoding/json"
	"fmt"
	"go-event-analyser/repository"
	"net/http"
	"strconv"
	"strings"
)

var repo repository.Repository = repository.NewPostgreSQLRepository()

type CreateSubjectRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type SubjectResponse struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func CreateResponse(s repository.Subject) repository.Subject {
	return repository.Subject{
		ID:          s.ID,
		Name:        s.Name,
		Description: s.Description,
	}
}

func CreateSubject(w http.ResponseWriter, r *http.Request) {
	request := CreateSubjectRequest{}
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		fmt.Println("error decoding request body: ", err, " Body:", r.Body)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	newSubject := repository.Subject{
		Name:        request.Name,
		Description: request.Description,
	}
	newSubject.ID, err = repo.InsertSubject(newSubject)
	if err != nil {
		fmt.Println("error inserting subject: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	response := CreateResponse(newSubject)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func GetSubject(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.URL.Query().Get("id"), 10, 64)
	if err != nil {
		fmt.Println("error getting query params: ", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	subject, err := repo.GetSubject(id)
	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		fmt.Println("error getting subject: ", err)
		return
	}

	response := CreateResponse(subject)

	json.NewEncoder(w).Encode(response)
}
