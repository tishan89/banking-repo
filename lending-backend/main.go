package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
)

type LendingClient struct {
	ID              string  `json:"id"`
	Name            string  `json:"name"`
	Email           string  `json:"email"`
	TotalLend       float64 `json:"totalLend"`
}

type LendingClientStore interface {
	GetLendingClients(minLend, maxLend *float64) ([]LendingClient, error)
	// Other CRUD methods can be added here
}

type CSVLendingClientStore struct {
	filePath string
}

func NewCSVLendingClientStore(filePath string) *CSVLendingClientStore {
	return &CSVLendingClientStore{filePath: filePath}
}

func (s *CSVLendingClientStore) GetLendingClients(minLend, maxLend *float64) ([]LendingClient, error) {
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

	var clients []LendingClient
	for i, rec := range records {
		if i == 0 {
			continue // skip header
		}
		lend, err := strconv.ParseFloat(rec[3], 64)
		if err != nil {
			continue
		}
		client := LendingClient{
			ID:              rec[0],
			Name:            rec[1],
			Email:           rec[2],
			TotalLend:      lend,
		}
		if minLend != nil && client.TotalLend < *minLend {
			continue
		}
		if maxLend != nil && client.TotalLend > *maxLend {
			continue
		}
		clients = append(clients, client)
	}
	return clients, nil
}

func (s *CSVLendingClientStore) GetLendingClientByID(id string) (*LendingClient, error) {
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
			lend, _ := strconv.ParseFloat(rec[3], 64)
			return &LendingClient{
				ID: rec[0], Name: rec[1], Email: rec[2], TotalLend: lend,
			}, nil
		}
	}
	return nil, nil
}

func (s *CSVLendingClientStore) CreateLendingClient(c LendingClient) error {
	file, err := os.OpenFile(s.filePath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	w := csv.NewWriter(file)
	defer w.Flush()
	return w.Write([]string{c.ID, c.Name, c.Email, fmt.Sprintf("%.2f", c.TotalLend)})
}

func (s *CSVLendingClientStore) UpdateLendingClient(id string, updated LendingClient) (bool, error) {
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
			records[i] = []string{id, updated.Name, updated.Email, fmt.Sprintf("%.2f", updated.TotalLend)}
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

func (s *CSVLendingClientStore) DeleteLendingClient(id string) (bool, error) {
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
func getLendingClientsHandler(store LendingClientStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var minLend, maxLend *float64
		if min := r.URL.Query().Get("minLend"); min != "" {
			val, err := strconv.ParseFloat(min, 64)
			if err == nil {
				minLend = &val
			}
		}
		if max := r.URL.Query().Get("maxLend"); max != "" {
			val, err := strconv.ParseFloat(max, 64)
			if err == nil {
				maxLend = &val
			}
		}
		clients, err := store.GetLendingClients(minLend, maxLend)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Error: %v", err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(clients)
	}
}

func getLendingClientByIDHandler(store LendingClientStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Path[len("/lending-clients/"):]
		client, err := store.(*CSVLendingClientStore).GetLendingClientByID(id)
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

func createLendingClientHandler(store LendingClientStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var c LendingClient
		if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "Invalid body")
			return
		}
		if c.ID == "" {
			c.ID = strconv.FormatInt(int64(os.Getpid())+int64(os.Getuid())+int64(os.Geteuid())+int64(os.Getppid())+int64(os.Getgid())+int64(os.Getegid())+int64(os.Getppid())+int64(os.Getpid())+int64(os.Getppid()), 10)
		}
		err := store.(*CSVLendingClientStore).CreateLendingClient(c)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Error: %v", err)
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(c)
	}
}

func updateLendingClientHandler(store LendingClientStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Path[len("/lending-clients/"):]
		var c LendingClient
		if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "Invalid body")
			return
		}
		ok, err := store.(*CSVLendingClientStore).UpdateLendingClient(id, c)
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

func deleteLendingClientHandler(store LendingClientStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Path[len("/lending-clients/"):]
		ok, err := store.(*CSVLendingClientStore).DeleteLendingClient(id)
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
	store := NewCSVLendingClientStore("lending_accounts.csv")

	http.HandleFunc("/lending-clients", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			createLendingClientHandler(store)(w, r)
			return
		}
		getLendingClientsHandler(store)(w, r)
	})

	http.HandleFunc("/lending-clients/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			getLendingClientByIDHandler(store)(w, r)
		case http.MethodPut:
			updateLendingClientHandler(store)(w, r)
		case http.MethodDelete:
			deleteLendingClientHandler(store)(w, r)
		}
	})

	fmt.Println("Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
