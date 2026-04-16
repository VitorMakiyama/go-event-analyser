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
