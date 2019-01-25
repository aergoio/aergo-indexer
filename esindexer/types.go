package esindexer

import (
	"fmt"
	"math/big"
	"strings"
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
	Payload0  byte      `json:"payload0"` // first byte of payload
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

// ConvTx converts Tx from RPC into Elasticsearch type
func ConvTx(tx *types.Tx) EsTx {
	account := ""
	if tx.Body.Account != nil {
		account = types.EncodeAddress(tx.Body.Account)
	}
	recipient := ""
	if tx.Body.Recipient != nil {
		if len(tx.Body.Recipient) <= 12 {
			recipient = string(tx.Body.Recipient)
		} else {
			recipient = types.EncodeAddress(tx.Body.Recipient)
		}
	}
	amount := big.NewInt(0).SetBytes(tx.GetBody().Amount).String()
	var payload0 byte
	if len(tx.Body.Payload) > 0 {
		payload0 = tx.Body.Payload[0]
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

// ConvNameTx parses a name transaction into Elasticsearch type
func ConvNameTx(tx *types.Tx) EsName {
	var name string
	var address string
	payload := tx.GetBody().GetPayload()
	action := payload[0]
	if action == 'c' {
		name = string(payload[1:])
		address = types.EncodeAddress(tx.Body.Account)
	}
	if action == 'u' {
		nameByte, addressByte := parseUpdatePayload(tx.Body.GetPayload())
		name = string(nameByte)
		address = types.EncodeAddress(addressByte)
	}
	hash := base58.Encode(tx.Hash)
	return EsName{
		BaseEsType: &BaseEsType{fmt.Sprintf("%s-%s", name, hash)},
		Name:       name,
		Address:    address,
		UpdateTx:   hash,
	}
}

func parseUpdatePayload(p []byte) ([]byte, []byte) {
	comma := strings.IndexByte(string(p), ',')
	if comma < 0 {
		return nil, nil
	}
	name := p[:comma]
	to := p[comma+1:]
	return []byte(name), []byte(to)
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
						"type": "byte"
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
