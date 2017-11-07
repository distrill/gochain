package main

import (
	"log"

	"github.com/boltdb/bolt"
)

const dbFile = "blockchain.db"
const blocksBucket = "blocks"

// Blockchain - shut up golinter
type Blockchain struct {
	tip []byte
	db  *bolt.DB
}

// BlockchainIterator - u kno, sometimes u gotta iterate the ol' blockchain
type BlockchainIterator struct {
	currentHash []byte
	db          *bolt.DB
}

// AddBlock - shut up golinter
func (bc *Blockchain) AddBlock(data string) {
	var lastHash []byte

	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("l"))

		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	newBlock := NewBlock(data, lastHash)

	err = bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		err := b.Put(newBlock.Hash, newBlock.Serialize())
		if err != nil {
			log.Fatal(err)
		}

		err = b.Put([]byte("l"), newBlock.Hash)
		if err != nil {
			log.Fatal(err)
		}

		bc.tip = newBlock.Hash

		return nil
	})
}

// NewBlockChain - shut up golinter
func NewBlockChain() *Blockchain {
	var tip []byte
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))

		if b == nil {
			// there is no blockchain yet. create and store, set tip
			genesis := NewGenesisBlock()
			b, err := tx.CreateBucket([]byte(blocksBucket))
			err = b.Put(genesis.Hash, genesis.Serialize())
			if err != nil {
				log.Fatal(err)
			}
			err = b.Put([]byte("l"), genesis.Hash)
			if err != nil {
				log.Fatal(err)
			}
			tip = genesis.Hash
		} else {
			// there is a blockchain. set tip
			tip = b.Get([]byte("l"))
		}
		return nil
	})

	if err != nil {
		log.Fatal(err)
	}

	return &Blockchain{tip, db}
}

// Iterator - for iterating over your blockchain
func (bc *Blockchain) Iterator() *BlockchainIterator {
	return &BlockchainIterator{bc.tip, bc.db}
}

// Next - also for iterating over your blockchain
func (i *BlockchainIterator) Next() *Block {
	var block *Block

	err := i.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		encodedBlock := b.Get(i.currentHash)
		block = DeserializeBlock(encodedBlock)
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	i.currentHash = block.PrevBlockHash
	return block
}
