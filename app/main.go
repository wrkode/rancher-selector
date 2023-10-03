package main

import (
	"encoding/json"
	"github.com/wrkode/rancher-selector/cmd"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/", handler)

	log.Println("Server starting on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST is supported", http.StatusMethodNotAllowed)
		return
	}

	var event cmd.ProjectEvent
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	err := cmd.CreateOrUpdateConfigMap(event, "rancher-data", "kube-system")
	if err != nil {
		log.Printf("Error creating or updating ConfigMap: %v", err)
		http.Error(w, "Failed to create or update ConfigMap", http.StatusInternalServerError)
		return
	}
}
