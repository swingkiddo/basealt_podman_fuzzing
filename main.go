package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"log"
	"net/http"
)

const (
	DB_USER     = "postgres"
	DB_PASSWORD = "pravda"
	DB_NAME     = "podman_fuzzing"
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
	fmt.Println("hold on")
	db := setupDB()
	rows, err := db.Query("SELECT * FROM targets")
	checkErr(err)

	var targets []Target

	for rows.Next() {
		var name string
		var packageDir string
		var source string

		err = rows.Scan(&name, &packageDir, &source)
		checkErr(err)

		targets = append(targets, Target{Name: name, PackageDir: packageDir, Source: source})
	}

	var response = JsonResponse{Type: "success", Data: targets}
	json.NewEncoder(w).Encode(response)
}

func main() {
	router := mux.NewRouter()

	router.HandleFunc("/targets/", GetTargets).Methods("GET")

	fmt.Println("Server at 8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
