package main

import (
	"log"
	"net/http"
	
	"backend-wisata/config"
	"backend-wisata/controllers"
)

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin != "" {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, Accept")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			
			next.ServeHTTP(w, r)
		},
	)
}

func main() {
	config.ConnectDB()
	
	mux := http.NewServeMux()
	
	mux.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("./uploads"))))
	
	// Auth
	mux.HandleFunc("/api/login", controllers.Login)
	mux.HandleFunc("/api/logout", controllers.Logout)
	mux.HandleFunc("/api/register", controllers.Register)
	
	// Wisata
	mux.HandleFunc("/api/wisata", controllers.GetAllWisata)
	mux.HandleFunc("/api/wisata/detail", controllers.GetWisataDetail)
	mux.HandleFunc("/api/wisata/create", controllers.CreateWisata)
	mux.HandleFunc("/api/wisata/update", controllers.UpdateWisata)
	mux.HandleFunc("/api/wisata/delete", controllers.DeleteWisata)
	
	// Categories (BARU)
	mux.HandleFunc("/api/categories", controllers.GetAllCategories)
	mux.HandleFunc("/api/categories/create", controllers.CreateCategory)
	mux.HandleFunc("/api/categories/update", controllers.UpdateCategory)
	mux.HandleFunc("/api/categories/delete", controllers.DeleteCategory)
	
	// Booking
	mux.HandleFunc("/api/booking/create", controllers.CreateBooking)
	mux.HandleFunc("/api/booking/history", controllers.GetBookingHistory)
	mux.HandleFunc("/api/booking/detail", controllers.GetBookingDetail)
	mux.HandleFunc("/api/booking/pay", controllers.ProcessPayment)
	mux.HandleFunc("/api/booking/cancel", controllers.CancelBooking)
	
	// Dashboard
	mux.HandleFunc("/api/dashboard/stats", controllers.GetDashboardStats)
	mux.HandleFunc("/api/dashboard/recent-bookings", controllers.GetRecentBookings)
	mux.HandleFunc("/api/dashboard/popular-wisata", controllers.GetPopularWisata)
	
	// User Admin
	mux.HandleFunc("/api/users", controllers.GetAllUsers)
	mux.HandleFunc("/api/users/update", controllers.UpdateUserStatus)
	mux.HandleFunc("/api/users/delete", controllers.DeleteUser)
	
	// Booking Admin
	mux.HandleFunc("/api/bookings", controllers.GetAllBookings)
	
	log.Println("Server running on http://localhost:8080")
	
	if err := http.ListenAndServe("0.0.0.0:8080", corsMiddleware(mux)); err != nil {
		log.Fatal(err)
	}
}
