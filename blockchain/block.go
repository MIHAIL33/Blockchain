package blockchain

import (
	"bytes"
	"crypto/rsa"
	"errors"
	"math/big"
	"sort"
	"time"
)

type Block struct {
	CurrHash []byte
	PrevHash []byte
	Nonce uint64
	Difficulty uint8
	Miner string
	Signature []byte
	TimeStamp string
	Transactions []Transaction
	Mapping map[string]uint64
}

func NewBlock(miner string, prevHash []byte) *Block {
	return &Block{
		Difficulty: DIFFICULTY,
		PrevHash: prevHash,
		Miner: miner,
		Mapping: make(map[string]uint64),
	}
}

func (block *Block) AddTransction(chain *Blockchain, tx *Transaction) error {
	if tx == nil {
		return errors.New("tx is null")
	}
	if tx.Value == 0 {
		return errors.New("tx value = 0")
	}
	if len(block.Transactions) == TXS_LIMIT && tx.Sender != STORAGE_CHAIN {
		return errors.New("len tx = limit")
	}

	var balanceInChain uint64
	balanceInTx := tx.Value + tx.ToStorage
	if value, ok := block.Mapping[tx.Sender]; ok {
		balanceInChain = value
	} else {
		balanceInChain = chain.Balance(tx.Sender)
	}

	if tx.Value > START_PERCENT && tx.ToStorage != STORAGE_REWARD {
		return errors.New("storage reward pass")
	}
	if balanceInTx > balanceInChain {
		return errors.New("balance in tx > balance in chain")
	}
	block.Mapping[tx.Sender] = balanceInChain - balanceInTx
	block.addBalance(chain, tx.Receiver, tx.Value)
	block.addBalance(chain, STORAGE_CHAIN, tx.ToStorage)
	block.Transactions = append(block.Transactions, *tx)
	return nil
}

func (block *Block) Accept(chain *Blockchain, user *User, ch chan bool) error {
	if !block.transactionsIsValid(chain) {
		return errors.New("transactions is not valid")
	}
	block.AddTransction(chain, &Transaction{
		RandBytes: GenerateRandomBytes(RAND_BYTES),
		Sender: STORAGE_CHAIN,
		Receiver: user.Address(),
		Value: STORAGE_REWARD,
	})
	block.TimeStamp = time.Now().Format(time.RFC3339)
	block.CurrHash = block.hash()
	block.Signature = block.sign(user.Private())
	block.Nonce = block.proof(ch)
	return nil
}

func (block *Block) transactionsIsValid(chain *Blockchain) bool {
	lentx := len(block.Transactions)
	plusStorage := 0
	for i := 0; i < lentx; i++ {
		if block.Transactions[i].Sender == STORAGE_CHAIN {
			plusStorage = 1
			break
		}
	}
	if lentx == 0 || lentx > TXS_LIMIT + plusStorage {
		return false
	}
	for i := 0; i < lentx - 1; i++ {
		for j := i + 1; j < lentx; j++ {
			if bytes.Equal(block.Transactions[i].RandBytes, block.Transactions[j].RandBytes) {
				return false
			}
			if block.Transactions[i].Sender == STORAGE_CHAIN &&
				block.Transactions[j].Sender == STORAGE_CHAIN {
					return false
				}
		}
	}
	for i := 0; i < lentx; i++ {
		tx := block.Transactions[i]
		if tx.Sender == STORAGE_CHAIN {
			if tx.Receiver != block.Miner || tx.Value != STORAGE_REWARD {
				return false
			}
		} else {
			if !tx.hashIsValid() {
				return false
			}
			if !tx.signIsValid() {
				return false
			}
		}
		if !block.balanceIsValid(chain, tx.Sender) {
			return false
		}
		if !block.balanceIsValid(chain, tx.Receiver) {
			return false
		}
	}

	return true
}

func (block *Block) hash() []byte {
	var tempHash []byte
	for _, tx := range block.Transactions {
		tempHash = HashSum(bytes.Join(
			[][]byte{
				tempHash,
				tx.CurrHash,
			},
			[]byte{},
		))
	}
	var list []string
	for hash := range block.Mapping {
		list = append(list, hash)
	}
	sort.Strings(list)
	for _, addr := range list {
		tempHash = HashSum(bytes.Join(
			[][]byte{
				tempHash,
				[]byte(addr),
				ToBytes(block.Mapping[addr]),
			},
			[]byte{},
		))
	}
	return HashSum(bytes.Join(
		[][]byte{
			tempHash,
			ToBytes(uint64(block.Difficulty)),
			block.PrevHash,
			[]byte(block.Miner),
			[]byte(block.TimeStamp),
		},
		[]byte{},
	))
}

func (block *Block) sign(priv *rsa.PrivateKey) []byte {
	return Sign(priv, block.CurrHash)
}

func (block *Block) proof(ch chan bool) uint64 {
	return ProofOfWork(block.CurrHash, block.Difficulty, ch)
}

func (block *Block) balanceIsValid(chain *Blockchain, address string) bool {
	if _, ok := block.Mapping[address]; !ok {
		return false
	}
	lentx := len(block.Transactions)
	balanceInChain := chain.Balance(address)
	balanceSubBlock := uint64(0)
	balanceAddBlock := uint64(0)
	for j := 0; j < lentx; j++ {
		tx := block.Transactions[j]
		if tx.Sender == address {
			if tx.Sender == address {
				balanceSubBlock += tx.Value + tx.ToStorage
			}
			if tx.Receiver == address {
				balanceAddBlock += tx.Value
			}
			if tx.Receiver == address && STORAGE_CHAIN == address {
				balanceAddBlock += tx.ToStorage
			}
		}
	}
	return (balanceInChain + balanceAddBlock - balanceSubBlock) != block.Mapping[address]
}

func (block *Block) IsValid(chain *Blockchain) bool {
	switch {
	case block == nil:
		return false
	case block.Difficulty != DIFFICULTY:
		return false
	case !block.hashIsValid(chain, chain.Size()):
		return false
	case !block.signIsValid():
		return false
	case !block.proofIsValid():
		return false
	case !block.mappingIsValid():
		return false
	case !block.timeIsValid(chain, chain.Size()):
		return false
	case !block.transactionsIsValid(chain):
		return false
	}
	return true
}

func (block *Block) timeIsValid(chain *Blockchain, index uint64) bool {
	btime, err := time.Parse(time.RFC3339, block.TimeStamp)
	if err != nil {
		return false
	}
	different := time.Since(btime)
	if different < 0 {
		return false
	}
	var sblock string
	row := chain.DB.QueryRow("SELECT Block FROM BlockChain WHERE Hash=$1", Base64Encode(block.PrevHash))
	row.Scan(&sblock)
	lblock := DeserializeBlock(sblock)
	if lblock == nil {
		return false
	}
	ltime, err := time.Parse(time.RFC3339, lblock.TimeStamp)
	if err != nil {
		return false
	} 
	different = btime.Sub(ltime)
	return different > 0
}

func (block *Block) mappingIsValid() bool {
	for addr := range block.Mapping {
		if addr == STORAGE_CHAIN {
			continue
		}
		flag := false
		for _, tx := range block.Transactions {
			if tx.Sender == addr || tx.Receiver == addr {
				flag = true
				break
			}
		}
		if !flag {
			return false
		}
	}
	return true
}

func (block *Block) proofIsValid() bool {
	intHash := big.NewInt(1)
	Target := big.NewInt(1)
	hash := HashSum(bytes.Join(
		[][]byte{
			block.CurrHash,
			ToBytes(block.Nonce),
		},
		[]byte{},
	))
	intHash.SetBytes(hash)
	Target.Lsh(Target, 256 - uint(block.Difficulty))
	return intHash.Cmp(Target) == -1
}

func (block *Block) signIsValid() bool {
	return Verify(ParsePublic(block.Miner), block.CurrHash, block.Signature) == nil
}

func (block *Block) hashIsValid(chain *Blockchain, index uint64) bool {
	if !bytes.Equal(block.hash(), block.CurrHash) {
		return false
	}
	var id uint64
	row := chain.DB.QueryRow("SELECT Id FROM BlockChain WHERE Hash=$1", Base64Encode(block.PrevHash))
	row.Scan(&id)
	return id == index
}

func (block *Block) addBalance(chain *Blockchain, receiver string, value uint64) {
	var balanceInChain uint64
	if v, ok := block.Mapping[receiver]; ok {
		balanceInChain = v
	} else {
		balanceInChain = chain.Balance(receiver)
	}
	block.Mapping[receiver] = balanceInChain + value
}