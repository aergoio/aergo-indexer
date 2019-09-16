package indexer

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/aergoio/aergo-indexer/indexer/category"
	doc "github.com/aergoio/aergo-indexer/indexer/documents"
	"github.com/aergoio/aergo-indexer/indexer/transaction"
	"github.com/aergoio/aergo-indexer/types"
	"github.com/golang/protobuf/proto"
	"github.com/mr-tron/base58/base58"
)

// ConvBlock converts Block from RPC into Elasticsearch type
func (ns *Indexer) ConvBlock(block *types.Block) doc.EsBlock {
	return doc.EsBlock{
		BaseEsType: &doc.BaseEsType{base58.Encode(block.Hash)},
		Timestamp:  time.Unix(0, block.Header.Timestamp),
		BlockNo:    block.Header.BlockNo,
		TxCount:    uint(len(block.Body.Txs)),
		Size:       int64(proto.Size(block)),
	}
}

func isInternalName(name string) bool {
	switch name {
	case
		"aergo.name",
		"aergo.system",
		"aergo.enterprise":
		return true
	}
	return false
}

func encodeAccount(account []byte) string {
	if account == nil {
		return ""
	}
	if len(account) <= 12 || isInternalName(string(account)) {
		return string(account)
	}
	return types.EncodeAddress(account)
}

func (ns *Indexer) encodeAndResolveAccount(account []byte, blockNo uint64) string {
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
func (ns *Indexer) ConvTx(tx *types.Tx, blockNo uint64) doc.EsTx {
	account := ns.encodeAndResolveAccount(tx.Body.Account, blockNo)
	recipient := ns.encodeAndResolveAccount(tx.Body.Recipient, blockNo)
	amount := big.NewInt(0).SetBytes(tx.GetBody().Amount).String()
	payload0 := ""
	if len(tx.Body.Payload) > 0 {
		payload0 = fmt.Sprintf("%d", tx.Body.Payload[0])
	}
	doc := doc.EsTx{
		BaseEsType: &doc.BaseEsType{base58.Encode(tx.Hash)},
		Account:    account,
		Recipient:  recipient,
		Amount:     amount,
		Type:       fmt.Sprintf("%d", tx.Body.Type),
		Payload0:   payload0,
		Category:   category.DetectTxCategory(tx),
	}
	return doc
}

// ConvNameTx parses a name transaction into Elasticsearch type
func (ns *Indexer) ConvNameTx(tx *types.Tx, blockNo uint64) doc.EsName {
	var name = "error"
	var address string
	payload, err := transaction.UnmarshalPayloadWithArgs(tx)
	if err == nil {
		name = payload.Args[0]
		if payload.Name == "v1createName" {
			address = ns.encodeAndResolveAccount(tx.Body.Account, blockNo)
		}
		if payload.Name == "v1updateName" {
			address = payload.Args[1]
		}
	}
	hash := base58.Encode(tx.Hash)
	return doc.EsName{
		BaseEsType: &doc.BaseEsType{fmt.Sprintf("%s-%s", name, hash)},
		Name:       name,
		Address:    address,
		UpdateTx:   hash,
	}
}
