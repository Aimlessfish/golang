package db

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	util "telegramconnect/util"

	_ "github.com/lib/pq"
)

// Data structures for database entities
type Customer struct {
	ID         int       `db:"customer_id"`
	TelegramID int64     `db:"telegram_id"`
	Username   string    `db:"username"`
	FirstName  string    `db:"first_name"`
	LastName   string    `db:"last_name"`
	CreatedAt  time.Time `db:"created_at"`
	UpdatedAt  time.Time `db:"updated_at"`
}

type Vendor struct {
	ID               int       `db:"vendor_id"`
	Name             string    `db:"name"`
	Description      string    `db:"description"`
	TelegramUsername string    `db:"telegram_username"`
	CreatedAt        time.Time `db:"created_at"`
}

type Category struct {
	ID          int       `db:"category_id"`
	Name        string    `db:"name"`
	Description string    `db:"description"`
	CreatedAt   time.Time `db:"created_at"`
}

type Product struct {
	ID          int       `db:"product_id"`
	VendorID    int       `db:"vendor_id"`
	CategoryID  int       `db:"category_id"`
	Name        string    `db:"name"`
	Description string    `db:"description"`
	Price       float64   `db:"price"`
	StockQty    int       `db:"stock_quantity"`
	ImageURL    string    `db:"image_url"`
	CreatedAt   time.Time `db:"created_at"`
}

var db *sql.DB

func Connect() (string, int, error) {
	dbVars, err := util.ValueGetter("DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME")
	if err != nil {
		return "Failed to get database variables", 500, err
	}
	host := dbVars["DB_HOST"]
	portStr := dbVars["DB_PORT"]
	user := dbVars["DB_USER"]
	pass := dbVars["DB_PASSWORD"]
	dbname := dbVars["DB_NAME"]

	// Convert port to int if needed
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return "Failed to convert port to int", 500, err
	}

	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, pass, dbname)

	var dbErr error
	db, dbErr = sql.Open("postgres", connStr)
	if dbErr != nil {
		return "Failed to open database connection", 500, dbErr
	}

	err = db.Ping()
	if err != nil {
		return "Failed to ping database", 500, err
	}

	return "Connected to database", 200, nil
}

func Disconnect() {
	if db != nil {
		db.Close()
		db = nil
	}
}

func interact(query string, args ...interface{}) (interface{}, error) {
	if db == nil {
		return nil, sql.ErrConnDone
	}

	// Check if it's a SELECT query (returns rows) or modifying query (returns result)
	trimmed := strings.TrimSpace(strings.ToUpper(query))
	if strings.HasPrefix(trimmed, "SELECT") {
		return db.Query(query, args...)
	} else {
		return db.Exec(query, args...)
	}
}

// Product operations
func (p *Product) Insert() error {
	result, err := interact("INSERT INTO products (vendor_id, category_id, name, description, price, stock_quantity, image_url) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING product_id",
		p.VendorID, p.CategoryID, p.Name, p.Description, p.Price, p.StockQty, p.ImageURL)
	if err != nil {
		return err
	}
	sqlResult := result.(sql.Result)
	id, err := sqlResult.LastInsertId()
	p.ID = int(id)
	return err
}

func GetProduct(id int) (*Product, error) {
	result, err := interact("SELECT product_id, vendor_id, category_id, name, description, price, stock_quantity, image_url, created_at FROM products WHERE product_id = $1", id)
	if err != nil {
		return nil, err
	}
	rows := result.(*sql.Rows)
	defer rows.Close()

	if rows.Next() {
		var p Product
		err = rows.Scan(&p.ID, &p.VendorID, &p.CategoryID, &p.Name, &p.Description, &p.Price, &p.StockQty, &p.ImageURL, &p.CreatedAt)
		return &p, err
	}
	return nil, sql.ErrNoRows
}

func (p *Product) Update() error {
	_, err := interact("UPDATE products SET vendor_id = $1, category_id = $2, name = $3, description = $4, price = $5, stock_quantity = $6, image_url = $7 WHERE product_id = $8",
		p.VendorID, p.CategoryID, p.Name, p.Description, p.Price, p.StockQty, p.ImageURL, p.ID)
	return err
}

func (p *Product) Delete() error {
	_, err := interact("DELETE FROM products WHERE product_id = $1", p.ID)
	return err
}

// Customer operations
func (c *Customer) Insert() error {
	result, err := interact("INSERT INTO customers (telegram_id, username, first_name, last_name) VALUES ($1, $2, $3, $4) RETURNING customer_id",
		c.TelegramID, c.Username, c.FirstName, c.LastName)
	if err != nil {
		return err
	}
	sqlResult := result.(sql.Result)
	id, err := sqlResult.LastInsertId()
	c.ID = int(id)
	return err
}

func GetCustomerByTelegramID(telegramID int64) (*Customer, error) {
	result, err := interact("SELECT customer_id, telegram_id, username, first_name, last_name, created_at, updated_at FROM customers WHERE telegram_id = $1", telegramID)
	if err != nil {
		return nil, err
	}
	rows := result.(*sql.Rows)
	defer rows.Close()

	if rows.Next() {
		var c Customer
		err = rows.Scan(&c.ID, &c.TelegramID, &c.Username, &c.FirstName, &c.LastName, &c.CreatedAt, &c.UpdatedAt)
		return &c, err
	}
	return nil, sql.ErrNoRows
}

// Category operations
func (cat *Category) Insert() error {
	result, err := interact("INSERT INTO categories (name, description) VALUES ($1, $2) RETURNING category_id",
		cat.Name, cat.Description)
	if err != nil {
		return err
	}
	sqlResult := result.(sql.Result)
	id, err := sqlResult.LastInsertId()
	cat.ID = int(id)
	return err
}

func GetAllCategories() ([]Category, error) {
	result, err := interact("SELECT category_id, name, description, created_at FROM categories ORDER BY name")
	if err != nil {
		return nil, err
	}
	rows := result.(*sql.Rows)
	defer rows.Close()

	var categories []Category
	for rows.Next() {
		var cat Category
		err = rows.Scan(&cat.ID, &cat.Name, &cat.Description, &cat.CreatedAt)
		if err != nil {
			return nil, err
		}
		categories = append(categories, cat)
	}
	return categories, nil
}
