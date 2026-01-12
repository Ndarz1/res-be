package models

type DashboardStats struct {
	TotalWisata   int     `json:"total_wisata"`
	ActiveWisata  int     `json:"active_wisata"`
	TotalUsers    int     `json:"total_users"`
	TotalVisitors int     `json:"total_visitors"`
	TotalRevenue  float64 `json:"total_revenue"`
	TotalBookings int     `json:"total_bookings"`
	AverageRating float64 `json:"average_rating"`
}

type PopularWisata struct {
	ID          int     `json:"id"`
	NamaTempat  string  `json:"nama_tempat"`
	Deskripsi   string  `json:"deskripsi"`
	Lokasi      string  `json:"lokasi"`
	HargaTiket  float64 `json:"harga_tiket"`
	ImageURL    string  `json:"image_url"`
	TotalVisits int     `json:"total_visits"`
	Label       string  `json:"label"`
	Percentage  float64 `json:"percentage"`
}
