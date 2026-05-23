package services

import (
	"go-event-analyser/internal/repository"
	"log"
)

type SubjectsService struct {
	repository repository.Repository
}


type SubjectsServiceBase interface {
	Create(newSubject repository.Subject) (repository.Subject, error)
	Get(id int64) (repository.Subject, error)
	GetAll() []repository.Subject
	Update(s repository.Subject) (repository.Subject, error)
	Delete(id int64) (int64, error)
}

func NewSubjectsService(repository repository.Repository) SubjectsServiceBase {
	return &SubjectsService{
		repository: repository,
	}
}

func (ss *SubjectsService) Create(newSubject repository.Subject) (repository.Subject, error) {
	id, err := ss.repository.InsertSubject(newSubject)
	if err != nil {
		log.Println("SubjectsService - error inserting subject: ", err)
		return newSubject, err
	}
	
	newSubject.ID = id
	return newSubject, nil
}

// Get implements [SubjectsServiceBase].
func (ss *SubjectsService) Get(id int64) (repository.Subject, error) {
	subject, err := ss.repository.GetSubject(id)
	if err != nil {
		log.Println("SubjectsService - error getting subject: ", err)
		return repository.Subject{}, err
	}
	return subject, nil
}

// GetAll implements [SubjectsServiceBase].
func (ss *SubjectsService) GetAll() []repository.Subject {
	subjects, err := ss.repository.GetAllSubjects()
	if err != nil {
		// Unknown error
		log.Println("SubjectsService - error getting all subjects: ", err)
	}
	return subjects
}

// Update implements [SubjectsServiceBase].
func (ss *SubjectsService) Update(s repository.Subject) (repository.Subject, error) {
	panic("unimplemented")
}

// Delete implements [SubjectsServiceBase].
func (ss *SubjectsService) Delete(id int64) (int64, error) {
	panic("unimplemented")
}