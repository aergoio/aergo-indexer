package esindexer

import (
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"github.com/aergoio/aergo-esindexer/types"
	"github.com/mr-tron/base58/base58"
)

// EsType is an interface for structs to be used as ES documents
type EsType interface {
	GetID() string
}

// BaseEsType implements EsType and contains the document's id
type BaseEsType struct {
	id string
}

// GetID returns the document's id
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
	Type      string    `json:"type"`
	Payload0  string    `json:"payload0"` // first byte of payload
}

// EsName is a name-address mapping stored in elasticsearch
type EsName struct {
	*BaseEsType
	Name        string `json:"name"`
	Address     string `json:"address"`
	UpdateBlock uint64 `json:"blockno"`
	UpdateTx    string `json:"tx"`
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

func encodeAccount(account []byte) string {
	if account == nil {
		return ""
	}
	if len(account) <= 12 {
		return string(account)
	}
	return types.EncodeAddress(account)
}

// ConvTx converts Tx from RPC into Elasticsearch type
func ConvTx(tx *types.Tx) EsTx {
	account := encodeAccount(tx.Body.Account)
	recipient := encodeAccount(tx.Body.Recipient)
	amount := big.NewInt(0).SetBytes(tx.GetBody().Amount).String()
	payload0 := ""
	if len(tx.Body.Payload) > 0 {
		payload0 = fmt.Sprintf("%d", tx.Body.Payload[0])
	}
	doc := EsTx{
		BaseEsType: &BaseEsType{base58.Encode(tx.Hash)},
		Account:    account,
		Recipient:  recipient,
		Amount:     amount,
		Type:       fmt.Sprintf("%d", tx.Body.Type),
		Payload0:   payload0,
	}
	return doc
}

type txPayload struct {
	Name string   `json:"Name"`
	Args []string `json:"Args"`
}

// ConvNameTx parses a name transaction into Elasticsearch type
func ConvNameTx(tx *types.Tx) EsName {
	var name = "error"
	var address string
	payloadSource := tx.GetBody().GetPayload()
	payload := new(txPayload)
	if err := json.Unmarshal(payloadSource, payload); err == nil {
		if payload.Name == "v1createName" {
			name = payload.Args[0]
			address = types.EncodeAddress(tx.Body.Account)
		}
		if payload.Name == "v1updateName" {
			name = payload.Args[0]
			address = payload.Args[1]
		}
	}
	hash := base58.Encode(tx.Hash)
	return EsName{
		BaseEsType: &BaseEsType{fmt.Sprintf("%s-%s", name, hash)},
		Name:       name,
		Address:    address,
		UpdateTx:   hash,
	}
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
					},
					"type": {
						"type": "keyword"
					},
					"payload0": {
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
	"name": `{
		"mappings":{
			"name":{
				"properties": {
					"name": {
						"type": "keyword"
					},
					"address": {
						"type": "keyword"
					},
					"blockno": {
						"type": "long"
					},
					"tx": {
						"type": "keyword"
					}
				}
			}
		}
	}`,
}
