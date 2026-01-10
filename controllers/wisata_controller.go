package controllers

import (
	"encoding/json"
	"net/http"
	
	"backend-wisata/config"
	"backend-wisata/models"
)

func GetAllWisata(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	
	searchQuery := r.URL.Query().Get("q")
	
	query := `
   SELECT
    w.id, w.uuid, w.nama_tempat, w.slug, w.lokasi,
    w.harga_tiket, w.rating_total,
    COALESCE((SELECT image_url FROM wisata_images WHERE wisata_id = w.id AND is_primary = true LIMIT 1), '') as image_url,
    c.name as category_name
   FROM wisata w
   JOIN categories c ON w.category_id = c.id
   WHERE w.deleted_at IS NULL
  `
	
	var args []interface{}
	if searchQuery != "" {
		query += " AND w.nama_tempat ILIKE '%' || $1 || '%'"
		args = append(args, searchQuery)
	}
	
	query += " ORDER BY w.created_at DESC"
	
	rows, err := config.DB.Query(query, args...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	
	var listWisata []models.Wisata
	for rows.Next() {
		var w models.Wisata
		if err := rows.Scan(
			&w.ID, &w.UUID, &w.NamaTempat, &w.Slug, &w.Lokasi,
			&w.HargaTiket, &w.RatingTotal, &w.ImageURL,
			&w.CategoryName,
		); err != nil {
			continue
		}
		listWisata = append(listWisata, w)
	}
	
	json.NewEncoder(w).Encode(
		models.Response{
			Status:  200,
			Message: "Success fetch data",
			Data:    listWisata,
		},
	)
}

func GetWisataDetail(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	if r.Method == "OPTIONS" {
		return
	}
	
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "ID required", http.StatusBadRequest)
		return
	}
	
	query := `
   SELECT
    w.id, w.uuid, w.nama_tempat, w.deskripsi, w.fasilitas,
    w.harga_tiket,
    c.name as category_name
   FROM wisata w
   JOIN categories c ON w.category_id = c.id
   WHERE w.id = $1 AND w.deleted_at IS NULL
  `
	
	var data models.Wisata
	err := config.DB.QueryRow(query, id).Scan(
		&data.ID, &data.UUID, &data.NamaTempat, &data.Deskripsi, &data.Fasilitas,
		&data.HargaTiket,
		&data.CategoryName,
	)
	
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	
	json.NewEncoder(w).Encode(
		models.Response{
			Status:  200,
			Message: "Detail Wisata",
			Data:    data,
		},
	)
}

func CreateWisata(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	var input models.Wisata
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	query := `
   INSERT INTO wisata (nama_tempat, slug, category_id, lokasi, harga_tiket, deskripsi)
   VALUES ($1, $2, $3, $4, $5, $6)
   RETURNING id
  `
	
	var newID int
	err := config.DB.QueryRow(
		query,
		input.NamaTempat,
		input.Slug,
		input.CategoryID,
		input.Lokasi,
		input.HargaTiket,
		input.Deskripsi,
	).Scan(&newID)
	
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	json.NewEncoder(w).Encode(
		models.Response{
			Status:  201,
			Message: "Wisata Created",
			Data:    map[string]int{"id": newID},
		},
	)
}

func UpdateWisata(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	
	if r.Method != "PUT" && r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "ID required", http.StatusBadRequest)
		return
	}
	
	var input models.Wisata
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	query := `
   UPDATE wisata
   SET nama_tempat=$1, category_id=$2, lokasi=$3, harga_tiket=$4, deskripsi=$5, updated_at=NOW()
   WHERE id=$6
  `
	_, err := config.DB.Exec(
		query,
		input.NamaTempat, input.CategoryID, input.Lokasi,
		input.HargaTiket, input.Deskripsi, id,
	)
	
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	json.NewEncoder(w).Encode(
		models.Response{
			Status:  200,
			Message: "Wisata Updated",
		},
	)
}

func DeleteWisata(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "ID required", http.StatusBadRequest)
		return
	}
	
	query := "UPDATE wisata SET deleted_at = NOW() WHERE id = $1"
	_, err := config.DB.Exec(query, id)
	
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	json.NewEncoder(w).Encode(
		models.Response{
			Status:  200,
			Message: "Wisata Deleted",
		},
	)
}
