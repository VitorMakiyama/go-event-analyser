package repository

import (
	"errors"
	"slices"
	"time"
)

type RepositoryFake struct {
	subjects map[int64]Subject
	events   map[int64]Event
}

func NewRepositoryFake() Repository {
	subjects := map[int64]Subject{
		1: {
			ID:          1,
			Name:        "T1",
			Description: "T1-Desc",
		},
		2: {
			ID:          2,
			Name:        "T2",
			Description: "T2-Desc",
		},
	}

	events := map[int64]Event{
		1: {
			ID:          1,
			SubjectID:   1,
			Occurrences: 1,
			InsertTS:    time.Date(2026, time.January, 01, 0, 0, 0, 0, time.Now().Location()),
			LastUpdate:  time.Now(),
		},
		2: {
			ID:          2,
			SubjectID:   2,
			Occurrences: 5,
			InsertTS:    time.Date(2025, time.December, 1, 0, 0, 0, 0, time.Now().Location()),
			LastUpdate:  time.Now(),
		},
		3: {
			ID:          3,
			SubjectID:   2,
			Occurrences: 1,
			InsertTS:    time.Date(2026, time.January, 01, 0, 0, 0, 0, time.Now().Location()),
			LastUpdate:  time.Now(),
		},
		4: {
			ID:          4,
			SubjectID:   2,
			Occurrences: 2,
			InsertTS:    time.Date(2026, time.January, 8, 0, 0, 0, 0, time.Now().Location()),
			LastUpdate:  time.Now(),
		},
		5: {
			ID:          5,
			SubjectID:   2,
			Occurrences: 3,
			InsertTS:    time.Date(2026, time.February, 1, 0, 0, 0, 0, time.Now().Location()),
			LastUpdate:  time.Now(),
		},
	}
	return RepositoryFake{
		subjects: subjects,
		events:   events,
	}
}

// CheckEventExistenceByDate implements [Repository].
func (r RepositoryFake) CheckEventExistenceByDate(insert_ts time.Time, subject_id int64) (foundE Event, err error) {
	if insert_ts.IsZero() {
		return Event{}, errors.New("")
	}

	for _, e := range r.events {
		if e.SubjectID == subject_id && e.InsertTS.Format(time.DateOnly) == insert_ts.Format(time.DateOnly) {
			return e, nil
		}
	}
	return Event{}, errors.New("no rows in result")
}

// DeleteEvent implements [Repository].
func (r RepositoryFake) DeleteEvent(id int64) (int64, error) {
	panic("unimplemented")
}

// DeleteSubject implements [Repository].
func (r RepositoryFake) DeleteSubject(id int64) (int64, error) {
	panic("unimplemented")
}

// GetEvent implements [Repository].
func (r RepositoryFake) GetEvent(id int64) (Event, error) {
	// unknown error
	if id == -1 {
		return Event{}, errors.New("")
	}

	e, ok := r.events[id]
	if !ok {
		return e, ErrorEventIDNotFound{EventID: id}
	}
	return e, nil
}

// GetAllEventsFromSubject implements [Repository].
func (r RepositoryFake) GetAllEventsFromSubject(subject_id int64) ([]Event, error) {
	// unknown error
	if subject_id == -1 {
		return []Event{}, errors.New("")
	}

	var events []Event
	for _, e := range r.events {
		if e.SubjectID == subject_id {
			events = append(events, e)
		}
	}
	slices.SortFunc(events, func(a, b Event) int {
		if a.InsertTS.Before(b.InsertTS) {
			return -1
		} else {
			return 1
		}
	})
	return events, nil
}

// GetSubject implements [Repository].
func (r RepositoryFake) GetSubject(id int64) (Subject, error) {
	s, ok := r.subjects[id]
	if !ok {
		return Subject{}, ErrorSubjectIDNotFound{}
	}
	return s, nil
}

// GetAllSubjects implements [Repository].
func (r RepositoryFake) GetAllSubjects() ([]Subject, error) {
	var subjectsSlice []Subject
	for _, s := range r.subjects {
		subjectsSlice = append(subjectsSlice, s)
	}
	return subjectsSlice, nil
}

// InsertEvent implements [Repository].
func (r RepositoryFake) InsertEvent(e Event) (int64, error) {
	if e.ID == -1 {
		return e.ID, errors.New("")
	}
	id := int64(len(r.events) + 1)
	r.events[id] = e
	return id, nil
}

// InsertSubject implements [Repository].
func (r RepositoryFake) InsertSubject(s Subject) (int64, error) {
	if s.ID < 0 {
		return s.ID, errors.New("")
	}
	id := int64(len(r.subjects) + 1)
	r.subjects[id] = s
	return id, nil
}

// UpdateEvent implements [Repository].
func (r RepositoryFake) UpdateEvent(e Event) (Event, error) {
	if e.SubjectID == -1 {
		return Event{}, errors.New("")
	}

	event, ok := r.events[e.ID]
	if !ok {
		return Event{}, ErrorEventIDNotFound{}
	}
	event.SubjectID = e.SubjectID
	event.Occurrences = e.Occurrences

	r.events[e.ID] = event
	return event, nil
}

// UpdateSubject implements [Repository].
func (r RepositoryFake) UpdateSubject(s Subject) (Subject, error) {
	panic("unimplemented")
}
