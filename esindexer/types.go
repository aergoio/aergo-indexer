package esindexer

import (
	"math/big"
	"time"

	"github.com/aergoio/aergo-esindexer/types"
	"github.com/mr-tron/base58/base58"
)

type EsType interface {
	GetID() string
}

type BaseEsType struct {
	id string
}

func (m BaseEsType) GetID() string {
	return m.id
}

// EsBlock is a block stored in elasticsearch
type EsBlock struct {
	*BaseEsType
	Timestamp time.Time `json:"ts"`
	BlockNo   uint64    `json:"no"`
	TxCount   uint      `json:"txs"`
}

// EsTx is a transaction stored in elasticsearch
type EsTx struct {
	*BaseEsType
	Timestamp time.Time `json:"ts"`
	BlockNo   uint64    `json:"blockno"`
	Account   string    `json:"from"`
	Recipient string    `json:"to"`
	Amount    string    `json:"amount"` // string of BigInt
}

// ConvBlock converts Block from RPC into Elasticsearch type
func ConvBlock(block *types.Block) EsBlock {
	return EsBlock{
		BaseEsType: &BaseEsType{base58.Encode(block.Hash)},
		Timestamp:  time.Unix(0, block.Header.Timestamp),
		BlockNo:    block.Header.BlockNo,
		TxCount:    uint(len(block.Body.Txs)),
	}
}

// ConvTx converts Tx from RPC into Elasticsearch type
func ConvTx(tx *types.Tx) EsTx {
	account := ""
	if tx.Body.Account != nil {
		account = types.EncodeAddress(tx.Body.Account)
	}
	recipient := ""
	if tx.Body.Recipient != nil {
		recipient = types.EncodeAddress(tx.Body.Recipient)
	}
	amount := big.NewInt(0).SetBytes(tx.GetBody().Amount).String()
	doc := EsTx{
		BaseEsType: &BaseEsType{base58.Encode(tx.Hash)},
		Account:    account,
		Recipient:  recipient,
		Amount:     amount,
	}
	return doc
}

// EsMappings contains the elasticsearch mappings
var mappings = map[string]string{
	"tx": `{
		"mappings":{
			"tx":{
				"properties":{
					"ts": {
						"type": "date"
					},
					"blockno": {
						"type": "long"
					},
					"from": {
						"type": "keyword"
					},
					"to": {
						"type": "keyword"
					},
					"amount": {
						"type": "keyword"
					}
				}
			}
		}
	}`,
	"block": `{
		"mappings":{
			"block":{
				"properties": {
					"ts": {
						"type": "date"
					},
					"no": {
						"type": "long"
					},
					"txs": {
						"type": "long"
					}
				}
			}
		}
	}`,
}
