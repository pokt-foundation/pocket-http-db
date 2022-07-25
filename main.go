package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pokt-foundation/pocket-http-driver/environment"
	postgresdriver "github.com/pokt-foundation/portal-api-go/postgres-driver"
)

var (
	connectionString = environment.GetString("CONNECTION_STRING", "")
	port             = environment.GetString("PORT", "8080")

	driver *postgresdriver.PostgresDriver
)

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, err := w.Write(response)
	if err != nil {
		panic(err)
	}
}

func getBlockchains(w http.ResponseWriter, r *http.Request) {
	blockchains, err := driver.ReadBlockchains()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())

		return
	}

	respondWithJSON(w, http.StatusOK, blockchains)
}

func getBlockchain(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	blockchain, err := driver.ReadBlockchainByID(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())

		return
	}

	respondWithJSON(w, http.StatusOK, blockchain)
}

func getApplication(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	application, err := driver.ReadApplicationByID(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())

		return
	}

	respondWithJSON(w, http.StatusOK, application)
}

func getLoadbalancer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	loadbalancer, err := driver.ReadLoadbalancerByID(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())

		return
	}

	respondWithJSON(w, http.StatusOK, loadbalancer)
}

func main() {
	var err error

	driver, err = postgresdriver.NewPostgresDriverFromConnectionString(connectionString)
	if err != nil {
		panic(fmt.Sprintf("connection to database failed with error: %s", err.Error()))
	}

	r := mux.NewRouter()
	r.HandleFunc("/blockchain", getBlockchains)
	r.HandleFunc("/blockchain/{id}", getBlockchain)
	r.HandleFunc("/application/{id}", getApplication)
	r.HandleFunc("/load_balancer/{id}", getLoadbalancer)

	http.Handle("/", r)

	log.Fatal(http.ListenAndServe(":"+port, nil))
}
