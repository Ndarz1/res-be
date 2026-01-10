package main

import (
	"log"
	"net/http"
	
	"backend-wisata/config"
	"backend-wisata/controllers"
)

func main() {
	config.ConnectDB()
	
	http.HandleFunc("/api/login", controllers.Login)
	http.HandleFunc("/api/logout", controllers.Logout)
	
	http.HandleFunc("/api/wisata", controllers.GetAllWisata)
	http.HandleFunc("/api/wisata/detail", controllers.GetWisataDetail)
	http.HandleFunc("/api/wisata/create", controllers.CreateWisata)
	http.HandleFunc("/api/wisata/update", controllers.UpdateWisata)
	http.HandleFunc("/api/wisata/delete", controllers.DeleteWisata)
	
	log.Println("Server running on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
