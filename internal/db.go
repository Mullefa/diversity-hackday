package internal

import (
	"database/sql"
	"fmt"
	"net/url"
	"time"

	"github.com/aws/aws-sdk-go/service/rekognition"

	_ "github.com/lib/pq"
)

type HackdayDB struct {
	db                         *sql.DB
	insertJournalist           *sql.Stmt
	insertArticle              *sql.Stmt
	isJournalistExist          *sql.Stmt
	selectJournalists          *sql.Stmt
	updateJournalistImage      *sql.Stmt
	updateJournalistHeuristics *sql.Stmt
}

func NewHackdayDB(db *sql.DB) (*HackdayDB, error) {
	rawStmts := map[int]string{
		0: "INSERT INTO journalists(journalist_name, profile_url) VALUES ($1, $2) ON CONFLICT DO NOTHING",
		1: "INSERT INTO articles(first_publication_date, id, web_url, journalist_name) VALUES ($1, $2, $3, $4) ON CONFLICT DO NOTHING",
		2: "SELECT true FROM journalists WHERE journalist_name = $1",
		3: "SELECT journalist_name, profile_url, image_filename, gender FROM journalists ORDER BY journalist_name",
		4: "UPDATE journalists SET image_filename = $2 WHERE journalist_name = $1",
		5: "UPDATE journalists SET gender = $2, gender_confidence = $3, age_range_low = $4, age_range_high = $5 WHERE journalist_name = $1",
	}
	stmts := make(map[int]*sql.Stmt)
	for id, rawStmt := range rawStmts {
		stmt, err := db.Prepare(rawStmt)
		if err != nil {
			return nil, fmt.Errorf("unable to prepare statement %s: %w", rawStmt, err)
		}
		stmts[id] = stmt
	}
	return &HackdayDB{
		db:                         db,
		insertJournalist:           stmts[0],
		insertArticle:              stmts[1],
		isJournalistExist:          stmts[2],
		selectJournalists:          stmts[3],
		updateJournalistImage:      stmts[4],
		updateJournalistHeuristics: stmts[5],
	}, nil
}

func NewDefaultHackdayDB() (*HackdayDB, error) {
	connStr := "user=gd password=pw dbname=hackday sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to Postgres: %w", err)
	}
	return NewHackdayDB(db)
}

func (db HackdayDB) Close() error {
	return db.db.Close()
}

func (db HackdayDB) InsertJournalist(info JournalistInfo) error {
	profileUrl := sql.NullString{}
	if info.ProfileURL != nil {
		profileUrl = sql.NullString{
			String: info.ProfileURL.String(),
			Valid:  true,
		}
	}
	_, err := db.insertJournalist.Exec(info.Name, profileUrl)
	return err
}

func (db HackdayDB) InsertArticle(firstPublicationDate time.Time, id string, webURL *url.URL, journalistName string) error {
	_, err := db.insertArticle.Exec(firstPublicationDate, id, webURL.String(), journalistName)
	return err
}

type JournalistRows struct {
	rows *sql.Rows
}

func (rows JournalistRows) Next() bool {
	return rows.rows.Next()
}

func (rows JournalistRows) Scan() (JournalistInfo, error) {
	var (
		name          string
		rawProfileUrl sql.NullString
		imageName     sql.NullString
		gender        sql.NullString
	)

	if err := rows.rows.Scan(&name, &rawProfileUrl, &imageName, &gender); err != nil {
		return JournalistInfo{}, err
	}

	var profileUrl *url.URL
	if rawProfileUrl.Valid {
		profileUrl, _ = url.Parse(rawProfileUrl.String)
	}

	return JournalistInfo{
		ProfileURL: profileUrl,
		Name:       name,
		ImageName:  imageName.String,
		Gender:     gender.String,
	}, nil
}

func (rows JournalistRows) Close() error {
	return rows.rows.Close()
}

func (rows JournalistRows) Err() error {
	return rows.rows.Err()
}

func (db HackdayDB) GetJournalists() (JournalistRows, error) {
	rows, err := db.selectJournalists.Query()
	if err != nil {
		return JournalistRows{}, err
	}
	return JournalistRows{rows: rows}, nil
}

func (db HackdayDB) UpdateJournalistImage(name string, filename string) error {
	res, err := db.updateJournalistImage.Exec(name, filename)
	if err != nil {
		return err
	}
	if n, err := res.RowsAffected(); err == nil && n != 1 {
		return fmt.Errorf("expected rows affected: 1, actual: %d", n)
	}
	return nil
}

func (db HackdayDB) UpdateJournalistHeuristics(name string, details *rekognition.FaceDetail) error {
	res, err := db.updateJournalistHeuristics.Exec(name, details.Gender.Value, details.Gender.Confidence, details.AgeRange.Low, details.AgeRange.High)
	if err != nil {
		return err
	}
	if n, err := res.RowsAffected(); err == nil && n != 1 {
		return fmt.Errorf("expected rows affected: 1, actual: %d", n)
	}
	return nil
}
