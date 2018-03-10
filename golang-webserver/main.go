package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
	lane "gopkg.in/oleiade/lane.v1"
)

type PictureResult struct {
	Filename string `json:"filename"`
	Detected int    `json:"detected"`
}

type ProgressResponse struct {
	Success   bool            `json:"success"`
	Pictures  []PictureResult `json:"pictures"`
	TotalTime int             `json:"total_time"`
}

type WorkPacket struct {
	Filename  string `json:"filename"`
	Processed int    `json:"processed"`
}

var (
	queue     *lane.Queue
	db        *sql.DB
	startTime time.Time
)

func tes() {

	stat, err := db.Prepare("INSERT INTO people (firstname, lastname) VALUES (?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	stat.Exec("Nic", "Raboy")
}

func init() {
	db, _ = sql.Open("sqlite3", "./pictures.db")
	queue = lane.NewQueue()
	LoadWorkQueue(queue)
	//SetupDatabase()
}

func SetupDatabase() {
	_, err := db.Exec("DROP TABLE IF EXISTS jobs;")
	if err != nil {
		log.Fatal(err)
	}

	stat, _ := db.Prepare(`CREATE TABLE jobs (
					id INTEGER PRIMARY KEY AUTOINCREMENT,
					filename NOT NULL,
					detected INTEGER DEFAULT NULL`)
	stat.Exec()
}

func LoadWorkQueue(queue *lane.Queue) {
	for i := 0; i < 10; i++ {
		queue.Enqueue(&WorkPacket{
			Filename:  "filename" + string(i),
			Processed: 0,
		})
	}
}

func ResetDatabase(w http.ResponseWriter, req *http.Request) {
	log.Fatal("Not Implemented")
}

func GetWork(w http.ResponseWriter, req *http.Request) {
	if queue.Head() != nil {
		work := queue.Dequeue()
		queue.Enqueue(work)
		json.NewEncoder(w).Encode(work)
	} else {
		json.NewEncoder(w).Encode(&WorkPacket{
			Filename:  "",
			Processed: 1,
		})
	}
}

func PostResult(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	filename := params["filename"]
	detected := params["detected"]

	if detected == "" || filename == "" {
		http.Error(w, "Must Provide filename and detected params", http.StatusBadRequest)
		return
	}

	stat, err := db.Prepare("INSERT INTO pictures (filename, detected) VALUES (?, ?)")
	if err != nil {
		log.Fatal(err)
	}

	stat.Exec(filename, detected)

	w.Write([]byte("OK"))
}

func GetProgress(w http.ResponseWriter, req *http.Request) {
	//Get all results from database
	var pictures []PictureResult

	rows, _ := db.Query("SELECT filename, detected FROM pictures")

	var filename string
	var detected int
	for rows.Next() {
		rows.Scan(&filename, &detected)
		pictures = append(pictures, PictureResult{
			Filename: filename,
			Detected: detected,
		})
	}

	//Calculate total time
	totalTime := 2

	json.NewEncoder(w).Encode(&ProgressResponse{
		Success:   true,
		Pictures:  pictures,
		TotalTime: totalTime,
	})

}

func test(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte("OK"))
}

func main() {
	router := mux.NewRouter()
	port := ":3000"

	router.HandleFunc("/", GetWork).Methods("GET")
	router.HandleFunc("/getProgress", GetProgress).Methods("GET")
	router.HandleFunc("/processed", PostResult).Methods("POST")
	router.HandleFunc("/resetDb", ResetDatabase).Methods("POST")
	router.HandleFunc("/test", test).Methods("GET")

	log.Println("Listening on port: ", port)
	log.Fatal(http.ListenAndServe(port, router))
}
