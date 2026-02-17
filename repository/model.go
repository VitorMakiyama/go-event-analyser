package repository

import (
	"time"
)

type Subject struct {
	ID          int
	Name        string
	Description string
}

type Event struct {
	ID         int
	SubjectID  int
	Ocurrences int
	InsertTS   time.Time
	LastUpdate time.Time
}

func (e *Event) GetDate() string {
	return e.InsertTS.Format(time.DateOnly)
}

type Repository interface {
	InsertSubject(s Subject) (int, error)
	GetSubject(id int) (Subject, error)
	UpdateSubject(s Subject) (int64, error)
	DeleteSubject(id int) (int64, error)

	InsertEvent(e Event) (int, error)
	GetEvent(id int) (Event, error)
	UpdateEvent(e Event) (int64, error)
	DeleteEvent(id int) (int64, error)
}
