package repository

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type PostgreSQLRepository struct {
	db *sql.DB
}

func NewPostgreSQLRepository() Repository {
	db, err := OpenConn()
	if err != nil {
		panic(fmt.Sprintf("could not connect to PostreSQL: %v", err))
	}

	return PostgreSQLRepository{
		db: db,
	}
}

type Settings struct {
	DBHost string
	DBPort int
	DBUser string
	DBPass string
	DBName string
}

func NewSettings() Settings {
	settings := Settings{}

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	dbPort, err := strconv.Atoi(os.Getenv("DB_PORT"))
	if err != nil {
		panic(fmt.Sprintf("could not find PostgreSQL port: %v", err))
	}

	settings.DBHost = os.Getenv("DB_HOST")
	settings.DBPort = dbPort
	settings.DBUser = os.Getenv("POSTGRES_USER")
	settings.DBPass = os.Getenv("POSTGRES_PASSWORD")
	settings.DBName = os.Getenv("POSTGRES_DB")
	return settings
}

func OpenConn() (*sql.DB, error) {
	settings := NewSettings()
	connString := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", settings.DBHost, settings.DBPort, settings.DBUser, settings.DBPass, settings.DBName)
	db, err := sql.Open("postgres", connString)
	if err != nil {
		panic(err)
	}

	err = db.Ping()

	return db, err
}

func (pr PostgreSQLRepository) InsertSubject(s Subject) (int64, error) {
	var id int64 = -1
	sql := `INSERT INTO subjects (name, description) VALUES ($1, $2) RETURNING id`

	err := pr.db.QueryRow(sql, s.Name, s.Description).Scan(&id)
	if err != nil {
		return id, err
	}

	return id, nil
}

func (pr PostgreSQLRepository) GetSubject(id int64) (s Subject, err error) {
	row := pr.db.QueryRow(`SELECT * FROM subjects WHERE id = $1`, id)

	err = row.Scan(&s.ID, &s.Name, &s.Description)
	return s, err
}

func (pr PostgreSQLRepository) UpdateSubject(s Subject) (Subject, error) {
	uSubject := Subject{}
	sql := `UPDATE subjects SET name=$2, description=$3 WHERE id=$1 RETURNING id, name, description`
	err := pr.db.QueryRow(sql, s.ID, s.Name, s.Description).Scan(&uSubject.ID, &uSubject.Name, &uSubject.Description)
	if err != nil {
		return uSubject, err
	}

	return uSubject, nil
}

func (pr PostgreSQLRepository) DeleteSubject(id int64) (int64, error) {
	result, err := pr.db.Exec(`DELETE FROM subjects WHERE id=$1`, id)
	if err != nil {
		return -1, err
	}

	return result.RowsAffected()
}

// Events

// InsertEvent implements [Repository].
func (pr PostgreSQLRepository) InsertEvent(e Event) (int64, error) {
	var id int64 = -1
	sql := `INSERT INTO events (subject_id, occurrences, insert_ts, insert_utc, last_update) VALUES ($1, $2, $3, $4, $5) RETURNING id`

	err := pr.db.QueryRow(sql, e.SubjectID, e.Occurrences, e.InsertTS, e.InsertUTC, e.LastUpdate).Scan(&id)
	if err != nil {
		if strings.Contains(err.Error(), "violates foreign key constraint \"fk_subjects\"") {
			return id, ErrorSubjectIDNotFound{
				SubjectID: e.SubjectID,
			}
		}
		return id, err
	}

	return id, nil
}

// GetEvent implements [Repository].
func (pr PostgreSQLRepository) GetEvent(id int64) (e Event, err error) {
	row := pr.db.QueryRow(`SELECT * FROM events WHERE id = $1`, id)

	err = row.Scan(&e.ID, &e.SubjectID, &e.Occurrences, &e.InsertTS, &e.InsertUTC, &e.LastUpdate)
	if err != nil && strings.Contains(err.Error(), "no rows in result set") {
		return e, ErrorEventIDNotFound{
			EventID: e.ID,
		}
	}
	return e, err
}

// UpdateEvent implements [Repository].
func (pr PostgreSQLRepository) UpdateEvent(e Event) (Event, error) {
	uEvent := Event{}
	sql := `UPDATE events SET subject_id=$2, occurrences=$3, last_update=$4 WHERE id=$1 RETURNING id, subject_id, occurrences, insert_ts, insert_utc, last_update`
	err := pr.db.QueryRow(sql, e.ID, e.SubjectID, e.Occurrences, e.LastUpdate).Scan(&uEvent.ID, &uEvent.SubjectID, &uEvent.Occurrences, &uEvent.InsertTS, &uEvent.InsertUTC, &uEvent.LastUpdate)
	if err != nil {
		if strings.Contains(err.Error(), "no rows in result set") {
			return e, ErrorEventIDNotFound{
				EventID: e.ID,
			}
		}
		return e, err
	}

	return uEvent, nil
}

// DeleteEvent implements [Repository].
func (pr PostgreSQLRepository) DeleteEvent(id int64) (int64, error) {
	result, err := pr.db.Exec(`DELETE FROM events WHERE id=$1`, id)
	if err != nil {
		return -1, err
	}

	return result.RowsAffected()
}

// Verifies if there is already a entry with the same date (based on insert_utc)
func (pr PostgreSQLRepository) CheckEventExistenceByDate(insert_utc time.Time) (foundE Event, err error) {
	err = pr.db.QueryRow(`SELECT * FROM events WHERE DATE(insert_utc)=$1`, insert_utc.Format(time.DateOnly)).Scan(&foundE.ID, &foundE.SubjectID, &foundE.Occurrences, &foundE.InsertTS, &foundE.LastUpdate)
	if err != nil {
		return foundE, err
	}

	return foundE, nil
}

// Custom Errors for this postgres driver
type ErrorSubjectIDNotFound struct {
	SubjectID int64
}

func (e ErrorSubjectIDNotFound) Error() string {
	return fmt.Sprintf("subject id %d not found", e.SubjectID)
}

type ErrorEventIDNotFound struct {
	EventID int64
}

func (e ErrorEventIDNotFound) Error() string {
	return fmt.Sprintf("event id %d not found", e.EventID)
}
