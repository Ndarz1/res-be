package controllers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	
	"backend-wisata/config"
	"backend-wisata/models"
	
	"golang.org/x/crypto/bcrypt"
)

func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func responseError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(
		models.Response{
			Status:  code,
			Message: message,
		},
	)
}

func Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		responseError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	
	var input models.LoginInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		responseError(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}
	
	var user models.User
	var passwordHash string
	
	query := `
		SELECT id, uuid, username, email, full_name, phone, role, is_active, password_hash
		FROM users
		WHERE (username = $1 OR email = $1)
		AND deleted_at IS NULL
	`
	
	err := config.DB.QueryRow(query, input.Username).Scan(
		&user.ID, &user.UUID, &user.Username, &user.Email,
		&user.FullName, &user.Phone, &user.Role, &user.IsActive, &passwordHash,
	)
	
	if err == sql.ErrNoRows {
		responseError(w, http.StatusUnauthorized, "Akun tidak ditemukan")
		return
	} else if err != nil {
		responseError(w, http.StatusInternalServerError, err.Error())
		return
	}
	
	if !user.IsActive {
		responseError(w, http.StatusForbidden, "Akun nonaktif")
		return
	}
	
	if !checkPasswordHash(input.Password, passwordHash) {
		responseError(w, http.StatusUnauthorized, "Password salah")
		return
	}
	
	_, _ = config.DB.Exec("UPDATE users SET last_login = NOW() WHERE id = $1", user.ID)
	
	if user.Role == "admin" || user.Role == "superadmin" {
		session, _ := config.AdminStore.Get(r, "admin-session-token")
		session.Values["user_id"] = user.ID
		session.Values["authenticated"] = true
		session.Save(r, w)
	} else {
		session, _ := config.UserStore.Get(r, "user-session-token")
		session.Values["user_id"] = user.ID
		session.Values["authenticated"] = true
		session.Save(r, w)
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(
		models.Response{
			Status:  200,
			Message: "Login Berhasil",
			Data:    user,
		},
	)
}

func Logout(w http.ResponseWriter, r *http.Request) {
	adminSession, _ := config.AdminStore.Get(r, "admin-session-token")
	adminSession.Values["authenticated"] = false
	adminSession.Options.MaxAge = -1
	adminSession.Save(r, w)
	
	userSession, _ := config.UserStore.Get(r, "user-session-token")
	userSession.Values["authenticated"] = false
	userSession.Options.MaxAge = -1
	userSession.Save(r, w)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(
		models.Response{
			Status:  200,
			Message: "Logout Berhasil",
		},
	)
}

func GetMe(w http.ResponseWriter, r *http.Request) {
	var userID interface{}
	var authenticated bool
	
	adminSession, _ := config.AdminStore.Get(r, "admin-session-token")
	if auth, ok := adminSession.Values["authenticated"].(bool); ok && auth {
		userID = adminSession.Values["user_id"]
		authenticated = true
	} else {
		userSession, _ := config.UserStore.Get(r, "user-session-token")
		if auth, ok := userSession.Values["authenticated"].(bool); ok && auth {
			userID = userSession.Values["user_id"]
			authenticated = true
		}
	}
	
	if !authenticated {
		responseError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	
	var user models.User
	query := `
		SELECT id, uuid, username, email, full_name, phone, role, is_active,
		COALESCE(profile_image, '') as profile_image
		FROM users WHERE id = $1
	`
	
	err := config.DB.QueryRow(query, userID).Scan(
		&user.ID, &user.UUID, &user.Username, &user.Email,
		&user.FullName, &user.Phone, &user.Role, &user.IsActive, &user.ProfileImage,
	)
	
	if err != nil {
		responseError(w, http.StatusNotFound, "User not found")
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(
		models.Response{
			Status:  200,
			Message: "Success",
			Data:    user,
		},
	)
}

func Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		responseError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	
	var input models.RegisterInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		responseError(w, http.StatusBadRequest, "Invalid JSON data")
		return
	}
	
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		responseError(w, http.StatusInternalServerError, "Gagal enkripsi password")
		return
	}
	
	var phone *string
	if input.Phone != "" {
		phone = &input.Phone
	}
	
	query := `
		INSERT INTO users (username, email, password_hash, full_name, phone, role, is_active)
		VALUES ($1, $2, $3, $4, $5, 'user', TRUE)
	`
	
	_, err = config.DB.Exec(query, input.Username, input.Email, string(hashedPassword), input.FullName, phone)
	
	if err != nil {
		responseError(w, http.StatusConflict, "Username atau Email sudah terdaftar")
		return
	}
	
	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(
		models.Response{
			Status:  201,
			Message: "Registrasi Berhasil",
		},
	)
}
