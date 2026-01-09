-- Default Shop Database Schema for Telegram Shop Bot
-- This file creates the basic tables and inserts sample data for a shop system

-- Create database (optional, user can run this separately)
-- CREATE DATABASE telegram_shop;

-- Use the database
-- \c telegram_shop;  -- For PostgreSQL
-- USE telegram_shop; -- For MySQL

-- Customers table
CREATE TABLE customers (
    customer_id SERIAL PRIMARY KEY,
    telegram_id BIGINT UNIQUE NOT NULL,
    username VARCHAR(255),
    first_name VARCHAR(255),
    last_name VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Vendors table
CREATE TABLE vendors (
    vendor_id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    telegram_username VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Categories table
CREATE TABLE categories (
    category_id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Products table
CREATE TABLE products (
    product_id SERIAL PRIMARY KEY,
    vendor_id INT REFERENCES vendors(vendor_id),
    category_id INT REFERENCES categories(category_id),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    price DECIMAL(10,2) NOT NULL,
    stock_quantity INT DEFAULT 0,
    image_url VARCHAR(500),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Orders table
CREATE TABLE orders (
    order_id SERIAL PRIMARY KEY,
    customer_id INT REFERENCES customers(customer_id),
    total_amount DECIMAL(10,2) NOT NULL,
    status VARCHAR(50) DEFAULT 'pending', -- pending, confirmed, shipped, delivered, cancelled
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Order items table
CREATE TABLE order_items (
    order_item_id SERIAL PRIMARY KEY,
    order_id INT REFERENCES orders(order_id),
    product_id INT REFERENCES products(product_id),
    quantity INT NOT NULL,
    unit_price DECIMAL(10,2) NOT NULL,
    total_price DECIMAL(10,2) NOT NULL
);

-- Insert sample data

-- Sample vendors
INSERT INTO vendors (name, description, telegram_username) VALUES
('Fashion Hub', 'Premium clothing and accessories', '@fashionhub'),
('Tech Gadgets', 'Latest electronics and gadgets', '@techgadgets'),
('Home Essentials', 'Everything for your home', '@homeessentials');

-- Sample categories
INSERT INTO categories (name, description) VALUES
('Clothing', 'Apparel and fashion items'),
('Electronics', 'Gadgets and electronic devices'),
('Home & Garden', 'Home improvement and garden supplies'),
('Accessories', 'Jewelry, bags, and other accessories');

-- Sample products
INSERT INTO products (vendor_id, category_id, name, description, price, stock_quantity, image_url) VALUES
(1, 1, 'Classic T-Shirt', 'Comfortable cotton t-shirt in various colors', 19.99, 100, 'https://example.com/tshirt.jpg'),
(1, 1, 'Denim Jacket', 'Stylish denim jacket for all seasons', 79.99, 50, 'https://example.com/jacket.jpg'),
(1, 4, 'Leather Wallet', 'Genuine leather wallet with multiple compartments', 34.99, 75, 'https://example.com/wallet.jpg'),
(2, 2, 'Wireless Headphones', 'High-quality wireless headphones with noise cancellation', 149.99, 30, 'https://example.com/headphones.jpg'),
(2, 2, 'Smartphone Case', 'Protective case for smartphones', 24.99, 200, 'https://example.com/case.jpg'),
(3, 3, 'Garden Hose', 'Durable garden hose, 50ft length', 39.99, 40, 'https://example.com/hose.jpg'),
(3, 3, 'Tool Set', 'Complete set of basic hand tools', 89.99, 25, 'https://example.com/tools.jpg');

-- Sample customers (these would typically be added when users interact with the bot)
-- INSERT INTO customers (telegram_id, username, first_name) VALUES

-- Create indexes for better performance
CREATE INDEX idx_products_vendor_id ON products(vendor_id);
CREATE INDEX idx_products_category_id ON products(category_id);
CREATE INDEX idx_orders_customer_id ON orders(customer_id);
CREATE INDEX idx_order_items_order_id ON order_items(order_id);
CREATE INDEX idx_customers_telegram_id ON customers(telegram_id);