package controllers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"
	
	"backend-wisata/config"
	"backend-wisata/models"
)

func GetBlogPosts(w http.ResponseWriter, r *http.Request) {
	categorySlug := r.URL.Query().Get("category")
	search := r.URL.Query().Get("q")
	
	query := `
			SELECT
				b.id, b.title, b.slug, b.excerpt, b.thumbnail,
				b.published_at, u.full_name, c.name, c.id
			FROM blog_posts b
			LEFT JOIN users u ON b.author_id = u.id
			LEFT JOIN blog_categories c ON b.blog_category_id = c.id
			WHERE b.status = 'published'
		`
	
	var args []interface{}
	counter := 1
	
	if categorySlug != "" {
		query += " AND c.slug = $" + strconv.Itoa(counter)
		args = append(args, categorySlug)
		counter++
	}
	
	if search != "" {
		query += " AND b.title ILIKE '%' || $" + strconv.Itoa(counter) + " || '%'"
		args = append(args, search)
		counter++
	}
	
	query += " ORDER BY b.published_at DESC"
	
	rows, err := config.DB.Query(query, args...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	
	var posts []models.BlogPost
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	baseURL := scheme + "://" + r.Host
	
	for rows.Next() {
		var p models.BlogPost
		var pubDate time.Time
		var thumb *string
		
		if err := rows.Scan(
			&p.ID, &p.Title, &p.Slug, &p.Excerpt, &thumb,
			&pubDate, &p.AuthorName, &p.CategoryName, &p.CategoryID,
		); err != nil {
			continue
		}
		
		if thumb != nil {
			p.Thumbnail = baseURL + *thumb
		}
		p.PublishedAt = pubDate.Format("02 Jan 2006")
		posts = append(posts, p)
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(
		models.Response{
			Status:  200,
			Message: "Blog posts fetched",
			Data:    posts,
		},
	)
}

func GetBlogDetail(w http.ResponseWriter, r *http.Request) {
	slug := r.URL.Query().Get("slug")
	
	query := `
			SELECT
				b.id, b.title, b.slug, b.content, b.thumbnail,
				b.published_at, u.full_name, c.name, c.id
			FROM blog_posts b
			LEFT JOIN users u ON b.author_id = u.id
			LEFT JOIN blog_categories c ON b.blog_category_id = c.id
			WHERE b.slug = $1
		`
	
	var p models.BlogPost
	var pubDate time.Time
	var thumb *string
	
	err := config.DB.QueryRow(query, slug).Scan(
		&p.ID, &p.Title, &p.Slug, &p.Content, &thumb,
		&pubDate, &p.AuthorName, &p.CategoryName, &p.CategoryID,
	)
	
	if err == sql.ErrNoRows {
		http.Error(w, "Article not found", http.StatusNotFound)
		return
	}
	
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	baseURL := scheme + "://" + r.Host
	
	if thumb != nil {
		p.Thumbnail = baseURL + *thumb
	}
	p.PublishedAt = pubDate.Format("02 January 2006")
	
	relatedQuery := `
			SELECT w.id, w.nama_tempat, w.lokasi, w.harga_tiket, w.rating_total,
			COALESCE((SELECT image_url FROM wisata_images WHERE wisata_id = w.id AND is_primary = true LIMIT 1), '')
			FROM blog_related_wisata br
			JOIN wisata w ON br.wisata_id = w.id
			WHERE br.blog_post_id = $1 AND w.deleted_at IS NULL
		`
	rows, err := config.DB.Query(relatedQuery, p.ID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var w models.PopularWisata
			var wImg string
			rows.Scan(&w.ID, &w.NamaTempat, &w.Lokasi, &w.HargaTiket, &w.RatingTotal, &wImg)
			if wImg != "" {
				w.ImageURL = baseURL + wImg
			}
			p.RelatedWisata = append(p.RelatedWisata, w)
		}
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(
		models.Response{
			Status:  200,
			Message: "Blog detail fetched",
			Data:    p,
		},
	)
}

func CreateBlogPost(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	session, _ := config.AdminStore.Get(r, "admin-session-token")
	if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	authorID := session.Values["user_id"]
	
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "File too large", http.StatusBadRequest)
		return
	}
	
	title := r.FormValue("title")
	slug := r.FormValue("slug")
	excerpt := r.FormValue("excerpt")
	content := r.FormValue("content")
	categoryID := r.FormValue("category_id")
	relatedIDs := r.FormValue("related_wisata_ids")
	
	var imagePath string
	file, header, err := r.FormFile("thumbnail")
	if err == nil {
		defer file.Close()
		path, err := saveUploadedFile(file, header.Filename)
		if err == nil {
			imagePath = path
		}
	}
	
	var newID int
	query := `
			INSERT INTO blog_posts (title, slug, excerpt, content, thumbnail, author_id, blog_category_id, status)
			VALUES ($1, $2, $3, $4, $5, $6, $7, 'published')
			RETURNING id
		`
	err = config.DB.QueryRow(query, title, slug, excerpt, content, imagePath, authorID, categoryID).Scan(&newID)
	
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	if relatedIDs != "" {
		ids := strings.Split(relatedIDs, ",")
		for _, idStr := range ids {
			wid, _ := strconv.Atoi(strings.TrimSpace(idStr))
			if wid > 0 {
				config.DB.Exec("INSERT INTO blog_related_wisata (blog_post_id, wisata_id) VALUES ($1, $2)", newID, wid)
			}
		}
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.Response{Status: 201, Message: "Blog created"})
}

func UpdateBlogPost(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" && r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	session, _ := config.AdminStore.Get(r, "admin-session-token")
	if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	id := r.URL.Query().Get("id")
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "Form error", http.StatusBadRequest)
		return
	}
	
	title := r.FormValue("title")
	slug := r.FormValue("slug")
	excerpt := r.FormValue("excerpt")
	content := r.FormValue("content")
	categoryID := r.FormValue("category_id")
	relatedIDs := r.FormValue("related_wisata_ids")
	
	file, header, err := r.FormFile("thumbnail")
	var imagePath string
	if err == nil {
		defer file.Close()
		path, err := saveUploadedFile(file, header.Filename)
		if err == nil {
			imagePath = path
		}
	}
	
	if imagePath != "" {
		config.DB.Exec(
			"UPDATE blog_posts SET title=$1, slug=$2, excerpt=$3, content=$4, blog_category_id=$5, thumbnail=$6 WHERE id=$7",
			title, slug, excerpt, content, categoryID, imagePath, id,
		)
	} else {
		config.DB.Exec(
			"UPDATE blog_posts SET title=$1, slug=$2, excerpt=$3, content=$4, blog_category_id=$5 WHERE id=$6",
			title, slug, excerpt, content, categoryID, id,
		)
	}
	
	config.DB.Exec("DELETE FROM blog_related_wisata WHERE blog_post_id = $1", id)
	if relatedIDs != "" {
		ids := strings.Split(relatedIDs, ",")
		for _, idStr := range ids {
			wid, _ := strconv.Atoi(strings.TrimSpace(idStr))
			if wid > 0 {
				config.DB.Exec("INSERT INTO blog_related_wisata (blog_post_id, wisata_id) VALUES ($1, $2)", id, wid)
			}
		}
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.Response{Status: 200, Message: "Blog updated"})
}

func DeleteBlogPost(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" && r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	session, _ := config.AdminStore.Get(r, "admin-session-token")
	if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	id := r.URL.Query().Get("id")
	_, err := config.DB.Exec("DELETE FROM blog_posts WHERE id = $1", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.Response{Status: 200, Message: "Blog deleted"})
}

func GetBlogCategories(w http.ResponseWriter, r *http.Request) {
	rows, err := config.DB.Query("SELECT id, name, slug FROM blog_categories ORDER BY id ASC")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	
	var cats []models.BlogCategory
	for rows.Next() {
		var c models.BlogCategory
		rows.Scan(&c.ID, &c.Name, &c.Slug)
		cats = append(cats, c)
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.Response{Status: 200, Data: cats})
}
