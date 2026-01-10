package controllers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"
	
	"backend-wisata/config"
	"backend-wisata/models"
	
	"golang.org/x/crypto/bcrypt"
)

func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func Login(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	
	var input struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	var user models.User
	var passwordHash string
	
	query := "SELECT id, username, role, password_hash FROM users WHERE username = $1 AND deleted_at IS NULL"
	err := config.DB.QueryRow(query, input.Username).Scan(&user.ID, &user.Username, &user.Role, &passwordHash)
	
	if err == sql.ErrNoRows {
		http.Error(w, "Username tidak ditemukan", http.StatusUnauthorized)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	if !checkPasswordHash(input.Password, passwordHash) {
		http.Error(w, "Password salah", http.StatusUnauthorized)
		return
	}
	
	http.SetCookie(
		w, &http.Cookie{
			Name:    "session_token",
			Value:   user.Username,
			Expires: time.Now().Add(24 * time.Hour),
			Path:    "/",
		},
	)
	
	json.NewEncoder(w).Encode(
		models.Response{
			Status:  200,
			Message: "Login Berhasil",
			Data:    user,
		},
	)
}

func Logout(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	http.SetCookie(
		w, &http.Cookie{
			Name:    "session_token",
			Value:   "",
			Expires: time.Now().Add(-1 * time.Hour),
			Path:    "/",
		},
	)
	json.NewEncoder(w).Encode(
		models.Response{
			Status:  200,
			Message: "Logout Berhasil",
		},
	)
}
