package services

import "go-event-analyser/internal/repository"

type EventsServiceFake struct {
	CallbackCreate func(e repository.Event) (repository.Event, error)
	CallbackGet    func(id int64) (repository.Event, error)
	CallbackUpdate func(e repository.Event) (repository.Event, error)
	CallbackDelete func(id int64) (int64, error)
}

func (es *EventsServiceFake) Create(e repository.Event) (repository.Event, error) {
	if es.CallbackCreate != nil {
		return es.CallbackCreate(e)
	}
	return e, nil
}

func (es *EventsServiceFake) Get(id int64) (repository.Event, error) {
	if es.CallbackGet != nil {
		return es.CallbackGet(id)
	}
	return repository.Event{ID: id}, nil
}

func (es *EventsServiceFake) Update(e repository.Event) (repository.Event, error) {
	if es.CallbackUpdate != nil {
		return es.CallbackUpdate(e)
	}
	return e, nil
}

func (es *EventsServiceFake) Delete(id int64) (int64, error) {
	if es.CallbackDelete != nil {
		return es.CallbackDelete(id)
	}
	return id, nil
}

// Subject

type SubjectsServiceFake struct {
	CallbackCreate func(s repository.Subject) (repository.Subject, error)
	CallbackGet    func(id int64) (repository.Subject, error)
	CallbackGetAll func() []repository.Subject
	CallbackUpdate func(s repository.Subject) (repository.Subject, error)
	CallbackDelete func(id int64) (int64, error)
}

func (ss *SubjectsServiceFake) Create(newSubject repository.Subject) (repository.Subject, error) {
	if ss.CallbackCreate != nil {
		return ss.CallbackCreate(newSubject)
	}
	return newSubject, nil
}
func (ss *SubjectsServiceFake) Get(id int64) (repository.Subject, error) {
	if ss.CallbackGet != nil {
		return ss.CallbackGet(id)
	}
	return repository.Subject{ID: id}, nil
}
func (ss *SubjectsServiceFake) GetAll() []repository.Subject {
	if ss.CallbackGetAll != nil {
		return ss.CallbackGetAll()
	}
	return []repository.Subject{{}, {}}
}
func (ss *SubjectsServiceFake) Update(s repository.Subject) (repository.Subject, error) {
	panic("unimplemented")
}
func (ss *SubjectsServiceFake) Delete(id int64) (int64, error) {
	panic("unimplemented")
}
