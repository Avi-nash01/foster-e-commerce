package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "welcome"
	dbname   = "contact_form"
)

func main() {
	// Build PostgreSQL connection string
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	// Open DB connection
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Error opening DB: %v", err)
	}
	defer db.Close()

	// Verify connection
	if err = db.Ping(); err != nil {
		log.Fatalf("Error connecting to DB: %v", err)
	}

	router := gin.Default()

	// Configure CORS to allow frontend origin
	config := cors.Config{
		AllowOrigins:     []string{"http://localhost:5500", "http://127.0.0.1:5500"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
	router.Use(cors.New(config))

	// POST /submit: Receive contact form data and save to DB
	router.POST("/submit", func(c *gin.Context) {
		var request struct {
			Name    string `json:"name"`
			Email   string `json:"email"`
			Message string `json:"message"`
		}

		if err := c.BindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}

		_, err := db.Exec("INSERT INTO messages (name, email, message) VALUES ($1, $2, $3)",
			request.Name, request.Email, request.Message)
		if err != nil {
			log.Printf("DB insert error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save message"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Submission successful!"})
	})

	// GET /messages: Retrieve all submitted messages
	router.GET("/messages", func(c *gin.Context) {
		rows, err := db.Query("SELECT id, name, email, message, submitted_at FROM messages ORDER BY submitted_at DESC")
		if err != nil {
			log.Printf("DB query error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch messages"})
			return
		}
		defer rows.Close()

		var messages []map[string]interface{}
		for rows.Next() {
			var (
				id          int
				name        string
				email       string
				message     string
				submittedAt time.Time
			)

			if err := rows.Scan(&id, &name, &email, &message, &submittedAt); err != nil {
				log.Printf("Row scan error: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error reading data"})
				return
			}

			messages = append(messages, gin.H{
				"id":           id,
				"name":         name,
				"email":        email,
				"message":      message,
				"submitted_at": submittedAt.Format(time.RFC3339),
			})
		}

		c.JSON(http.StatusOK, messages)
	})

	// Run the server
	log.Println("Server running on http://localhost:8080")
	router.Run(":8080")
}
