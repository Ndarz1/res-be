package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"
	
	"backend-wisata/config"
	"backend-wisata/models"
)

func GetAllCategories(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	if r.Method == "OPTIONS" {
		return
	}
	
	query := `
		SELECT
			c.id, c.name, c.slug, c.icon, c.is_active,
			(SELECT COUNT(*) FROM wisata w WHERE w.category_id = c.id AND w.deleted_at IS NULL) as count
		FROM categories c
		ORDER BY c.sort_order ASC, c.id ASC
	`
	
	rows, err := config.DB.Query(query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	
	var categories []map[string]interface{}
	
	for rows.Next() {
		var id, count int
		var name, slug string
		var icon *string
		var isActive bool
		
		if err := rows.Scan(&id, &name, &slug, &icon, &isActive, &count); err != nil {
			continue
		}
		
		iconVal := "grid"
		if icon != nil {
			iconVal = *icon
		}
		
		categories = append(
			categories, map[string]interface{}{
				"id":        id,
				"name":      name,
				"slug":      slug,
				"icon":      iconVal,
				"count":     count,
				"is_active": isActive,
			},
		)
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(
		models.Response{
			Status:  200,
			Message: "Categories Fetched",
			Data:    categories,
		},
	)
}

func CreateCategory(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	var input struct {
		Name     string `json:"name"`
		Slug     string `json:"slug"`
		IsActive bool   `json:"is_active"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}
	
	icon := "grid"
	
	query := "INSERT INTO categories (name, slug, icon, is_active) VALUES ($1, $2, $3, $4)"
	_, err := config.DB.Exec(query, input.Name, input.Slug, icon, input.IsActive)
	
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	json.NewEncoder(w).Encode(models.Response{Status: 201, Message: "Category Created"})
}

func UpdateCategory(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	
	if r.Method != "PUT" && r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	idStr := r.URL.Query().Get("id")
	id, _ := strconv.Atoi(idStr)
	
	var input struct {
		Name     string `json:"name"`
		Slug     string `json:"slug"`
		IsActive bool   `json:"is_active"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}
	
	query := "UPDATE categories SET name=$1, slug=$2, is_active=$3, updated_at=NOW() WHERE id=$4"
	_, err := config.DB.Exec(query, input.Name, input.Slug, input.IsActive, id)
	
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	json.NewEncoder(w).Encode(models.Response{Status: 200, Message: "Category Updated"})
}

func DeleteCategory(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	
	idStr := r.URL.Query().Get("id")
	id, _ := strconv.Atoi(idStr)
	
	var count int
	config.DB.QueryRow("SELECT COUNT(*) FROM wisata WHERE category_id = $1 AND deleted_at IS NULL", id).Scan(&count)
	
	if count > 0 {
		http.Error(w, "Kategori sedang digunakan oleh wisata aktif", http.StatusBadRequest)
		return
	}
	
	_, err := config.DB.Exec("DELETE FROM categories WHERE id = $1", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	json.NewEncoder(w).Encode(models.Response{Status: 200, Message: "Category Deleted"})
}
