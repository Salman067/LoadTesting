package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

// Assume we have a MySQL database connection
var db *sql.DB

// Assume we have a Redis client connection
var rdb *redis.Client

// Assume we have an internal cache
var cache map[string]string

type User struct {
	ID       int    `json:"user_id"`
	UserName string `json:"user_name"`
	Email    string `json:"email"`
}

func main() {
	r := mux.NewRouter()
	// localConfig := config.LocalConfig
	// Initialize database connection
	// dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
	// 	localConfig.DBUser, localConfig.DBPass, localConfig.DBHOST, localConfig.DBPort, localConfig.DBName)
	// fmt.Println(localConfig.DBHOST)
	Db, err := sql.Open("mysql", "root:Salman12@@tcp(mysql:3306)/loadtesting?charset=utf8mb4&parseTime=True&loc=Local")
	// Db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	// if err := migrate(db); err != nil {
	// 	log.Println("hello", err)
	// }

	db = Db
	// userSeed()
	// Initialize Redis client
	rdb = redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	// Initialize internal cache
	cache = make(map[string]string)

	// Define HTTP routes
	r.HandleFunc("/dataFromDatabase", getDataFromDatabase).Methods("GET")
	r.HandleFunc("/dataFromDatabaseByParams", getDataFromDatabaseByParams).Methods("GET")
	r.HandleFunc("/dataFromRedis", getDataFromRedis).Methods("GET")
	r.HandleFunc("/dataFromCache", getDataFromCache).Methods("GET")
	http.Handle("/", r)

	// Start server
	fmt.Println("Server starting....")
	log.Fatal(http.ListenAndServe("0.0.0.0:4000", r))
}

func getDataFromDatabase(w http.ResponseWriter, r *http.Request) {
	// userSeed()
	// Query data from database
	rows, err := db.Query("SELECT * FROM users")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Convert rows to JSON
	var result []*User
	user := User{}

	for rows.Next() {
		if err := rows.Scan(&user.ID, &user.UserName, &user.Email); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		result = append(result, &User{
			ID:       user.ID,
			UserName: user.UserName,
			Email:    user.Email,
		})
	}
	useID := strconv.FormatUint(uint64(user.ID), 10)
	err = setDataToRedis(useID, result)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Write JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func getDataFromDatabaseByParams(w http.ResponseWriter, r *http.Request) {
	// Query data from database
	id := r.URL.Query().Get("user_id")
	userEmail := r.URL.Query().Get("email")
	rows, err := db.Query("SELECT * FROM users WHERE user_id=? OR email=?", id, userEmail)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Convert rows to JSON
	var result []*User
	user := User{}

	for rows.Next() {
		if err := rows.Scan(&user.ID, &user.UserName, &user.Email); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		result = append(result, &User{
			ID:       user.ID,
			UserName: user.UserName,
			Email:    user.Email,
		})
	}
	useID := strconv.FormatUint(uint64(user.ID), 10)
	err = setDataToRedis(useID, result)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Write JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
func getDataFromRedis(w http.ResponseWriter, r *http.Request) {
	var result []*User
	user := User{}
	useID := strconv.FormatUint(uint64(user.ID), 10)
	getData, _ := getDataToRedis(useID)
	if getData == nil {
		fmt.Println("hhgfgh")
		rows, err := db.Query("SELECT * FROM users")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// Convert rows to JSON

		for rows.Next() {
			if err := rows.Scan(&user.ID, &user.UserName, &user.Email); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			result = append(result, &User{
				ID:       user.ID,
				UserName: user.UserName,
				Email:    user.Email,
			})
		}
		fmt.Println(result)
		err = setDataToRedis(useID, result)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Write JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(getData)
}

func getDataFromCache(w http.ResponseWriter, r *http.Request) {
	// Get data from cache
	users := []*User{}
	for i := 1; i <= 6000; i++ {
		user := User{
			ID:       i,
			UserName: faker.Name(),
			Email:    faker.Email(),
		}
		users = append(users, &user)
	}
	// Write JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func setDataToRedis(key string, user []*User) error {
	data, err := json.Marshal(user)
	// Set data to Redis
	if err != nil {
		return err
	}

	return rdb.Set(rdb.Context(), key, data, 0).Err()
}

func getDataToRedis(key string) ([]*User, error) {
	var msg []*User
	data, err := rdb.Get(rdb.Context(), key).Bytes()
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(data, &msg)
	if err != nil {
		return nil, err
	}

	return msg, nil
}

func userSeed() {
	for i := 1; i <= 6000; i++ {
		user := User{
			ID:       i,
			UserName: faker.Name(),
			Email:    faker.Email(),
		}

		// Create the record
		stmt, err := db.Prepare("INSERT INTO users(user_id, user_name,email) VALUES(?, ?,?)")
		if err != nil {
			log.Fatal(err)
		}
		// defer stmt.Close()

		_, err = stmt.Exec(user.ID, user.UserName, user.Email)
		if err != nil {
			log.Fatal(err)
		}
	}
}

// func migrate(db *sql.DB) error {
// 	_, err := db.Exec(`
//         CREATE TABLE IF NOT EXISTS users (
//             user_id int PRIMARY KEY,
//             user_name VARCHAR(50) NOT NULL,
//             email VARCHAR(50) NOT NULL
//         )
//     `)
// 	return err
// }
