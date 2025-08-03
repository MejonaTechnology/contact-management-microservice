package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

type ContactSubmission struct {
	ID           int       `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	Phone        *string   `json:"phone"`
	Subject      *string   `json:"subject"`
	Message      string    `json:"message"`
	Source       *string   `json:"source"`
	Status       string    `json:"status"`
	AssignedTo   *int      `json:"assigned_to"`
	ResponseSent bool      `json:"response_sent"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Meta    interface{} `json:"meta,omitempty"`
}

type Stats struct {
	Total      int64 `json:"total"`
	New        int64 `json:"new"`
	InProgress int64 `json:"in_progress"`
	Resolved   int64 `json:"resolved"`
}

var db *sql.DB

func init() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Initialize database
	dbHost := os.Getenv("DB_HOST")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbPort := os.Getenv("DB_PORT")

	dsn := dbUser + ":" + dbPassword + "@tcp(" + dbHost + ":" + dbPort + ")/" + dbName + "?parseTime=true"
	
	var err error
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	if err = db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	log.Println("Database connected successfully")
}

func enableCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	w.Header().Set("Content-Type", "application/json")
	
	response := Response{
		Success: true,
		Message: "Contact Service is healthy",
		Data: map[string]interface{}{
			"status":  "OK",
			"service": "Contact Service",
			"version": "1.0.0",
		},
	}
	
	json.NewEncoder(w).Encode(response)
}

func getContacts(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	rows, err := db.Query(`
		SELECT id, name, email, phone, subject, message, status, source, assigned_to, response_sent, created_at, updated_at 
		FROM contact_submissions 
		ORDER BY created_at DESC 
		LIMIT 10
	`)
	if err != nil {
		log.Printf("Database query error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{
			Success: false,
			Message: "Database error: " + err.Error(),
		})
		return
	}
	defer rows.Close()

	var contacts []ContactSubmission
	for rows.Next() {
		var contact ContactSubmission
		err := rows.Scan(
			&contact.ID, &contact.Name, &contact.Email, &contact.Phone,
			&contact.Subject, &contact.Message, &contact.Status, &contact.Source,
			&contact.AssignedTo, &contact.ResponseSent, &contact.CreatedAt, &contact.UpdatedAt,
		)
		if err != nil {
			log.Printf("Row scan error: %v", err)
			continue
		}
		contacts = append(contacts, contact)
	}

	response := Response{
		Success: true,
		Message: "Contacts retrieved successfully",
		Data:    contacts,
		Meta: map[string]interface{}{
			"total":    len(contacts),
			"page":     1,
			"per_page": 10,
		},
	}

	json.NewEncoder(w).Encode(response)
}

func getContactStats(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var stats Stats

	// Get counts
	db.QueryRow("SELECT COUNT(*) FROM contact_submissions").Scan(&stats.Total)
	db.QueryRow("SELECT COUNT(*) FROM contact_submissions WHERE status = 'new'").Scan(&stats.New)
	db.QueryRow("SELECT COUNT(*) FROM contact_submissions WHERE status = 'in_progress'").Scan(&stats.InProgress)
	db.QueryRow("SELECT COUNT(*) FROM contact_submissions WHERE status = 'resolved'").Scan(&stats.Resolved)

	response := Response{
		Success: true,
		Message: "Contact statistics retrieved successfully",
		Data:    stats,
	}

	json.NewEncoder(w).Encode(response)
}

func createContact(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(Response{
			Success: false,
			Message: "Method not allowed",
		})
		return
	}

	var req struct {
		Name    string `json:"name"`
		Email   string `json:"email"`
		Phone   string `json:"phone"`
		Subject string `json:"subject"`
		Message string `json:"message"`
		Source  string `json:"source"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{
			Success: false,
			Message: "Invalid JSON: " + err.Error(),
		})
		return
	}

	if req.Name == "" || req.Email == "" || req.Message == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{
			Success: false,
			Message: "Name, email, and message are required",
		})
		return
	}

	_, err := db.Exec(`
		INSERT INTO contact_submissions (name, email, phone, subject, message, source, status) 
		VALUES (?, ?, ?, ?, ?, ?, 'new')
	`, req.Name, req.Email, req.Phone, req.Subject, req.Message, req.Source)

	if err != nil {
		log.Printf("Insert error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{
			Success: false,
			Message: "Failed to create contact: " + err.Error(),
		})
		return
	}

	response := Response{
		Success: true,
		Message: "Contact created successfully",
		Data: map[string]interface{}{
			"name":    req.Name,
			"email":   req.Email,
			"status":  "new",
			"message": "Contact submission received",
		},
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func updateContactStatus(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "PUT" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(Response{
			Success: false,
			Message: "Method not allowed",
		})
		return
	}

	// Extract ID from URL path
	path := r.URL.Path
	idStr := ""
	if len(path) > len("/api/v1/dashboard/contacts/") {
		parts := strings.Split(path, "/")
		for i, part := range parts {
			if part == "contacts" && i+1 < len(parts) {
				idStr = parts[i+1]
				break
			}
		}
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{
			Success: false,
			Message: "Invalid contact ID",
		})
		return
	}

	var req struct {
		Status string `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{
			Success: false,
			Message: "Invalid JSON: " + err.Error(),
		})
		return
	}

	_, err = db.Exec("UPDATE contact_submissions SET status = ? WHERE id = ?", req.Status, id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{
			Success: false,
			Message: "Failed to update contact: " + err.Error(),
		})
		return
	}

	response := Response{
		Success: true,
		Message: "Contact status updated successfully",
		Data: map[string]interface{}{
			"id":     id,
			"status": req.Status,
		},
	}

	json.NewEncoder(w).Encode(response)
}

func login(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{
			Success: false,
			Message: "Invalid JSON",
		})
		return
	}

	// Simple admin check
	if req.Email == "admin@mejona.com" && req.Password == "admin123" {
		response := Response{
			Success: true,
			Message: "Login successful",
			Data: map[string]interface{}{
				"user": map[string]interface{}{
					"id":    1,
					"email": req.Email,
					"role":  "admin",
				},
				"token": "simple-test-token",
			},
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	w.WriteHeader(http.StatusUnauthorized)
	json.NewEncoder(w).Encode(Response{
		Success: false,
		Message: "Invalid credentials",
	})
}

func main() {
	defer db.Close()

	// Set up routes
	http.HandleFunc("/health", healthCheck)
	http.HandleFunc("/api/v1/dashboard/contacts", getContacts)
	http.HandleFunc("/api/v1/dashboard/contacts/stats", getContactStats)
	http.HandleFunc("/api/v1/dashboard/contact", createContact)
	http.HandleFunc("/api/v1/dashboard/contacts/", updateContactStatus) // Handle PUT with ID
	http.HandleFunc("/api/v1/auth/login", login)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	log.Printf("Starting Contact Service on port %s", port)
	log.Printf("Available endpoints:")
	log.Printf("  GET  /health - Health check")
	log.Printf("  GET  /api/v1/dashboard/contacts - Get contacts list")
	log.Printf("  GET  /api/v1/dashboard/contacts/stats - Get contact statistics")
	log.Printf("  POST /api/v1/dashboard/contact - Create contact")
	log.Printf("  PUT  /api/v1/dashboard/contacts/{id} - Update contact status")
	log.Printf("  POST /api/v1/auth/login - Login")

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}