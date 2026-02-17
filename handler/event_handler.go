package handler

import (
	"encoding/json"
	"fmt"
	"go-event-analyser/repository"
	"net/http"
	"strconv"
)

var repo repository.Repository = repository.NewPostgreSQLRepository()

func CreateEvent(w http.ResponseWriter, r *http.Request) {
	panic("not implemented")
}

func GetEvent(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		fmt.Println("error getting query params: ", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	event, err := repo.GetEvent(id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(event)
}
