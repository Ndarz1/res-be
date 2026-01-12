package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"
	
	"backend-wisata/config"
	"backend-wisata/models"
)

func GetAllUsers(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	
	query := `
		SELECT id, full_name, email, phone, role, is_active, created_at
		FROM users
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
	`
	
	rows, err := config.DB.Query(query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	
	var users []models.User
	
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.FullName, &u.Email, &u.Phone, &u.Role, &u.IsActive, &u.CreatedAt); err != nil {
			continue
		}
		users = append(users, u)
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(
		models.Response{
			Status:  200,
			Message: "Users Fetched",
			Data:    users,
		},
	)
}

func UpdateUserStatus(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	
	if r.Method != "PUT" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	idStr := r.URL.Query().Get("id")
	id, _ := strconv.Atoi(idStr)
	
	var input struct {
		IsActive bool   `json:"is_active"`
		Role     string `json:"role"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}
	
	query := "UPDATE users SET is_active=$1, role=$2, updated_at=NOW() WHERE id=$3"
	_, err := config.DB.Exec(query, input.IsActive, input.Role, id)
	
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	json.NewEncoder(w).Encode(models.Response{Status: 200, Message: "User Updated"})
}

func DeleteUser(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	
	if r.Method != "DELETE" && r.Method != "POST" { // Allow POST for easier frontend handling
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	idStr := r.URL.Query().Get("id")
	id, _ := strconv.Atoi(idStr)
	
	// Soft delete
	query := "UPDATE users SET deleted_at=NOW() WHERE id=$1"
	_, err := config.DB.Exec(query, id)
	
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	json.NewEncoder(w).Encode(models.Response{Status: 200, Message: "User Deleted"})
}
