package repository

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"

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

func (pr PostgreSQLRepository) InsertSubject(s Subject) (id int, err error) {
	sql := `INSERT INTO subjects (name, description) VALUES ($1, $2)`

	err = pr.db.QueryRow(sql, s.Name, s.Description).Scan(&id)

	return id, err
}

func (pr PostgreSQLRepository) GetSubject(id int) (s Subject, err error) {
	row := pr.db.QueryRow(`SELECT * FROM subjects WHERE id = $1`, id)

	err = row.Scan(&s.ID, &s.Name, &s.Description)
	return s, err
}

func (pr PostgreSQLRepository) UpdateSubject(s Subject) (int64, error) {
	result, err := pr.db.Exec(`UPDATE subjects SET name=$2, description=$3 WHERE id=$1`, s.ID, s.Name, s.Description)
	if err != nil {
		return -1, err
	}

	return result.RowsAffected()
}

func (pr PostgreSQLRepository) DeleteSubject(id int) (int64, error) {
	panic("unimplemented")
}

// InsertEvent implements [Repository].
func (pr PostgreSQLRepository) InsertEvent(e Event) (id int, err error) {
	sql := `INSERT INTO events (subject_id, ocurrences, insert_ts, last_update) VALUES ($1, $2, $3, $4)`

	err = pr.db.QueryRow(sql, e.SubjectID, e.Ocurrences, e.InsertTS, e.LastUpdate).Scan(&id)

	return id, err
}

// GetEvent implements [Repository].
func (pr PostgreSQLRepository) GetEvent(id int) (e Event, err error) {
	row := pr.db.QueryRow(`SELECT * FROM events WHERE id = $1`, id)

	err = row.Scan(&e.SubjectID, &e.Ocurrences, &e.InsertTS, &e.LastUpdate)
	return e, err
}

// UpdateEvent implements [Repository].
func (pr PostgreSQLRepository) UpdateEvent(e Event) (int64, error) {
	result, err := pr.db.Exec(`UPDATE events SET subject_id=$2, ocurrences=$3, insert_ts=$4, last_update=CURRENT_TIMESTAMP() WHERE id=$1`, e.ID, e.SubjectID, e.Ocurrences, e.InsertTS)
	if err != nil {
		return -1, err
	}

	return result.RowsAffected()
}

// DeleteEvent implements [Repository].
func (pr PostgreSQLRepository) DeleteEvent(id int) (int64, error) {
	result, err := pr.db.Exec(`DELETE FROM events WHERE id=$1`, id)
	if err != nil {
		return -1, err
	}

	return result.RowsAffected()
}
