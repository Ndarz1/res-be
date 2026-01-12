package controllers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"
	
	"backend-wisata/config"
	"backend-wisata/models"
)

func CreateBooking(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		responseError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	
	var input struct {
		WisataID      int    `json:"wisata_id"`
		UserID        int    `json:"user_id"`
		VisitDate     string `json:"visit_date"`
		Quantity      int    `json:"quantity"`
		PaymentMethod string `json:"payment_method"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		responseError(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}
	
	var hargaTiket float64
	err := config.DB.QueryRow("SELECT harga_tiket FROM wisata WHERE id = $1", input.WisataID).Scan(&hargaTiket)
	if err != nil {
		log.Println("ERROR FETCH WISATA:", err)
		responseError(w, http.StatusNotFound, "Wisata tidak ditemukan")
		return
	}
	
	finalPrice := hargaTiket * float64(input.Quantity)
	
	query := `
   INSERT INTO bookings (wisata_id, user_id, visit_date, quantity, total_price, final_price, status, payment_method, created_at)
   VALUES ($1, $2, $3, $4, $5, $6, 'pending', $7, NOW())
   RETURNING id, booking_code
  `
	
	var newBookingID int
	var newBookingCode string
	
	err = config.DB.QueryRow(
		query,
		input.WisataID,
		input.UserID,
		input.VisitDate,
		input.Quantity,
		finalPrice,
		finalPrice,
		input.PaymentMethod,
	).Scan(&newBookingID, &newBookingCode)
	
	if err != nil {
		log.Println("ERROR DATABASE:", err)
		responseError(w, http.StatusInternalServerError, "Gagal menyimpan booking: "+err.Error())
		return
	}
	
	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(
		models.Response{
			Status:  201,
			Message: "Booking Berhasil Dibuat",
			Data: map[string]interface{}{
				"booking_id":   newBookingID,
				"booking_code": newBookingCode,
				"final_price":  finalPrice,
			},
		},
	)
}

func GetBookingHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		responseError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		responseError(w, http.StatusBadRequest, "User ID required")
		return
	}
	
	query := `
   SELECT
    b.id, b.booking_code, b.wisata_id, w.nama_tempat,
    b.user_id, b.visit_date, b.quantity,
    b.total_price, b.final_price, b.status, b.payment_method, b.created_at
   FROM bookings b
   JOIN wisata w ON b.wisata_id = w.id
   WHERE b.user_id = $1
   ORDER BY b.created_at DESC
  `
	
	rows, err := config.DB.Query(query, userID)
	if err != nil {
		log.Println("ERROR FETCH HISTORY:", err)
		responseError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	
	var history []models.Booking
	for rows.Next() {
		var b models.Booking
		var visitDate time.Time
		
		err := rows.Scan(
			&b.ID, &b.BookingCode, &b.WisataID, &b.WisataNama,
			&b.UserID, &visitDate, &b.Quantity,
			&b.TotalPrice, &b.FinalPrice, &b.Status, &b.PaymentMethod, &b.CreatedAt,
		)
		if err != nil {
			log.Println("SCAN ERROR:", err)
			continue
		}
		
		b.VisitDate = visitDate.Format("2006-01-02")
		history = append(history, b)
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(
		models.Response{
			Status:  200,
			Message: "Booking History Fetched",
			Data:    history,
		},
	)
}

func GetBookingDetail(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	code := r.URL.Query().Get("code")
	
	query := `
		SELECT b.id, b.booking_code, b.wisata_id, w.nama_tempat,
		       b.visit_date, b.quantity, b.final_price, b.status
		FROM bookings b
		JOIN wisata w ON b.wisata_id = w.id
		WHERE b.booking_code = $1
	`
	
	var b models.Booking
	var visitDateRaw time.Time
	
	err := config.DB.QueryRow(query, code).Scan(
		&b.ID, &b.BookingCode, &b.WisataID, &b.WisataNama,
		&visitDateRaw, &b.Quantity, &b.FinalPrice, &b.Status,
	)
	
	if err == sql.ErrNoRows {
		http.Error(w, "Booking not found", http.StatusNotFound)
		return
	}
	
	b.VisitDate = visitDateRaw.Format("2006-01-02")
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(
		models.Response{
			Status:  200,
			Message: "Booking Detail",
			Data:    b,
		},
	)
}

func ProcessPayment(w http.ResponseWriter, r *http.Request) {
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
		BookingCode string `json:"booking_code"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}
	
	query := "UPDATE bookings SET status = 'paid', updated_at = NOW() WHERE booking_code = $1"
	res, err := config.DB.Exec(query, input.BookingCode)
	
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Booking code not found", http.StatusNotFound)
		return
	}
	
	json.NewEncoder(w).Encode(
		models.Response{
			Status:  200,
			Message: "Payment Success",
		},
	)
}

func GetAllBookings(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	
	query := `
		SELECT
			b.id, b.booking_code,
			u.full_name as user_name,
			w.nama_tempat as wisata_nama,
			b.visit_date, b.quantity, b.final_price, b.status, b.payment_method
		FROM bookings b
		JOIN users u ON b.user_id = u.id
		JOIN wisata w ON b.wisata_id = w.id
		ORDER BY b.created_at DESC
	`
	
	rows, err := config.DB.Query(query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	
	var bookings []map[string]interface{}
	
	for rows.Next() {
		var id, quantity int
		var code, userName, wisataNama, status, method string
		var finalPrice float64
		var visitDate time.Time
		
		if err := rows.Scan(
			&id,
			&code,
			&userName,
			&wisataNama,
			&visitDate,
			&quantity,
			&finalPrice,
			&status,
			&method,
		); err != nil {
			continue
		}
		
		bookings = append(
			bookings, map[string]interface{}{
				"id":             id,
				"booking_code":   code,
				"user_name":      userName,
				"wisata_nama":    wisataNama,
				"visit_date":     visitDate.Format("2006-01-02"),
				"quantity":       quantity,
				"total_price":    finalPrice,
				"status":         status,
				"payment_method": method,
			},
		)
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(
		models.Response{
			Status:  200,
			Message: "All Bookings Fetched",
			Data:    bookings,
		},
	)
}

func CancelBooking(w http.ResponseWriter, r *http.Request) {
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
		BookingCode string `json:"booking_code"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}
	
	var currentStatus string
	err := config.DB.QueryRow(
		"SELECT status FROM bookings WHERE booking_code = $1",
		input.BookingCode,
	).Scan(&currentStatus)
	
	if err != nil {
		http.Error(w, "Booking tidak ditemukan", http.StatusNotFound)
		return
	}
	
	if currentStatus != "pending" {
		http.Error(w, "Hanya pesanan pending yang bisa dibatalkan", http.StatusBadRequest)
		return
	}
	
	query := "UPDATE bookings SET status = 'cancelled', updated_at = NOW() WHERE booking_code = $1"
	_, err = config.DB.Exec(query, input.BookingCode)
	
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	json.NewEncoder(w).Encode(
		models.Response{
			Status:  200,
			Message: "Booking Cancelled",
		},
	)
}
