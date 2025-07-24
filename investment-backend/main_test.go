package main

import (
	"encoding/csv"
	"os"
	"strconv"
	"testing"
)

func setupTestCSV(t *testing.T) string {
	file, err := os.CreateTemp("", "portfolio_test_*.csv")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	w := csv.NewWriter(file)
	headers := []string{"id", "name", "email", "totalInvestment"}
	_ = w.Write(headers)
	_ = w.Write([]string{"1", "Alice", "alice@example.com", "1000.00"})
	_ = w.Write([]string{"2", "Bob", "bob@example.com", "20000.00"})
	_ = w.Write([]string{"3", "Carol", "carol@example.com", "300000.00"})
	w.Flush()
	file.Close()
	return file.Name()
}

func cleanupTestCSV(path string) {
	_ = os.Remove(path)
}

func TestGetClients(t *testing.T) {
	csvPath := setupTestCSV(t)
	defer cleanupTestCSV(csvPath)
	store := NewCSVClientStore(csvPath)

	clients, err := store.GetClients(nil, nil)
	if err != nil || len(clients) != 3 {
		t.Fatalf("expected 3 clients, got %d, err: %v", len(clients), err)
	}

	min := 20000.0
	clients, err = store.GetClients(&min, nil)
	if err != nil || len(clients) != 2 {
		t.Fatalf("expected 2 clients >= 20000, got %d, err: %v", len(clients), err)
	}

	max := 20000.0
	clients, err = store.GetClients(nil, &max)
	if err != nil || len(clients) != 2 {
		t.Fatalf("expected 2 clients <= 20000, got %d, err: %v", len(clients), err)
	}

	clients, err = store.GetClients(&min, &max)
	if err != nil || len(clients) != 1 {
		t.Fatalf("expected 1 client == 20000, got %d, err: %v", len(clients), err)
	}
}

func TestGetClientByID(t *testing.T) {
	csvPath := setupTestCSV(t)
	defer cleanupTestCSV(csvPath)
	store := NewCSVClientStore(csvPath)

	client, err := store.GetClientByID("2")
	if err != nil || client == nil || client.Name != "Bob" {
		t.Fatalf("expected Bob, got %+v, err: %v", client, err)
	}

	client, err = store.GetClientByID("999")
	if err != nil || client != nil {
		t.Fatalf("expected nil for non-existent client, got %+v, err: %v", client, err)
	}
}

func TestCreateClient(t *testing.T) {
	csvPath := setupTestCSV(t)
	defer cleanupTestCSV(csvPath)
	store := NewCSVClientStore(csvPath)

	newClient := Client{ID: "4", Name: "Dave", Email: "dave@example.com", TotalInvestment: 4000.0}
	err := store.CreateClient(newClient)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	client, err := store.GetClientByID("4")
	if err != nil || client == nil || client.Name != "Dave" {
		t.Fatalf("expected Dave, got %+v, err: %v", client, err)
	}
}

func TestUpdateClient(t *testing.T) {
	csvPath := setupTestCSV(t)
	defer cleanupTestCSV(csvPath)
	store := NewCSVClientStore(csvPath)

	updated := Client{ID: "2", Name: "Bobby", Email: "bobby@example.com", TotalInvestment: 22222.0}
	ok, err := store.UpdateClient("2", updated)
	if err != nil || !ok {
		t.Fatalf("failed to update client: %v", err)
	}
	client, _ := store.GetClientByID("2")
	if client.Name != "Bobby" || client.TotalInvestment != 22222.0 {
		t.Fatalf("update did not persist, got %+v", client)
	}

	ok, err = store.UpdateClient("999", updated)
	if ok || err != nil {
		t.Fatalf("should not update non-existent client")
	}
}

func TestDeleteClient(t *testing.T) {
	csvPath := setupTestCSV(t)
	defer cleanupTestCSV(csvPath)
	store := NewCSVClientStore(csvPath)

	ok, err := store.DeleteClient("2")
	if err != nil || !ok {
		t.Fatalf("failed to delete client: %v", err)
	}
	client, _ := store.GetClientByID("2")
	if client != nil {
		t.Fatalf("client not deleted")
	}

	ok, err = store.DeleteClient("999")
	if ok || err != nil {
		t.Fatalf("should not delete non-existent client")
	}
} 