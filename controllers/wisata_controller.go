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

func saveUploadedFile(file io.Reader, filename string) (string, error) {
	uploadDir := "./uploads"
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		os.Mkdir(uploadDir, 0755)
	}
	
	uniqueName := strconv.FormatInt(time.Now().UnixNano(), 10) + filepath.Ext(filename)
	filePath := filepath.Join(uploadDir, uniqueName)
	
	out, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer out.Close()
	
	_, err = io.Copy(out, file)
	if err != nil {
		return "", err
	}
	
	return "/uploads/" + uniqueName, nil
}

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
    w.harga_tiket, w.rating_total, w.category_id,
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
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	baseURL := scheme + "://" + r.Host
	
	for rows.Next() {
		var w models.Wisata
		if err := rows.Scan(
			&w.ID, &w.UUID, &w.NamaTempat, &w.Slug, &w.Lokasi,
			&w.HargaTiket, &w.RatingTotal, &w.CategoryID, &w.ImageURL,
			&w.CategoryName,
		); err != nil {
			continue
		}
		
		if w.ImageURL != "" {
			w.ImageURL = baseURL + w.ImageURL
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
    w.harga_tiket, w.lokasi, w.category_id,
    c.name as category_name,
    COALESCE((SELECT image_url FROM wisata_images WHERE wisata_id = w.id AND is_primary = true LIMIT 1), '') as image_url
   FROM wisata w
   JOIN categories c ON w.category_id = c.id
   WHERE w.id = $1 AND w.deleted_at IS NULL
  `
	
	var data models.Wisata
	err := config.DB.QueryRow(query, id).Scan(
		&data.ID, &data.UUID, &data.NamaTempat, &data.Deskripsi, &data.Fasilitas,
		&data.HargaTiket, &data.Lokasi, &data.CategoryID,
		&data.CategoryName, &data.ImageURL,
	)
	
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	
	if data.ImageURL != "" {
		scheme := "http"
		if r.TLS != nil {
			scheme = "https"
		}
		data.ImageURL = scheme + "://" + r.Host + data.ImageURL
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
	
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "File too large", http.StatusBadRequest)
		return
	}
	
	namaTempat := r.FormValue("nama_tempat")
	slug := r.FormValue("slug")
	categoryID := r.FormValue("category_id")
	lokasi := r.FormValue("lokasi")
	hargaTiket := r.FormValue("harga_tiket")
	deskripsi := r.FormValue("deskripsi")
	fasilitas := r.FormValue("fasilitas")
	
	query := `
   INSERT INTO wisata (nama_tempat, slug, category_id, lokasi, harga_tiket, deskripsi, fasilitas)
   VALUES ($1, $2, $3, $4, $5, $6, $7)
   RETURNING id
  `
	
	var newID int
	err = config.DB.QueryRow(
		query,
		namaTempat, slug, categoryID, lokasi, hargaTiket, deskripsi, fasilitas,
	).Scan(&newID)
	
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	file, header, err := r.FormFile("image")
	if err == nil {
		defer file.Close()
		imagePath, err := saveUploadedFile(file, header.Filename)
		if err == nil {
			_, _ = config.DB.Exec(
				"INSERT INTO wisata_images (wisata_id, image_url, is_primary) VALUES ($1, $2, true)",
				newID,
				imagePath,
			)
		}
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
	
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "File too large", http.StatusBadRequest)
		return
	}
	
	namaTempat := r.FormValue("nama_tempat")
	categoryID := r.FormValue("category_id")
	lokasi := r.FormValue("lokasi")
	hargaTiket := r.FormValue("harga_tiket")
	deskripsi := r.FormValue("deskripsi")
	fasilitas := r.FormValue("fasilitas")
	
	query := `
   UPDATE wisata
   SET nama_tempat=$1, category_id=$2, lokasi=$3, harga_tiket=$4, deskripsi=$5, fasilitas=$6, updated_at=NOW()
   WHERE id=$7
  `
	_, err = config.DB.Exec(
		query,
		namaTempat, categoryID, lokasi, hargaTiket, deskripsi, fasilitas, id,
	)
	
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	file, header, err := r.FormFile("image")
	if err == nil {
		defer file.Close()
		imagePath, err := saveUploadedFile(file, header.Filename)
		if err == nil {
			_, _ = config.DB.Exec("DELETE FROM wisata_images WHERE wisata_id = $1", id)
			_, _ = config.DB.Exec(
				"INSERT INTO wisata_images (wisata_id, image_url, is_primary) VALUES ($1, $2, true)",
				id,
				imagePath,
			)
		}
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
