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
	db              *sql.DB
	currentLocation *time.Location
}

func NewPostgreSQLRepository() Repository {
	db, err := OpenConn()
	if err != nil {
		panic(fmt.Sprintf("could not connect to PostreSQL: %v", err))
	}

	local, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		panic(fmt.Sprintf("could not load location: %v", err))
	}

	return PostgreSQLRepository{
		db:              db,
		currentLocation: local,
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
		log.Println("Error loading .env file: ", err)
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
	if err != nil && strings.Contains(err.Error(), "no rows in result set") {
		return s, ErrorSubjectIDNotFound{
			SubjectID: id,
		}
	}

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

/* Events */

// InsertEvent implements [Repository].
func (pr PostgreSQLRepository) InsertEvent(e Event) (int64, error) {
	var id int64 = -1
	sql := `INSERT INTO events (subject_id, occurrences, insert_ts, last_update) VALUES ($1, $2, $3, $4) RETURNING id`

	err := pr.db.QueryRow(sql, e.SubjectID, e.Occurrences, e.InsertTS, e.LastUpdate).Scan(&id)
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

	err = row.Scan(&e.ID, &e.SubjectID, &e.Occurrences, &e.InsertTS, &e.LastUpdate)
	if err != nil && strings.Contains(err.Error(), "no rows in result set") {
		return e, ErrorEventIDNotFound{
			EventID: id,
		}
	}
	convertTimesToLocalTime(&e)
	return e, err
}

// UpdateEvent implements [Repository].
func (pr PostgreSQLRepository) UpdateEvent(e Event) (Event, error) {
	uEvent := Event{}
	sql := `UPDATE events SET subject_id=$2, occurrences=$3, last_update=$4 WHERE id=$1 RETURNING id, subject_id, occurrences, insert_ts, last_update`
	err := pr.db.QueryRow(sql, e.ID, e.SubjectID, e.Occurrences, e.LastUpdate).Scan(&uEvent.ID, &uEvent.SubjectID, &uEvent.Occurrences, &uEvent.InsertTS, &uEvent.LastUpdate)
	if err != nil {
		if strings.Contains(err.Error(), "no rows in result set") {
			return e, ErrorEventIDNotFound{
				EventID: e.ID,
			}
		}
		return e, err
	}

	convertTimesToLocalTime(&uEvent)
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


//	Verifies if there is already a entry with the same date (based on insert_ts) converting both to local date, via DATE and AT TIME ZONE
// passing the currently loaded IANA timezone, in database. insert_ts should always be on localtime
func (pr PostgreSQLRepository) CheckEventExistenceByDate(insertTS time.Time) (foundE Event, err error) {
	sql := `SELECT * FROM events WHERE DATE(insert_ts AT TIME ZONE $2)=DATE($1 AT TIME ZONE $2)`
	err = pr.db.QueryRow(sql, insertTS.Format(time.RFC3339), pr.currentLocation.String()).Scan(&foundE.ID, &foundE.SubjectID, &foundE.Occurrences, &foundE.InsertTS, &foundE.LastUpdate)
	if err != nil {
		return foundE, err
	}

	convertTimesToLocalTime(&foundE)
	return foundE, nil
}

func convertTimesToLocalTime(event *Event) {
	event.InsertTS = event.InsertTS.Local()
	event.LastUpdate = event.LastUpdate.Local()
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
