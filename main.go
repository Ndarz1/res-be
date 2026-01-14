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
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}
			
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, Accept")
			
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			
			next.ServeHTTP(w, r)
		},
	)
}

func main() {
	config.ConnectDB()
	config.InitSession()
	
	mux := http.NewServeMux()
	
	fileServer := http.FileServer(http.Dir("./uploads"))
	mux.Handle("/uploads/", http.StripPrefix("/uploads/", fileServer))
	
	mux.HandleFunc("/api/login", controllers.Login)
	mux.HandleFunc("/api/logout", controllers.Logout)
	mux.HandleFunc("/api/register", controllers.Register)
	mux.HandleFunc("/api/me", controllers.GetMe)
	mux.HandleFunc("/api/profile", controllers.GetProfile)
	mux.HandleFunc("/api/profile/update", controllers.UpdateProfile)
	
	mux.HandleFunc("/api/wisata", controllers.GetAllWisata)
	mux.HandleFunc("/api/wisata/detail", controllers.GetWisataDetail)
	mux.HandleFunc("/api/wisata/create", controllers.CreateWisata)
	mux.HandleFunc("/api/wisata/update", controllers.UpdateWisata)
	mux.HandleFunc("/api/wisata/delete", controllers.DeleteWisata)
	
	mux.HandleFunc("/api/categories", controllers.GetAllCategories)
	mux.HandleFunc("/api/categories/create", controllers.CreateCategory)
	mux.HandleFunc("/api/categories/update", controllers.UpdateCategory)
	mux.HandleFunc("/api/categories/delete", controllers.DeleteCategory)
	
	mux.HandleFunc("/api/booking/create", controllers.CreateBooking)
	mux.HandleFunc("/api/booking/history", controllers.GetBookingHistory)
	mux.HandleFunc("/api/booking/detail", controllers.GetBookingDetail)
	mux.HandleFunc("/api/booking/pay", controllers.ProcessPayment)
	mux.HandleFunc("/api/booking/cancel", controllers.CancelBooking)
	
	mux.HandleFunc("/api/dashboard/stats", controllers.GetDashboardStats)
	mux.HandleFunc("/api/dashboard/recent-bookings", controllers.GetRecentBookings)
	mux.HandleFunc("/api/dashboard/popular-wisata", controllers.GetPopularWisata)
	
	mux.HandleFunc("/api/users", controllers.GetAllUsers)
	mux.HandleFunc("/api/users/update", controllers.UpdateUserStatus)
	mux.HandleFunc("/api/users/delete", controllers.DeleteUser)
	
	mux.HandleFunc("/api/reviews/submit", controllers.SubmitReview)
	mux.HandleFunc("/api/reviews/list", controllers.GetReviews)
	
	mux.HandleFunc("/api/admin/reviews", controllers.GetAdminReviews)
	mux.HandleFunc("/api/admin/reviews/approve", controllers.ApproveReview)
	mux.HandleFunc("/api/admin/reviews/delete", controllers.DeleteReview)
	mux.HandleFunc("/api/bookings", controllers.GetAllBookings)
	
	mux.HandleFunc("/api/blog/posts", controllers.GetBlogPosts)
	mux.HandleFunc("/api/blog/detail", controllers.GetBlogDetail)
	mux.HandleFunc("/api/blog/categories", controllers.GetBlogCategories)
	mux.HandleFunc("/api/blog/create", controllers.CreateBlogPost)
	mux.HandleFunc("/api/blog/update", controllers.UpdateBlogPost)
	mux.HandleFunc("/api/blog/delete", controllers.DeleteBlogPost)
	
	log.Println("Server running on http://localhost:8080")
	
	if err := http.ListenAndServe("0.0.0.0:8080", corsMiddleware(mux)); err != nil {
		log.Fatal(err)
	}
}
