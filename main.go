package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"io"
	"log"
	"net/http"
)

const (
	DB_USER     = "postgres"
	DB_PASSWORD = "pravda"
	DB_NAME     = "podman_fuzzing"
	PORT        = "9000"
)

var (
	db *sql.DB
)

func setupDB() *sql.DB {
	dbinfo := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", DB_USER, DB_PASSWORD, DB_NAME)
	db, err := sql.Open("postgres", dbinfo)

	checkErr(err)

	return db
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

type Target struct {
	Name                 string `json:"name"`
	PackageDir           string `json:"package"`
	Source               string `json:"source"`
	Coverage             int    `json:"coverage"`
	CyclomaticComplexity int    `json:"cyclomaticComplexity"`
}

type JsonResponse struct {
	Type    string   `json:"type"`
	Data    []Target `json:"data"`
	Message string   `json:"message"`
}

func GetTargets(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT * FROM targets")
	checkErr(err)

	var targets []Target

	for rows.Next() {
		var (
			id                   int
			name                 string
			packageDir           string
			source               string
			coverage             interface{}
			cyclomaticComplexity interface{}
		)

		err = rows.Scan(&id, &name, &packageDir, &source, &coverage, &cyclomaticComplexity)
		checkErr(err)

		targets = append(targets, Target{Name: name, PackageDir: packageDir, Source: source})
	}

	var response = JsonResponse{Type: "success", Data: targets}
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Fatal(err)
	}
}

func AddTarget(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}
	data := make(map[string]interface{})
	err = json.Unmarshal(body, &data)
	checkErr(err)

	name := data["name"]
	packageDir := data["package"]
	source := data["source"]
	_, err = db.Query("INSERT INTO targets(name, package, source) VALUES($1, $2, $3);", name, packageDir, source)
	checkErr(err)

	response := JsonResponse{Type: "success", Message: "The target has been added successfully"}
	json.NewEncoder(w).Encode(response)
}

func DeleteTarget(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]
	_, err := db.Exec("DELETE FROM targets WHERE  id = $1", id)
	checkErr(err)
	response := JsonResponse{Type: "success", Message: "The target has been deleted succesfully"}
	json.NewEncoder(w).Encode(response)
}

func main() {
	db = setupDB()
	defer db.Close()

	router := mux.NewRouter()

	router.HandleFunc("/targets", GetTargets).Methods("GET")
	router.HandleFunc("/targets", AddTarget).Methods("POST")
	router.HandleFunc("/targets/{id}", DeleteTarget).Methods("DELETE")
	fmt.Println(router)

	fmt.Println("Server at ", PORT)
	if err := http.ListenAndServe("localhost:"+PORT, router); err != nil {
		log.Fatal(err)
	}
}
