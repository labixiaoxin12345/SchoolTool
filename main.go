package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"MyGPACalculator/models"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"golang.org/x/crypto/bcrypt"
)

var db *sql.DB

func main() {
	// Initialize DB Connection
	var err error
	db, err = sql.Open("mysql", "username:password@tcp(127.0.0.1:3306)/rental_management")
	if err != nil {
		log.Fatalf("Error connecting to database: %v\n", err)
	}
	defer db.Close()

	// Test the DB connection
	if err = db.Ping(); err != nil {
		log.Fatalf("Cannot reach the database: %v\n", err)
	}
	// Start servers for each service
	go startUserService(":8080")

	password := "mysecurepassword"
	hashed := hashPassword(password)
	fmt.Println("Hashed Password:", hashed)
	select {}

}

// USER
func startUserService(port string) {
	// Set up routes
	r := mux.NewRouter()
	r.HandleFunc("/login", userLogin).Methods("POST")
	r.HandleFunc("/user", userService).Methods("POST", "PUT")
	r.HandleFunc("/user", getUserInfo).Methods("GET", "PUT")
	r.HandleFunc("/register", createUser).Methods("POST")
	r.HandleFunc("/validate-otp", validateOTP).Methods("POST")
	fmt.Printf("User service running at http://localhost%s\n", port)
	log.Fatal(http.ListenAndServe(port, cors.Default().Handler(r)))
}

// USER
func userLogin(w http.ResponseWriter, r *http.Request) {
	//create another structure to handle log in credentials
	var credentials struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		log.Printf("Invalid input: %v", err)
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}
	// use User strcture from models.go
	var user models.User
	//sql query for database
	err := db.QueryRow("SELECT id, name, email, phone, password_hash FROM users WHERE email = ?", credentials.Email).
		Scan(&user.ID, &user.Name, &user.Email, &user.Phone, &user.PasswordHash)

	if err != nil {
		// check for user email in database
		if err == sql.ErrNoRows {
			log.Printf("User not found for email: %s", credentials.Email)
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		} else {
			log.Printf("Database error: %v", err)
			http.Error(w, "Database error", http.StatusInternalServerError)
		}
		return
	}

	// Compare the provided password with the stored hash
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(credentials.Password))
	if err != nil {
		log.Printf("Password mismatch for email: %s", credentials.Email)
		log.Printf("Stored Hash: %s", user.PasswordHash)
		log.Printf("Provided Password: %s", credentials.Password)

		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	log.Printf("User %s logged in successfully", user.Email)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Login successful",
		"user_id": user.ID,
		"name":    user.Name,
		"email":   user.Email,
	})
}
func getUserInfo(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from query parameters
	userID := r.URL.Query().Get("id")
	if userID == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	// query to fetch user details by id
	row := db.QueryRow("SELECT id, name, email, phone, membership_tier FROM users WHERE id = ?", userID)
	// row := db.QueryRow("SELECT id, name, email, phone, membership_tier, created_at FROM users WHERE id = ?", userID)

	// use User struct from models.go
	var user models.User
	if err := row.Scan(&user.ID, &user.Name, &user.Email, &user.Phone, &user.MembershipTier); err != nil {
		// if err := row.Scan(&user.ID, &user.Name, &user.Email, &user.Phone, &user.MembershipTier, &user.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "User not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to fetch user details", http.StatusInternalServerError)
		}
		return
	}

	// Secure HTTP headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("X-XSS-Protection", "1; mode=block")

	// Respond with user details as JSON
	json.NewEncoder(w).Encode(user)
}

// create a new user account
func createUser(w http.ResponseWriter, r *http.Request) {
	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, `{"message": "Invalid input"}`, http.StatusBadRequest)
		return
	}

	// Fetch the last user ID from the database and generate a new ID
	var lastID string
	err := db.QueryRow("SELECT MAX(id) FROM users WHERE id LIKE 'U%'").Scan(&lastID)
	if err != nil {
		http.Error(w, `{"message": "Failed to fetch the last user ID"}`, http.StatusInternalServerError)
		return
	}
	num := 1
	if lastID != "" {
		numPart := lastID[1:]
		num, _ = strconv.Atoi(numPart)
		num++
	}
	user.ID = fmt.Sprintf("U%03d", num)

	// Hash the password before storing it
	user.PasswordHash = hashPassword(user.Password)

	// Insert the new user into the database
	insertQuery := "INSERT INTO users (id, name, email, phone, password_hash, membership_tier, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)"
	_, err = db.Exec(insertQuery, user.ID, user.Name, user.Email, user.Phone, user.PasswordHash, "Basic", time.Now()) // all new users start with basic tier
	if err != nil {
		log.Printf("Error inserting user: %v\n", err)
		http.Error(w, `Email has been registered. Please log in.`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message":       "User created successfully",
		"user_id":       user.Name,
		"password hash": user.PasswordHash,
	})
}
func hashPassword(password string) string {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hashedPassword)
}
func userService(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// Fetch user details (for profile)
		userID := r.URL.Query().Get("id") // Assume user ID is passed as a query parameter
		if userID == "" {
			http.Error(w, "User ID is required", http.StatusBadRequest)
			return
		}

		var user models.User
		err := db.QueryRow("SELECT id, name, email, phone, membership_tier FROM users WHERE id = ?", userID).
			Scan(&user.ID, &user.Name, &user.Email, &user.Phone, &user.MembershipTier)

		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "User not found", http.StatusNotFound)
			} else {
				http.Error(w, "Database error", http.StatusInternalServerError)
			}
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(user)

	case http.MethodPut:
		// Update user details
		var user models.User
		if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
			http.Error(w, "Invalid input", http.StatusBadRequest)
			return
		}

		_, err := db.Exec("UPDATE users SET name = ?, email = ?, phone = ? WHERE id = ?",
			user.Name, user.Email, user.Phone, user.ID)

		if err != nil {
			http.Error(w, "Failed to update user details", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "User details updated successfully"}`))

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
