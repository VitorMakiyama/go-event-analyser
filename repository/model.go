package repository

import (
	"time"
)

type Subject struct {
	ID          int64
	Name        string
	Description string
}

type Event struct {
	ID         int64
	SubjectID  int64
	Ocurrences int
	InsertTS   time.Time
	LastUpdate time.Time
}

func (e *Event) GetDate() string {
	return e.InsertTS.Format(time.DateOnly)
}

type Repository interface {
	InsertSubject(s Subject) (int64, error)
	GetSubject(id int64) (Subject, error)
	UpdateSubject(s Subject) (Subject, error)
	DeleteSubject(id int64) (int64, error)

	InsertEvent(e Event) (int64, error)
	GetEvent(id int64) (Event, error)
	UpdateEvent(e Event) (Event, error)
	DeleteEvent(id int64) (int64, error)
}
