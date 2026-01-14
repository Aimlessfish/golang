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
	UpdatedAt        time.Time `db:"updated_at"`
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
	UpdatedAt   time.Time `db:"updated_at"`
}

type Order struct {
	ID          int       `db:"order_id"`
	CustomerID  int       `db:"customer_id"`
	TotalAmount float64   `db:"total_amount"`
	Status      string    `db:"status"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

type OrderItem struct {
	ID         int     `db:"order_item_id"`
	OrderID    int     `db:"order_id"`
	ProductID  int     `db:"product_id"`
	Quantity   int     `db:"quantity"`
	UnitPrice  float64 `db:"unit_price"`
	TotalPrice float64 `db:"total_price"`
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
	if db == nil {
		return sql.ErrConnDone
	}
	err := db.QueryRow("INSERT INTO products (vendor_id, category_id, name, description, price, stock_quantity, image_url) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING product_id",
		p.VendorID, p.CategoryID, p.Name, p.Description, p.Price, p.StockQty, p.ImageURL).Scan(&p.ID)
	return err
}

func GetProduct(id int) (*Product, error) {
	result, err := interact("SELECT product_id, vendor_id, category_id, name, description, price, stock_quantity, image_url, created_at, updated_at FROM products WHERE product_id = $1", id)
	if err != nil {
		return nil, err
	}
	rows := result.(*sql.Rows)
	defer rows.Close()

	if rows.Next() {
		var p Product
		err = rows.Scan(&p.ID, &p.VendorID, &p.CategoryID, &p.Name, &p.Description, &p.Price, &p.StockQty, &p.ImageURL, &p.CreatedAt, &p.UpdatedAt)
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
	if db == nil {
		return sql.ErrConnDone
	}
	err := db.QueryRow("INSERT INTO customers (telegram_id, username, first_name, last_name) VALUES ($1, $2, $3, $4) RETURNING customer_id",
		c.TelegramID, c.Username, c.FirstName, c.LastName).Scan(&c.ID)
	return err
}

func (c *Customer) Update() error {
	_, err := interact("UPDATE customers SET username = $1, first_name = $2, last_name = $3, updated_at = CURRENT_TIMESTAMP WHERE customer_id = $4",
		c.Username, c.FirstName, c.LastName, c.ID)
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
	if db == nil {
		return sql.ErrConnDone
	}
	err := db.QueryRow("INSERT INTO categories (name, description) VALUES ($1, $2) RETURNING category_id",
		cat.Name, cat.Description).Scan(&cat.ID)
	return err
}

func (cat *Category) Update() error {
	_, err := interact("UPDATE categories SET name = $1, description = $2 WHERE category_id = $3",
		cat.Name, cat.Description, cat.ID)
	return err
}

func (cat *Category) Delete() error {
	_, err := interact("DELETE FROM categories WHERE category_id = $1", cat.ID)
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

func GetProductsByCategory(categoryID int) ([]Product, error) {
	result, err := interact("SELECT product_id, vendor_id, category_id, name, description, price, stock_quantity, image_url, created_at, updated_at FROM products WHERE category_id = $1 ORDER BY name", categoryID)
	if err != nil {
		return nil, err
	}
	rows := result.(*sql.Rows)
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var p Product
		err = rows.Scan(&p.ID, &p.VendorID, &p.CategoryID, &p.Name, &p.Description, &p.Price, &p.StockQty, &p.ImageURL, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	return products, nil
}

func GetProductsByVendor(vendorID int) ([]Product, error) {
	result, err := interact("SELECT product_id, vendor_id, category_id, name, description, price, stock_quantity, image_url, created_at, updated_at FROM products WHERE vendor_id = $1 ORDER BY name", vendorID)
	if err != nil {
		return nil, err
	}
	rows := result.(*sql.Rows)
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var p Product
		err = rows.Scan(&p.ID, &p.VendorID, &p.CategoryID, &p.Name, &p.Description, &p.Price, &p.StockQty, &p.ImageURL, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	return products, nil
}

func UpdateProductStock(productID int, quantity int) error {
	_, err := interact("UPDATE products SET stock_quantity = $1, updated_at = CURRENT_TIMESTAMP WHERE product_id = $2", quantity, productID)
	return err
}

// Vendor operations
func (v *Vendor) Insert() error {
	if db == nil {
		return sql.ErrConnDone
	}
	err := db.QueryRow("INSERT INTO vendors (name, description, telegram_username) VALUES ($1, $2, $3) RETURNING vendor_id",
		v.Name, v.Description, v.TelegramUsername).Scan(&v.ID)
	return err
}

func GetVendor(id int) (*Vendor, error) {
	result, err := interact("SELECT vendor_id, name, description, telegram_username, created_at, updated_at FROM vendors WHERE vendor_id = $1", id)
	if err != nil {
		return nil, err
	}
	rows := result.(*sql.Rows)
	defer rows.Close()

	if rows.Next() {
		var v Vendor
		err = rows.Scan(&v.ID, &v.Name, &v.Description, &v.TelegramUsername, &v.CreatedAt, &v.UpdatedAt)
		return &v, err
	}
	return nil, sql.ErrNoRows
}

func (v *Vendor) Update() error {
	_, err := interact("UPDATE vendors SET name = $1, description = $2, telegram_username = $3, updated_at = CURRENT_TIMESTAMP WHERE vendor_id = $4",
		v.Name, v.Description, v.TelegramUsername, v.ID)
	return err
}

// Order operations
func (o *Order) Insert() error {
	if db == nil {
		return sql.ErrConnDone
	}
	err := db.QueryRow("INSERT INTO orders (customer_id, total_amount, status) VALUES ($1, $2, $3) RETURNING order_id",
		o.CustomerID, o.TotalAmount, o.Status).Scan(&o.ID)
	return err
}

func GetOrder(id int) (*Order, error) {
	result, err := interact("SELECT order_id, customer_id, total_amount, status, created_at, updated_at FROM orders WHERE order_id = $1", id)
	if err != nil {
		return nil, err
	}
	rows := result.(*sql.Rows)
	defer rows.Close()

	if rows.Next() {
		var o Order
		err = rows.Scan(&o.ID, &o.CustomerID, &o.TotalAmount, &o.Status, &o.CreatedAt, &o.UpdatedAt)
		return &o, err
	}
	return nil, sql.ErrNoRows
}

func (o *Order) UpdateStatus(status string) error {
	_, err := interact("UPDATE orders SET status = $1, updated_at = CURRENT_TIMESTAMP WHERE order_id = $2", status, o.ID)
	if err == nil {
		o.Status = status
	}
	return err
}

func GetOrdersByCustomer(customerID int) ([]Order, error) {
	result, err := interact("SELECT order_id, customer_id, total_amount, status, created_at, updated_at FROM orders WHERE customer_id = $1 ORDER BY created_at DESC", customerID)
	if err != nil {
		return nil, err
	}
	rows := result.(*sql.Rows)
	defer rows.Close()

	var orders []Order
	for rows.Next() {
		var o Order
		err = rows.Scan(&o.ID, &o.CustomerID, &o.TotalAmount, &o.Status, &o.CreatedAt, &o.UpdatedAt)
		if err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	return orders, nil
}

// OrderItem operations
func (oi *OrderItem) Insert() error {
	if db == nil {
		return sql.ErrConnDone
	}
	err := db.QueryRow("INSERT INTO order_items (order_id, product_id, quantity, unit_price, total_price) VALUES ($1, $2, $3, $4, $5) RETURNING order_item_id",
		oi.OrderID, oi.ProductID, oi.Quantity, oi.UnitPrice, oi.TotalPrice).Scan(&oi.ID)
	return err
}

func GetOrderItems(orderID int) ([]OrderItem, error) {
	result, err := interact("SELECT order_item_id, order_id, product_id, quantity, unit_price, total_price FROM order_items WHERE order_id = $1", orderID)
	if err != nil {
		return nil, err
	}
	rows := result.(*sql.Rows)
	defer rows.Close()

	var items []OrderItem
	for rows.Next() {
		var oi OrderItem
		err = rows.Scan(&oi.ID, &oi.OrderID, &oi.ProductID, &oi.Quantity, &oi.UnitPrice, &oi.TotalPrice)
		if err != nil {
			return nil, err
		}
		items = append(items, oi)
	}
	return items, nil
}

// Transaction helper for creating order with items
func CreateOrderWithItems(customerID int, items []OrderItem) (*Order, error) {
	if db == nil {
		return nil, sql.ErrConnDone
	}

	// Start transaction
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Calculate total
	var total float64
	for _, item := range items {
		item.TotalPrice = item.UnitPrice * float64(item.Quantity)
		total += item.TotalPrice
	}

	// Create order
	var orderID int
	err = tx.QueryRow("INSERT INTO orders (customer_id, total_amount, status) VALUES ($1, $2, 'pending') RETURNING order_id",
		customerID, total).Scan(&orderID)
	if err != nil {
		return nil, err
	}

	// Insert order items
	for _, item := range items {
		_, err = tx.Exec("INSERT INTO order_items (order_id, product_id, quantity, unit_price, total_price) VALUES ($1, $2, $3, $4, $5)",
			orderID, item.ProductID, item.Quantity, item.UnitPrice, item.TotalPrice)
		if err != nil {
			return nil, err
		}

		// Update stock
		_, err = tx.Exec("UPDATE products SET stock_quantity = stock_quantity - $1 WHERE product_id = $2",
			item.Quantity, item.ProductID)
		if err != nil {
			return nil, err
		}
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return nil, err
	}

	// Return the created order
	return GetOrder(orderID)
}
