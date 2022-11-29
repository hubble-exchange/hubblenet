package limitorders

import (
	"database/sql"
	"errors"
	"fmt"
	"os"

	_ "github.com/mattn/go-sqlite3" // Import go-sqlite3 library
)

type LimitOrder struct {
	positionType      string
	id                int64
	userAddress       string
	baseAssetQuantity int
	price             float64
	status            string
	salt              string
	signature         []byte
}

type LimitOrderDatabase interface {
	InsertLimitOrder(positionType string, userAddress string, baseAssetQuantity int, price float64, salt string, signature []byte) error
	UpdateLimitOrderStatus(userAddress string, salt string, status string) error
	GetLimitOrderByPositionTypeAndPrice(positionType string, price float64) []LimitOrder
}

type limitOrderDatabase struct {
	db *sql.DB
}

func InitializeDatabase() (LimitOrderDatabase, error) {
	if _, err := os.Stat("hubble.db"); err != nil {
		file, err := os.Create("hubble.db") // Create SQLite file
		if err != nil {
			return nil, err
		}
		file.Close()
	}
	dbName := fmt.Sprintf("./hubble%d.db", os.Getpid()) // so that every node has a different database
	database, _ := sql.Open("sqlite3", dbName)          // Open the created SQLite File
	err := createTable(database)                        // Create Database Tables

	lod := &limitOrderDatabase{
		db: database,
	}

	return lod, err
}

func (lod *limitOrderDatabase) InsertLimitOrder(positionType string, userAddress string, baseAssetQuantity int, price float64, salt string, signature []byte) error {
	err := validateInsertLimitOrderInputs(positionType, userAddress, baseAssetQuantity, price, salt, signature)
	if err != nil {
		fmt.Println(err)
		return err
	}
	insertSQL := "INSERT INTO limit_orders(user_address, position_type, base_asset_quantity, price, salt, signature, status) VALUES (?, ?, ?, ?, ?, ?, 'open')"
	statement, err := lod.db.Prepare(insertSQL)
	if err != nil {
		return err
	}
	_, err = statement.Exec(userAddress, positionType, baseAssetQuantity, price, salt, signature)
	return err
}

func (lod *limitOrderDatabase) UpdateLimitOrderStatus(userAddress string, salt string, status string) error {
	// TODO: validate inputs
	updateSQL := "UPDATE limit_orders SET status = ? WHERE user_address = ? AND salt = ?"
	statement, err := lod.db.Prepare(updateSQL)
	if err != nil {
		return err
	}
	_, err = statement.Exec(status, userAddress, salt)
	return err
}

func (lod *limitOrderDatabase) GetLimitOrderByPositionTypeAndPrice(positionType string, price float64) []LimitOrder {
	var rows = &sql.Rows{}
	var limitOrders = []LimitOrder{}
	if positionType == "short" {
		rows = getShortLimitOrderByPrice(lod.db, price)
	}
	if positionType == "long" {
		rows = getLongLimitOrderByPrice(lod.db, price)
	}
	defer rows.Close()
	for rows.Next() {
		var id int64
		var userAddress string
		var baseAssetQuantity int
		var price float64
		var salt string
		var signature []byte
		_ = rows.Scan(&id, &userAddress, &baseAssetQuantity, &price, &salt, &signature)
		limitOrder := &LimitOrder{
			id:                id,
			positionType:      positionType,
			userAddress:       userAddress,
			baseAssetQuantity: baseAssetQuantity,
			price:             price,
			salt:              salt,
			signature:         signature,
		}
		limitOrders = append(limitOrders, *limitOrder)
	}
	return limitOrders
}

func getShortLimitOrderByPrice(db *sql.DB, price float64) *sql.Rows {
	stmt, _ := db.Prepare(`SELECT id, user_address, base_asset_quantity, price, salt, signature 
		from limit_orders
		where position_type = ? and price < ? and status = 'open'`)
	rows, _ := stmt.Query("short", price)
	return rows
}

func getLongLimitOrderByPrice(db *sql.DB, price float64) *sql.Rows {
	stmt, _ := db.Prepare(`SELECT id, user_address, base_asset_quantity, price, salt, signature
		from limit_orders
		where position_type = ? and price > ? and status = 'open'`)
	rows, _ := stmt.Query("long", price)
	return rows
}

func createTable(db *sql.DB) error {
	createLimitOrderTableSql := `CREATE TABLE if not exists limit_orders (
    	"id" integer NOT NULL PRIMARY KEY AUTOINCREMENT,
		"position_type" VARCHAR(64) NOT NULL, 
    	"user_address" VARCHAR(64) NOT NULL,
    	"base_asset_quantity" INTEGER NOT NULL,
    	"price" FLOAT NOT NULL,
    	"status" VARCHAR(64) NOT NULL,
    	"salt" VARCHAR(64) NOT NULL,
		"signature" TEXT NOT NULL
	);`

	statement, err := db.Prepare(createLimitOrderTableSql)
	if err != nil {
		return err
	}
	_, err = statement.Exec() // Execute SQL Statements
	return err
}

func validateInsertLimitOrderInputs(positionType string, userAddress string, baseAssetQuantity int, price float64, salt string, signature []byte) error {
	if positionType == "long" || positionType == "short" {
	} else {
		return errors.New("invalid position type")
	}

	if userAddress == "" {
		return errors.New("user address cannot be blank")
	}

	if baseAssetQuantity == 0 {
		return errors.New("baseAssetQuantity cannot be zero")
	}

	if price == 0 {
		return errors.New("price cannot be zero")
	}

	if salt == "" {
		return errors.New("salt cannot be blank")
	}

	if len(signature) == 0 {
		return errors.New("signature cannot be blank")
	}

	return nil
}
