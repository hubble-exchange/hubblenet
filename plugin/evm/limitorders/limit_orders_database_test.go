package limitorders

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitializeDatabaseFirstTime(t *testing.T) {
	lod, err := InitializeDatabase()
	dbName := fmt.Sprintf("./hubble%d.db", os.Getpid()) // so that every node has a different database
	assert.NotNil(t, lod)
	assert.Nil(t, err)

	_, err = os.Stat(dbName)
	assert.Nil(t, err)

	db, _ := sql.Open("sqlite3", dbName) // Open the created SQLite File
	rows, err := db.Query("SELECT * FROM limit_orders")
	assert.Nil(t, err)
	assert.False(t, rows.Next())
	os.Remove(dbName)
}

func TestInitializeDatabaseAfterInitializationAlreadyDone(t *testing.T) {
	InitializeDatabase()
	dbName := fmt.Sprintf("./hubble%d.db", os.Getpid()) // so that every node has a different database
	dbFileInfo1, _ := os.Stat(dbName)

	_, err := InitializeDatabase()
	assert.Nil(t, err)

	dbFileInfo2, err := os.Stat(dbName)
	assert.Nil(t, err)
	assert.Equal(t, dbFileInfo1.Size(), dbFileInfo2.Size())
	assert.Equal(t, dbFileInfo1.ModTime(), dbFileInfo2.ModTime())
	os.Remove(dbName)
}

func TestInsertLimitOrderFailureWhenPositionTypeIsWrong(t *testing.T) {
	dbName := fmt.Sprintf("./hubble%d.db", os.Getpid()) // so that every node has a different database
	lod, _ := InitializeDatabase()
	userAddress := ""
	baseAssetQuantity := 10
	price := 10.14
	salt := "123"
	signature := []byte("signature")
	positionType := "neutral"
	err := lod.InsertLimitOrder(positionType, userAddress, baseAssetQuantity, price, salt, signature)
	assert.NotNil(t, err)

	db, _ := sql.Open("sqlite3", dbName) // Open the created SQLite File
	stmt, _ := db.Prepare("SELECT id, base_asset_quantity, price from limit_orders where user_address = ?")
	rows, _ := stmt.Query(userAddress)
	assert.False(t, rows.Next())
	os.Remove(dbName)
}
func TestInsertLimitOrderFailureWhenUserAddressIsBlank(t *testing.T) {
	dbName := fmt.Sprintf("./hubble%d.db", os.Getpid()) // so that every node has a different database
	lod, _ := InitializeDatabase()
	userAddress := ""
	baseAssetQuantity := 10
	price := 10.14
	positionType := "long"
	salt := "123"
	signature := []byte("signature")
	err := lod.InsertLimitOrder(positionType, userAddress, baseAssetQuantity, price, salt, signature)
	assert.NotNil(t, err)

	db, _ := sql.Open("sqlite3", dbName) // Open the created SQLite File
	stmt, _ := db.Prepare("SELECT id, base_asset_quantity, price from limit_orders where user_address = ?")
	rows, _ := stmt.Query(userAddress)
	assert.False(t, rows.Next())
	os.Remove(dbName)
}

func TestInsertLimitOrderFailureWhenBaseAssetQuantityIsZero(t *testing.T) {
	dbName := fmt.Sprintf("./hubble%d.db", os.Getpid()) // so that every node has a different database
	lod, _ := InitializeDatabase()
	userAddress := "0x22Bb736b64A0b4D4081E103f83bccF864F0404aa"
	baseAssetQuantity := 0
	price := 10.14
	positionType := "long"
	salt := "123"
	signature := []byte("signature")
	err := lod.InsertLimitOrder(positionType, userAddress, baseAssetQuantity, price, salt, signature)
	assert.NotNil(t, err)

	db, _ := sql.Open("sqlite3", dbName) // Open the created SQLite File
	stmt, _ := db.Prepare("SELECT id, base_asset_quantity, price from limit_orders where user_address = ?")
	rows, _ := stmt.Query(userAddress)
	assert.False(t, rows.Next())
	os.Remove(dbName)
}

func TestInsertLimitOrderFailureWhenPriceIsZero(t *testing.T) {
	dbName := fmt.Sprintf("./hubble%d.db", os.Getpid()) // so that every node has a different database
	lod, _ := InitializeDatabase()
	userAddress := "0x22Bb736b64A0b4D4081E103f83bccF864F0404aa"
	baseAssetQuantity := 10
	price := 0.0
	positionType := "long"
	salt := "123"
	signature := []byte("signature")
	err := lod.InsertLimitOrder(positionType, userAddress, baseAssetQuantity, price, salt, signature)
	assert.NotNil(t, err)

	db, _ := sql.Open("sqlite3", dbName) // Open the created SQLite File
	stmt, _ := db.Prepare("SELECT id, base_asset_quantity, price from limit_orders where user_address = ?")
	rows, _ := stmt.Query(userAddress)
	assert.False(t, rows.Next())
	os.Remove(dbName)
}

func TestInsertLimitOrderSuccess(t *testing.T) {
	dbName := fmt.Sprintf("./hubble%d.db", os.Getpid()) // so that every node has a different database
	lod, _ := InitializeDatabase()
	userAddress := "0x22Bb736b64A0b4D4081E103f83bccF864F0404aa"
	baseAssetQuantity := 10
	price := 10.14
	positionType := "long"
	salt := "123"
	signature := []byte("signature")
	err := lod.InsertLimitOrder(positionType, userAddress, baseAssetQuantity, price, salt, signature)
	assert.Nil(t, err)

	db, _ := sql.Open("sqlite3", dbName) // Open the created SQLite File
	stmt, _ := db.Prepare("SELECT id, position_type, base_asset_quantity, price from limit_orders where user_address = ?")
	rows, _ := stmt.Query(userAddress)
	defer rows.Close()
	for rows.Next() {
		var queriedId int
		var queriedPositionType string
		var queriedBaseAssetQuantity int
		var queriedPrice float64
		_ = rows.Scan(&queriedId, &queriedPositionType, &queriedBaseAssetQuantity, &queriedPrice)
		assert.Equal(t, 1, queriedId)
		assert.Equal(t, positionType, queriedPositionType)
		assert.Equal(t, baseAssetQuantity, queriedBaseAssetQuantity)
		assert.Equal(t, price, queriedPrice)
	}
	positionType = "short"
	err = lod.InsertLimitOrder(positionType, userAddress, baseAssetQuantity, price, "1", signature)
	assert.Nil(t, err)
	stmt, _ = db.Prepare("SELECT id, user_address, base_asset_quantity, price from limit_orders where position_type = ?")
	rows, _ = stmt.Query(userAddress)
	defer rows.Close()
	for rows.Next() {
		var queriedId int
		var queriedUserAddress string
		var queriedBaseAssetQuantity int
		var queriedPrice float64
		_ = rows.Scan(&queriedId, &queriedUserAddress, &queriedBaseAssetQuantity, &queriedPrice)
		assert.Equal(t, 1, queriedId)
		assert.Equal(t, userAddress, queriedUserAddress)
		assert.Equal(t, baseAssetQuantity, queriedBaseAssetQuantity)
		assert.Equal(t, price, queriedPrice)
	}

	os.Remove(dbName)
}

func TestGetLimitOrderByPositionTypeAndPriceWhenShortOrders(t *testing.T) {
	dbName := fmt.Sprintf("./hubble%d.db", os.Getpid()) // so that every node has a different database
	lod, _ := InitializeDatabase()
	userAddress := "0x22Bb736b64A0b4D4081E103f83bccF864F0404aa"
	baseAssetQuantity := 10
	price1 := 10.14
	price2 := 11.14
	price3 := 12.14
	positionType := "short"
	signature := []byte("signature")
	lod.InsertLimitOrder(positionType, userAddress, baseAssetQuantity, price1, "1", signature)
	lod.InsertLimitOrder(positionType, userAddress, baseAssetQuantity, price2, "2", signature)
	lod.InsertLimitOrder(positionType, userAddress, baseAssetQuantity, price3, "3", signature)
	orders := lod.GetLimitOrderByPositionTypeAndPrice("short", 12.00)
	assert.Equal(t, 2, len(orders))
	for i := 0; i < len(orders); i++ {
		assert.Equal(t, orders[i].userAddress, userAddress)
		assert.Equal(t, orders[i].baseAssetQuantity, baseAssetQuantity)
		assert.Equal(t, orders[i].positionType, positionType)
	}
	assert.Equal(t, price1, orders[0].price)
	assert.Equal(t, price2, orders[1].price)
	os.Remove(dbName)
}

func TestGetLimitOrderByPositionTypeAndPriceWhenLongOrders(t *testing.T) {
	dbName := fmt.Sprintf("./hubble%d.db", os.Getpid()) // so that every node has a different database
	lod, _ := InitializeDatabase()
	userAddress := "0x22Bb736b64A0b4D4081E103f83bccF864F0404aa"
	baseAssetQuantity := 10
	price1 := 10.14
	price2 := 11.14
	price3 := 12.14
	positionType := "long"
	signature := []byte("signature")
	lod.InsertLimitOrder(positionType, userAddress, baseAssetQuantity, price1, "1", signature)
	lod.InsertLimitOrder(positionType, userAddress, baseAssetQuantity, price2, "2", signature)
	lod.InsertLimitOrder(positionType, userAddress, baseAssetQuantity, price3, "3", signature)
	orders := lod.GetLimitOrderByPositionTypeAndPrice("long", 11.00)
	assert.Equal(t, 2, len(orders))
	for i := 0; i < len(orders); i++ {
		assert.Equal(t, orders[i].userAddress, userAddress)
		assert.Equal(t, orders[i].baseAssetQuantity, baseAssetQuantity)
		assert.Equal(t, orders[i].positionType, positionType)
	}
	assert.Equal(t, price2, orders[0].price)
	assert.Equal(t, price3, orders[1].price)
	os.Remove(dbName)
}
