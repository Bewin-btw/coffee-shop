# Hot Coffee Project

## Overview
Hot Coffee is a web application that facilitates the management of inventory, menu items, orders, and aggregation features for a coffee shop. Built using Go, it offers a modular and scalable design.

## Features
- **Inventory Management**: Endpoints to manage inventory-related operations.
- **Menu Management**: Handle menu items and their attributes.
- **Order Processing**: APIs for placing and tracking orders.
- **Aggregation**: Aggregated views of data, such as sales or inventory summaries.

## Project Structure
hot-coffee/ 
├── cmd/ # Entry point for the application 
    │ 
    └── main.go # Main application file 
├── internal/ # Core business logic 
    │ 
    ├── config/ # Configuration management 
    │ 
    ├── dal/ # Data access layer 
    │ 
    ├── handler/ # HTTP handlers for various endpoints 
    │ 
    ├── service/ # Business services and domain logic 
    ├── models/ # Data models for the application 
    ├── data/ # Sample or seed data 
├── go.mod # Go module file


## Prerequisites
- Go (1.19 or later)
- Git
- A terminal or command prompt

## Setup Instructions

### Clone the Repository
```bash
git clone git@git.platform.alem.school:eaktaev/hot-coffee.git
cd hot-coffee
Install Dependencies
Ensure Go modules are enabled:

go mod tidy
Run the Application
Run the application locally:

go run cmd/main.go
The server will start, and the default port will be displayed in the logs.

API Endpoints

Orders:
         POST /orders: Create a new order.
         GET /orders: Retrieve all orders.
         GET /orders/{id}: Retrieve a specific order by ID.
         PUT /orders/{id}: Update an existing order.
         DELETE /orders/{id}: Delete an order.
         POST /orders/{id}/close: Close an order.

     Menu Items:
         POST /menu: Add a new menu item.
         GET /menu: Retrieve all menu items.
         GET /menu/{id}: Retrieve a specific menu item.
         PUT /menu/{id}: Update a menu item.
         DELETE /menu/{id}: Delete a menu item.

     Inventory:
         POST /inventory: Add a new inventory item.
         GET /inventory: Retrieve all inventory items.
         GET /inventory/{id}: Retrieve a specific inventory item.
         PUT /inventory/{id}: Update an inventory item.
         DELETE /inventory/{id}: Delete an inventory item.

     Aggregations:
         GET /reports/total-sales: Get the total sales amount.
         GET /reports/popular-items: Get a list of popular menu items.


Configurations are managed through the config package in internal/config. Ensure to update the configuration file for environment-specific settings.