package main

import (
	"fmt"

	bc "github.com/MIHAIL33/CryptoCoin/blockchain"
)

const (
	DBNAME = "blockchain.db"
)

func main() {
	miner := bc.NewUser()
	bc.NewChain(DBNAME, miner.Address())
	chain := bc.LoadChain(DBNAME)
	for i := 0; i < 3; i++ {
		block := bc.NewBlock(miner.Address(), chain.LastHash())
		block.AddTransaction(chain, bc.NewTransaction(miner, chain.LastHash(), "SomePeople1", 3))
		block.AddTransaction(chain, bc.NewTransaction(miner, chain.LastHash(), "SomePeople2", 2))
		block.Accept(chain, miner, make(chan bool))
		chain.AddBlock(block)
	}

	var sblock string
	rows, err := chain.DB.Query("SELECT Block FROM BlockChain")
	if err != nil {
		panic(err)
	}
	for rows.Next() {
		rows.Scan(&sblock)
		fmt.Println(sblock)
	} 
}