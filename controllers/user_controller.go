package controllers

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
	
	"backend-wisata/config"
	"backend-wisata/models"
)

func GetAllUsers(w http.ResponseWriter, r *http.Request) {
	// HAPUS enableCors dan OPTIONS check
	
	query := `
		SELECT id, full_name, email, phone, role, is_active, created_at
		FROM users
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
	`
	
	rows, err := config.DB.Query(query)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.Response{Status: 500, Message: err.Error()})
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
	// HAPUS enableCors dan OPTIONS check
	
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
	// HAPUS enableCors dan OPTIONS check
	
	if r.Method != "DELETE" && r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	idStr := r.URL.Query().Get("id")
	id, _ := strconv.Atoi(idStr)
	
	query := "UPDATE users SET deleted_at=NOW() WHERE id=$1"
	_, err := config.DB.Exec(query, id)
	
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	json.NewEncoder(w).Encode(models.Response{Status: 200, Message: "User Deleted"})
}

func GetProfile(w http.ResponseWriter, r *http.Request) {
	// HAPUS enableCors dan OPTIONS check
	
	w.Header().Set("Content-Type", "application/json")
	
	userIDStr := r.URL.Query().Get("user_id")
	userID, _ := strconv.Atoi(userIDStr)
	
	var user models.User
	query := "SELECT id, username, email, full_name, phone, profile_image FROM users WHERE id = $1"
	
	err := config.DB.QueryRow(query, userID).Scan(
		&user.ID, &user.Username, &user.Email, &user.FullName, &user.Phone, &user.ProfileImage,
	)
	
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(models.Response{Status: 404, Message: "User not found"})
		return
	}
	
	json.NewEncoder(w).Encode(models.Response{Status: 200, Data: user})
}

func UpdateProfile(w http.ResponseWriter, r *http.Request) {
	// HAPUS enableCors dan OPTIONS check
	
	w.Header().Set("Content-Type", "application/json")
	
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.Response{Status: 400, Message: "File too big or invalid form"})
		return
	}
	
	userIDStr := r.FormValue("user_id")
	fullName := r.FormValue("full_name")
	phone := r.FormValue("phone")
	
	file, handler, err := r.FormFile("profile_image")
	var imagePath string
	
	if err == nil {
		defer file.Close()
		
		os.MkdirAll("./uploads/profiles", os.ModePerm)
		
		filename := "user_" + userIDStr + "_" + strconv.FormatInt(time.Now().Unix(), 10) + filepath.Ext(handler.Filename)
		imagePath = "/uploads/profiles/" + filename
		
		dst, err := os.Create("." + imagePath)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(models.Response{Status: 500, Message: "Gagal menyimpan gambar"})
			return
		}
		defer dst.Close()
		io.Copy(dst, file)
	}
	
	var query string
	var args []interface{}
	
	if imagePath != "" {
		query = "UPDATE users SET full_name=$1, phone=$2, profile_image=$3, updated_at=NOW() WHERE id=$4"
		args = []interface{}{fullName, phone, imagePath, userIDStr}
	} else {
		query = "UPDATE users SET full_name=$1, phone=$2, updated_at=NOW() WHERE id=$3"
		args = []interface{}{fullName, phone, userIDStr}
	}
	
	_, err = config.DB.Exec(query, args...)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.Response{Status: 500, Message: err.Error()})
		return
	}
	
	var updatedUser models.User
	fetchQuery := "SELECT id, username, email, full_name, phone, profile_image FROM users WHERE id = $1"
	err = config.DB.QueryRow(fetchQuery, userIDStr).Scan(
		&updatedUser.ID,
		&updatedUser.Username,
		&updatedUser.Email,
		&updatedUser.FullName,
		&updatedUser.Phone,
		&updatedUser.ProfileImage,
	)
	
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.Response{Status: 500, Message: "Gagal mengambil data terbaru"})
		return
	}
	
	json.NewEncoder(w).Encode(
		models.Response{
			Status:  200,
			Message: "Profil Berhasil Diupdate",
			Data:    updatedUser,
		},
	)
}
