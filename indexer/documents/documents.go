package documents

import (
	"time"

	"github.com/aergoio/aergo-indexer/indexer/category"
)

// DocType is an interface for structs to be used as ES documents
type DocType interface {
	GetID() string
	SetID(string)
}

// BaseEsType implements DocType and contains the document's id
type BaseEsType struct {
	Id string
}

// GetID returns the document's id
func (m BaseEsType) GetID() string {
	return m.Id
}

// SetID sets the document's id
func (m BaseEsType) SetID(id string) {
	m.Id = id
}

// EsBlock is a block stored in elasticsearch
type EsBlock struct {
	*BaseEsType
	Timestamp time.Time `json:"ts"`
	BlockNo   uint64    `json:"no"`
	TxCount   uint      `json:"txs"`
	Size      int64     `json:"size"`
}

// EsTx is a transaction stored in elasticsearch
type EsTx struct {
	*BaseEsType
	Timestamp time.Time           `json:"ts"`
	BlockNo   uint64              `json:"blockno"`
	Account   string              `json:"from"`
	Recipient string              `json:"to"`
	Amount    string              `json:"amount"` // string of BigInt
	Type      string              `json:"type"`
	Payload0  string              `json:"payload0"` // first byte of payload
	Category  category.TxCategory `json:"category"`
}

// EsName is a name-address mapping stored in elasticsearch
type EsName struct {
	*BaseEsType
	Name        string `json:"name"`
	Address     string `json:"address"`
	UpdateBlock uint64 `json:"blockno"`
	UpdateTx    string `json:"tx"`
}

// EsMappings contains the elasticsearch mappings
var Mappings = map[string]string{
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
					},
					"category": {
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
					},
					"size": {
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
