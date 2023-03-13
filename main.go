package main

import (
	"bufio"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
)

var db *sql.DB
var iLog *log.Logger
var logFilePath = `./tmp/logger.log`

var DATA []Album

type Album struct {
	ReleaseDate string
	Price       float32
	Title       string
	Author      string
}

func main() {
	// Database configurations
	cfg := mysql.Config{
		User:                 "root",
		Passwd:               "",
		Net:                  "tcp",
		Addr:                 "127.0.0.1:3306",
		DBName:               "records",
		AllowNativePasswords: true,
	}

	// Opening the log file
	logFile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		log.Fatalf("Couldn't open the log file '%s'", logFilePath)
	}
	defer logFile.Close()

	iLog = log.New(logFile, "records-database:", log.LstdFlags)
	iLog.SetFlags(log.LstdFlags | log.Lshortfile)

	// Opening the database
	db, err = sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		iLog.Println(err)
	}

	// Testing connection to the database
	dbPing := db.Ping()
	if dbPing != nil {
		iLog.Fatalln("Database is not connected!")
	}
	fmt.Println("Connected!")

	// Reading data from the file
	_, _ = ReadingDataFromFiles(`./tmp/albums.txt`)

	// Inserting data into the table 'album' in the database
	err = InsertDataIntoDBFromFile()
	if err != nil {
		iLog.Fatalf("Error occured during inserting data into the db: '%v'\n", err)
	}
}

func ReadingDataFromFiles(filenames ...string) ([]Album, error) {
	for _, filename := range filenames {
		f, err := os.Open(filename)
		if err != nil {
			iLog.Printf("Didn't manage to read data from file '%s'\n", filename)
			continue
		}
		defer f.Close()

		reader := bufio.NewReader(f)
		for {
			line, err := reader.ReadString('\n')
			if err == io.EOF {
				break
			} else if err != nil {
				iLog.Printf("Error while reading a line in the file '%s'\n", filename)
				continue
			}

			lineSplited := strings.Split(line, "|")
			for _, value := range lineSplited {
				value = strings.TrimRight(value, " ")
			}

			if dateFormatCheck := checkDate(lineSplited); dateFormatCheck != nil {
				continue
			}

			var price float64
			price, err = strconv.ParseFloat(lineSplited[1], 64)
			if err != nil {
				price = -1
				iLog.Printf("Not a correct price format passed in the file '%s'\n", filename)
			}

			newAlbum := Album{
				ReleaseDate: lineSplited[0],
				Price:       float32(price),
				Title:       lineSplited[2],
				Author:      lineSplited[3],
			}
			DATA = append(DATA, newAlbum)
		}
	}
	return DATA, nil
}

func checkDate(date []string) error {
	r := regexp.MustCompile(`.*\[(\d\d\/\w+/\d\d\d\d:\d\d:\d\d:\d\d.*)\].*`)
	if r.MatchString(date[0]) {
		match := r.FindStringSubmatch(date[0])

		dt, err := time.Parse("02/Jan/2006:15:04:05 -0700", match[1])
		if err == nil {
			newFormat := dt.Format(time.RFC850)
			date[0] = newFormat

		} else {
			date[0] = "None"
		}

		return nil
	} else {
		return errors.New("Not a valid date format")
	}
}

func InsertDataIntoDBFromFile() error {
	for _, albumData := range DATA {
		_, err := db.Exec("INSERT INTO album (released, price, title, author) VALUES (?,?,?,?)", albumData.ReleaseDate, albumData.Price, albumData.Title, albumData.Author)
		if err != nil {
			return fmt.Errorf("Insertion error: %v", err)
		}
	}

	return nil
}

func InsertDataIntoDB(album Album) error {
	_, err := db.Exec("INSERT INTO album (released, price, title, author) VALUES (?,?,?,?)", album.ReleaseDate, album.Price, album.Title, album.Author)
	if err != nil {
		return fmt.Errorf("Insertion error: %v", err)
	}
	return nil
}
