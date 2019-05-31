package esindexer

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"github.com/aergoio/aergo-esindexer/types"
	"github.com/golang/protobuf/proto"
	"github.com/mr-tron/base58/base58"
)

func isInternalName(name string) bool {
	switch name {
	case
		"aergo.name",
		"aergo.system":
		return true
	}
	return false
}

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
	Size      int64     `json:"size"`
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
func (ns *EsIndexer) ConvBlock(block *types.Block) EsBlock {
	return EsBlock{
		BaseEsType: &BaseEsType{base58.Encode(block.Hash)},
		Timestamp:  time.Unix(0, block.Header.Timestamp),
		BlockNo:    block.Header.BlockNo,
		TxCount:    uint(len(block.Body.Txs)),
		Size:       int64(proto.Size(block)),
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

func (ns *EsIndexer) encodeAndResolveAccount(account []byte, blockNo uint64) string {
	var encoded = encodeAccount(account)
	if len(encoded) > 12 || isInternalName(encoded) || encoded == "" {
		return encoded
	}
	// Resolve name
	var nameRequest = &types.Name{
		Name:    encoded,
		BlockNo: blockNo,
	}
	ctx := context.Background()
	nameInfo, err := ns.grpcClient.GetNameInfo(ctx, nameRequest)
	if err != nil {
		return "UNRESOLVED: " + encoded
	}
	return encodeAccount(nameInfo.GetDestination())
}

// ConvTx converts Tx from RPC into Elasticsearch type
func (ns *EsIndexer) ConvTx(tx *types.Tx, blockNo uint64) EsTx {
	account := ns.encodeAndResolveAccount(tx.Body.Account, blockNo)
	recipient := ns.encodeAndResolveAccount(tx.Body.Recipient, blockNo)
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
func (ns *EsIndexer) ConvNameTx(tx *types.Tx, blockNo uint64) EsName {
	var name = "error"
	var address string
	payloadSource := tx.GetBody().GetPayload()
	payload := new(txPayload)
	if err := json.Unmarshal(payloadSource, payload); err == nil {
		name = payload.Args[0]
		if payload.Name == "v1createName" {
			address = ns.encodeAndResolveAccount(tx.Body.Account, blockNo)
		}
		if payload.Name == "v1updateName" {
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
