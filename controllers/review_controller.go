package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"
	
	"backend-wisata/config"
	"backend-wisata/models"
)

func SubmitReview(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(models.Response{Status: 405, Message: "Method not allowed"})
		return
	}
	
	var input struct {
		WisataID int    `json:"wisata_id"`
		UserID   int    `json:"user_id"`
		Rating   int    `json:"rating"`
		Comment  string `json:"comment"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.Response{Status: 400, Message: "Invalid body"})
		return
	}
	
	var hasVisited bool
	checkQuery := `
		SELECT EXISTS(
			SELECT 1 FROM bookings
			WHERE user_id = $1 AND wisata_id = $2 AND status IN ('paid', 'completed')
		)`
	config.DB.QueryRow(checkQuery, input.UserID, input.WisataID).Scan(&hasVisited)
	
	if !hasVisited {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(models.Response{Status: 403, Message: "Anda harus berkunjung sebelum memberi ulasan"})
		return
	}
	
	query := `
		INSERT INTO reviews (wisata_id, user_id, rating, comment, is_approved)
		VALUES ($1, $2, $3, $4, FALSE)
	`
	_, err := config.DB.Exec(query, input.WisataID, input.UserID, input.Rating, input.Comment)
	
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.Response{Status: 500, Message: "Gagal menyimpan ulasan"})
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(
		models.Response{
			Status:  201,
			Message: "Ulasan terkirim dan menunggu moderasi admin",
		},
	)
}

func GetReviews(w http.ResponseWriter, r *http.Request) {
	wisataID := r.URL.Query().Get("wisata_id")
	
	query := `
		SELECT r.id, r.wisata_id, r.user_id, r.rating, r.comment, r.created_at,
			u.full_name, u.profile_image
		FROM reviews r
		JOIN users u ON r.user_id = u.id
		WHERE r.wisata_id = $1 AND r.is_approved = TRUE
		ORDER BY r.created_at DESC
	`
	
	rows, err := config.DB.Query(query, wisataID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.Response{Status: 500, Message: err.Error()})
		return
	}
	defer rows.Close()
	
	var reviews []models.Review
	for rows.Next() {
		var r models.Review
		if err := rows.Scan(
			&r.ID,
			&r.WisataID,
			&r.UserID,
			&r.Rating,
			&r.Comment,
			&r.CreatedAt,
			&r.UserName,
			&r.UserProfile,
		); err != nil {
			continue
		}
		reviews = append(reviews, r)
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(
		models.Response{
			Status: 200,
			Data:   reviews,
		},
	)
}

func GetAdminReviews(w http.ResponseWriter, r *http.Request) {
	query := `
		SELECT r.id, r.rating, r.comment, r.is_approved, r.created_at,
			u.full_name, w.nama_tempat
		FROM reviews r
		JOIN users u ON r.user_id = u.id
		JOIN wisata w ON r.wisata_id = w.id
		ORDER BY r.created_at DESC
	`
	
	rows, err := config.DB.Query(query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	
	var reviews []map[string]interface{}
	for rows.Next() {
		var id, rating int
		var comment, userName, wisataName string
		var isApproved bool
		var createdAt string
		
		if err := rows.Scan(&id, &rating, &comment, &isApproved, &createdAt, &userName, &wisataName); err != nil {
			continue
		}
		
		status := "pending"
		if isApproved {
			status = "approved"
		}
		
		reviews = append(
			reviews, map[string]interface{}{
				"id":      id,
				"rating":  rating,
				"comment": comment,
				"status":  status,
				"user":    userName,
				"wisata":  wisataName,
				"date":    createdAt,
			},
		)
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(
		models.Response{
			Status: 200,
			Data:   reviews,
		},
	)
}

func ApproveReview(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	idStr := r.URL.Query().Get("id")
	id, _ := strconv.Atoi(idStr)
	
	_, err := config.DB.Exec("UPDATE reviews SET is_approved = TRUE WHERE id = $1", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	json.NewEncoder(w).Encode(models.Response{Status: 200, Message: "Review Approved"})
}

func DeleteReview(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" && r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	idStr := r.URL.Query().Get("id")
	id, _ := strconv.Atoi(idStr)
	
	_, err := config.DB.Exec("DELETE FROM reviews WHERE id = $1", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	json.NewEncoder(w).Encode(models.Response{Status: 200, Message: "Review Deleted"})
}
