package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

func mustEnv(k string) string {
	v := os.Getenv(k)
	if v == "" {
		log.Fatalf("%s not set", k)
	}
	return v
}

// Converts a mysql://user:pass@host:port/dbname URL into a DSN for go-sql-driver/mysql
func mysqlURLToDSN(mysqlURL string) (string, string) {
	u, err := url.Parse(mysqlURL)
	if err != nil {
		log.Fatalf("invalid DB URL: %v", err)
	}
	if u.Scheme != "mysql" {
		log.Fatalf("unsupported scheme: %s", u.Scheme)
	}

	user := ""
	pass := ""
	if u.User != nil {
		user = u.User.Username()
		pw, _ := u.User.Password()
		pass = pw
	}
	host := u.Host // includes :port
	db := strings.TrimPrefix(u.Path, "/")

	// Add parseTime=true for time scanning
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true", user, pass, host, db)
	return dsn, db
}

func main() {
	connURL := mustEnv("DB_CONNECTION_URL")
	log.Printf("DB_CONNECTION_URL=%s", connURL)

	dsn, dbName := mysqlURLToDSN(connURL)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("ping db: %v", err)
	}

	// Ensure table exists (id autoincrement, name)
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS users (
		id INT AUTO_INCREMENT PRIMARY KEY,
		name VARCHAR(255),
		email VARCHAR(255)
	)`)
	if err != nil {
		log.Fatalf("create table: %v", err)
	}

	http.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			name := r.URL.Query().Get("name")
			if name == "" {
				http.Error(w, "name required", http.StatusBadRequest)
				return
			}
			if _, err := db.Exec("INSERT INTO users(name) VALUES(?)", name); err != nil {
				http.Error(w, "insert failed: "+err.Error(), http.StatusInternalServerError)
				return
			}
			fmt.Fprintf(w, "ok: inserted %q into %s.users\n", name, dbName)
		case http.MethodGet:
			rows, err := db.Query("SELECT id, name FROM users ORDER BY id")
			if err != nil {
				http.Error(w, "select failed: "+err.Error(), http.StatusInternalServerError)
				return
			}
			defer rows.Close()
			w.Header().Set("Content-Type", "text/plain")
			for rows.Next() {
				var id int
				var name string
				_ = rows.Scan(&id, &name)
				fmt.Fprintf(w, "%d\t%s\n", id, name)
			}
		default:
			http.Error(w, "only GET/POST", http.StatusMethodNotAllowed)
		}
	})

	log.Println("users-api listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
