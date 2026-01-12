package controllers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
	
	"backend-wisata/config"
	"backend-wisata/models"
)

func GetDashboardStats(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	
	query := `
   SELECT
    total_wisata, active_wisata, total_users, total_visitors,
    total_revenue, total_bookings, average_rating
   FROM vw_dashboard_stats
  `
	
	var stats models.DashboardStats
	
	err := config.DB.QueryRow(query).Scan(
		&stats.TotalWisata,
		&stats.ActiveWisata,
		&stats.TotalUsers,
		&stats.TotalVisitors,
		&stats.TotalRevenue,
		&stats.TotalBookings,
		&stats.AverageRating,
	)
	
	if err != nil {
		log.Println("ERROR DASHBOARD STATS:", err)
		responseError(w, http.StatusInternalServerError, "Gagal mengambil data statistik")
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(
		models.Response{
			Status:  200,
			Message: "Dashboard Stats Fetched",
			Data:    stats,
		},
	)
}

func GetRecentBookings(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	
	query := `
   SELECT
    b.id, b.booking_code, b.wisata_id, w.nama_tempat,
    b.user_id, b.visit_date, b.quantity,
    b.total_price, b.final_price, b.status, b.payment_method, b.created_at
   FROM bookings b
   JOIN wisata w ON b.wisata_id = w.id
   ORDER BY b.created_at DESC
   LIMIT 5
  `
	
	rows, err := config.DB.Query(query)
	if err != nil {
		log.Println("ERROR RECENT BOOKINGS:", err)
		responseError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	
	var bookings []models.Booking
	
	for rows.Next() {
		var b models.Booking
		var visitDateRaw time.Time
		var createdAtRaw time.Time
		
		err := rows.Scan(
			&b.ID, &b.BookingCode, &b.WisataID, &b.WisataNama,
			&b.UserID, &visitDateRaw, &b.Quantity,
			&b.TotalPrice, &b.FinalPrice, &b.Status, &b.PaymentMethod, &createdAtRaw,
		)
		
		if err != nil {
			log.Println("SCAN ERROR:", err)
			continue
		}
		
		b.VisitDate = visitDateRaw.Format("2006-01-02")
		b.CreatedAt = createdAtRaw
		bookings = append(bookings, b)
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(
		models.Response{
			Status:  200,
			Message: "Recent Bookings Fetched",
			Data:    bookings,
		},
	)
}

func GetPopularWisata(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	
	var totalAllBookings int
	err := config.DB.QueryRow("SELECT COUNT(*) FROM bookings").Scan(&totalAllBookings)
	if err != nil {
		totalAllBookings = 1
	}
	if totalAllBookings == 0 {
		totalAllBookings = 1
	}
	
	query := `
   SELECT
    w.id, w.nama_tempat, COALESCE(w.deskripsi, ''), w.lokasi, w.harga_tiket,
    COALESCE((SELECT image_url FROM wisata_images WHERE wisata_id = w.id AND is_primary = true LIMIT 1), '') as image_url,
    COUNT(b.id) as total_visits
   FROM wisata w
   LEFT JOIN bookings b ON w.id = b.wisata_id
   WHERE w.deleted_at IS NULL
   GROUP BY w.id, w.nama_tempat, w.deskripsi, w.lokasi, w.harga_tiket
   ORDER BY total_visits DESC
   LIMIT 5
  `
	
	rows, err := config.DB.Query(query)
	if err != nil {
		log.Println("ERROR POPULAR:", err)
		responseError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	
	var popularList []models.PopularWisata
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	baseURL := scheme + "://" + r.Host
	
	for rows.Next() {
		var p models.PopularWisata
		var imgURL string
		
		err := rows.Scan(
			&p.ID, &p.NamaTempat, &p.Deskripsi, &p.Lokasi, &p.HargaTiket,
			&imgURL, &p.TotalVisits,
		)
		if err != nil {
			continue
		}
		
		if imgURL != "" {
			p.ImageURL = baseURL + imgURL
		}
		
		p.Percentage = (float64(p.TotalVisits) / float64(totalAllBookings)) * 100
		
		if p.TotalVisits > 20 {
			p.Label = "Trending"
		} else if p.TotalVisits > 10 {
			p.Label = "Populer"
		} else {
			p.Label = "Rekomendasi"
		}
		
		popularList = append(popularList, p)
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(
		models.Response{
			Status:  200,
			Message: "Popular Wisata Fetched",
			Data:    popularList,
		},
	)
}
