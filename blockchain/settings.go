package blockchain

import (
	"time"
	mrand "math/rand"
)

// len(base64(sha256(data))) = 44
const (
	CREATE_TABLE = `
		CREATE TABLE BlockChain (
			Id INTEGER PRIMARY KEY AUTOINCREMENT,
			Hash VARCHAR(44) UNIQUE,
			Block TEXT
		)
	`
)

const (
	KEY_SIZE = 512 //very small, only test
	DEBUG = true
	TXS_LIMIT = 2
	DIFFICULTY = 20
	RAND_BYTES = 32
	START_PERCENT = 10
	STORAGE_REWARD = 1
)

const (
	GENESIS_BLOCK = "GENESIS-BLOCK"
	STORAGE_VALUE = 100
	GENESIS_REWARD = 100
	STORAGE_CHAIN = "STORAGE-CHAIN"
)

func init() {
	mrand.Seed(time.Now().UnixNano())
}