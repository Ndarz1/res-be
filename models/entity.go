package models

import (
	"time"
)

type Category struct {
	ID        int     `json:"id"`
	Name      string  `json:"name"`
	Slug      string  `json:"slug"`
	Icon      *string `json:"icon"`
	SortOrder int     `json:"sort_order"`
	IsActive  bool    `json:"is_active"`
}

type Wisata struct {
	ID            int       `json:"id"`
	UUID          string    `json:"uuid"`
	CategoryID    int       `json:"category_id"`
	CategoryName  string    `json:"category_name,omitempty"`
	NamaTempat    string    `json:"nama_tempat"`
	Slug          string    `json:"slug"`
	Lokasi        string    `json:"lokasi"`
	Latitude      *float64  `json:"latitude"`
	Longitude     *float64  `json:"longitude"`
	AlamatLengkap *string   `json:"alamat_lengkap"`
	Deskripsi     *string   `json:"deskripsi"`
	Fasilitas     *string   `json:"fasilitas"`
	HargaTiket    float64   `json:"harga_tiket"`
	RatingTotal   float64   `json:"rating_total"`
	TotalReviews  int       `json:"total_reviews"`
	ImageURL      string    `json:"image_url"`
	CreatedAt     time.Time `json:"created_at"`
}

type Booking struct {
	ID            int       `json:"id"`
	BookingCode   string    `json:"booking_code"`
	WisataID      int       `json:"wisata_id"`
	UserID        int       `json:"user_id"`
	VisitDate     string    `json:"visit_date"`
	WisataNama    string    `json:"wisata_nama,omitempty"`
	Quantity      int       `json:"quantity"`
	TotalPrice    float64   `json:"total_price"`
	FinalPrice    float64   `json:"final_price"`
	Status        string    `json:"status"`
	PaymentMethod string    `json:"payment_method"`
	CreatedAt     time.Time `json:"created_at"`
}

type Response struct {
	Status  int         `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}
