package blockchain

import (
	"database/sql"
	"os"
	"time"
)

type Blockchain struct {
	DB *sql.DB
	index uint64
}

func NewChain(filename, receiver string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	file.Close()
	db, err := sql.Open("sqlite3", filename)
	if err != nil {
		return nil
	}
	defer db.Close()

	_, err = db.Exec(CREATE_TABLE)
	if err != nil {
		return nil
	}
	chain := &Blockchain{
		DB: db,
	}

	genesis := &Block{
		PrevHash: []byte(GENESIS_BLOCK),
		Mapping: make(map[string]uint64),
		Miner: receiver,
		TimeStamp: time.Now().Format(time.RFC3339),
	}
	genesis.Mapping[STORAGE_CHAIN] = STORAGE_VALUE
	genesis.Mapping[receiver] = GENESIS_REWARD
	genesis.CurrHash = genesis.hash()
	chain.AddBlock(genesis)
	return nil
}

func (chain *Blockchain) AddBlock(block *Block) {
	chain.index += 1
	chain.DB.Exec("INSERT INTO Blockchain (Hash, Block) VALUES ($1, $2)",
		Base64Encode(block.CurrHash),
		SerializeBlock(block),
	)
}

func LoadChain(filename string) *Blockchain {
	db, err := sql.Open("sqlite3", filename)
	if err != nil {
		return nil
	}
	chain := &Blockchain{
		DB: db,
	}
	chain.index = chain.Size()
	return chain
}

func (chain *Blockchain) Size() uint64 {
	var index uint64
	row := chain.DB.QueryRow("SELECT Id FROM BlockChain ORDER BY Id DESC")
	row.Scan(&index)
	return index
}

func (chain *Blockchain) LastHash() []byte {
	var hash string
	row := chain.DB.QueryRow("SELECT Hash FROM BlockChain ORDER BY Id DESC")
	row.Scan(&hash)
	return Base64Decode(hash)
}

func (chain *Blockchain) Balance(address string) uint64 {
	var (
		balance uint64
		sblock string
		block *Block
	)
	rows, err := chain.DB.Query("SELECT Block FROM BlockChain WHERE Id <= $1 ORDER BY Id DESC", chain.index)
	if err != nil {
		return balance
	}
	defer rows.Close()
	for rows.Next() {
		rows.Scan(&sblock)
		block = DeserializeBlock(sblock)
		if value, ok := block.Mapping[address]; ok {
			balance = value
			break
		}
	}
	return balance
}

