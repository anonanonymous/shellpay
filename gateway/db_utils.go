package main

import "time"

// Key - represents key table in the signatures database
type Key struct {
	ID      int
	Pub     string
	Priv    string
	Address string
	Expiry  int64
}

// getKey - retrieves a key from the database
func getKey(pub string) (*Key, error) {
	row := gatewayDB.QueryRow(`
		SELECT id, pub, priv, address, expiry FROM keys
		WHERE pub = $1;`, pub,
	)
	k := Key{}
	err := row.Scan(&k.ID, &k.Pub, &k.Priv, &k.Address, &k.Expiry)
	return &k, err
}

// setKey - inserts a key into the database
func setKey(pub, priv, addr string, expiry int64) bool {
	_, err := gatewayDB.Exec(`
		INSERT INTO	keys (pub, priv, address, expiry)
		VALUES ($1, $2, $3, $4);`, pub, priv, addr, expiry,
	)
	return err == nil
}

// cleanupDB - removes expired keys
func cleanupDB() {
	gatewayDB.Exec(`
		DELETE FROM keys
		WHERE expiry <= $1`, time.Now().Unix(),
	)
	time.AfterFunc(time.Hour*12, cleanupDB)
}
