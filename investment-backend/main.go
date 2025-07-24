package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
)

type Client struct {
	ID              string  `json:"id"`
	Name            string  `json:"name"`
	Email           string  `json:"email"`
	TotalInvestment float64 `json:"totalInvestment"`
}

type ClientStore interface {
	GetClients(minInvestment, maxInvestment *float64) ([]Client, error)
	// Other CRUD methods can be added here
}

type CSVClientStore struct {
	filePath string
}

func NewCSVClientStore(filePath string) *CSVClientStore {
	return &CSVClientStore{filePath: filePath}
}

func (s *CSVClientStore) GetClients(minInvestment, maxInvestment *float64) ([]Client, error) {
	file, err := os.Open(s.filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	r := csv.NewReader(file)
	records, err := r.ReadAll()
	if err != nil {
		return nil, err
	}

	var clients []Client
	for i, rec := range records {
		if i == 0 {
			continue // skip header
		}
		inv, err := strconv.ParseFloat(rec[3], 64)
		if err != nil {
			continue
		}
		client := Client{
			ID:              rec[0],
			Name:            rec[1],
			Email:           rec[2],
			TotalInvestment: inv,
		}
		if minInvestment != nil && client.TotalInvestment < *minInvestment {
			continue
		}
		if maxInvestment != nil && client.TotalInvestment > *maxInvestment {
			continue
		}
		clients = append(clients, client)
	}
	return clients, nil
}

func (s *CSVClientStore) GetClientByID(id string) (*Client, error) {
	file, err := os.Open(s.filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	r := csv.NewReader(file)
	records, err := r.ReadAll()
	if err != nil {
		return nil, err
	}
	for i, rec := range records {
		if i == 0 {
			continue
		}
		if rec[0] == id {
			inv, _ := strconv.ParseFloat(rec[3], 64)
			return &Client{
				ID: rec[0], Name: rec[1], Email: rec[2], TotalInvestment: inv,
			}, nil
		}
	}
	return nil, nil
}

func (s *CSVClientStore) CreateClient(c Client) error {
	file, err := os.OpenFile(s.filePath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	w := csv.NewWriter(file)
	defer w.Flush()
	return w.Write([]string{c.ID, c.Name, c.Email, fmt.Sprintf("%.2f", c.TotalInvestment)})
}

func (s *CSVClientStore) UpdateClient(id string, updated Client) (bool, error) {
	file, err := os.Open(s.filePath)
	if err != nil {
		return false, err
	}
	records, err := csv.NewReader(file).ReadAll()
	file.Close()
	if err != nil {
		return false, err
	}
	found := false
	for i, rec := range records {
		if i == 0 {
			continue
		}
		if rec[0] == id {
			records[i] = []string{id, updated.Name, updated.Email, fmt.Sprintf("%.2f", updated.TotalInvestment)}
			found = true
			break
		}
	}
	if !found {
		return false, nil
	}
	fileW, err := os.Create(s.filePath)
	if err != nil {
		return false, err
	}
	defer fileW.Close()
	w := csv.NewWriter(fileW)
	defer w.Flush()
	return true, w.WriteAll(records)
}

func (s *CSVClientStore) DeleteClient(id string) (bool, error) {
	file, err := os.Open(s.filePath)
	if err != nil {
		return false, err
	}
	records, err := csv.NewReader(file).ReadAll()
	file.Close()
	if err != nil {
		return false, err
	}
	newRecords := [][]string{}
	found := false
	for i, rec := range records {
		if i == 0 || rec[0] != id {
			newRecords = append(newRecords, rec)
		} else {
			found = true
		}
	}
	if !found {
		return false, nil
	}
	fileW, err := os.Create(s.filePath)
	if err != nil {
		return false, err
	}
	defer fileW.Close()
	w := csv.NewWriter(fileW)
	defer w.Flush()
	return true, w.WriteAll(newRecords)
}

// Handler functions
func getClientsHandler(store ClientStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var minInvestment, maxInvestment *float64
		if min := r.URL.Query().Get("minInvestment"); min != "" {
			val, err := strconv.ParseFloat(min, 64)
			if err == nil {
				minInvestment = &val
			}
		}
		if max := r.URL.Query().Get("maxInvestment"); max != "" {
			val, err := strconv.ParseFloat(max, 64)
			if err == nil {
				maxInvestment = &val
			}
		}
		clients, err := store.GetClients(minInvestment, maxInvestment)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Error: %v", err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(clients)
	}
}

func getClientByIDHandler(store ClientStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Path[len("/clients/"):]
		client, err := store.(*CSVClientStore).GetClientByID(id)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Error: %v", err)
			return
		}
		if client == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(client)
	}
}

func createClientHandler(store ClientStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var c Client
		if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "Invalid body")
			return
		}
		if c.ID == "" {
			c.ID = strconv.FormatInt(int64(os.Getpid())+int64(os.Getuid())+int64(os.Geteuid())+int64(os.Getppid())+int64(os.Getgid())+int64(os.Getegid())+int64(os.Getppid())+int64(os.Getpid())+int64(os.Getppid()), 10)
		}
		err := store.(*CSVClientStore).CreateClient(c)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Error: %v", err)
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(c)
	}
}

func updateClientHandler(store ClientStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Path[len("/clients/"):]
		var c Client
		if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "Invalid body")
			return
		}
		ok, err := store.(*CSVClientStore).UpdateClient(id, c)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Error: %v", err)
			return
		}
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(c)
	}
}

func deleteClientHandler(store ClientStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Path[len("/clients/"):]
		ok, err := store.(*CSVClientStore).DeleteClient(id)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Error: %v", err)
			return
		}
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func main() {
	store := NewCSVClientStore("portfolio.csv")

	http.HandleFunc("/clients", getClientsHandler(store))
	http.HandleFunc("/clients/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			getClientByIDHandler(store)(w, r)
		case http.MethodPut:
			updateClientHandler(store)(w, r)
		case http.MethodDelete:
			deleteClientHandler(store)(w, r)
		}
	})
	http.HandleFunc("/clients", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			createClientHandler(store)(w, r)
			return
		}
		getClientsHandler(store)(w, r)
	})

	fmt.Println("Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
