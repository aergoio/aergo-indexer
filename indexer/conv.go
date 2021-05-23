package indexer

import (
	"context"
	"fmt"
	"math/big"
	"strings"
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
	rewardAmount := ""
	if len(block.Header.Consensus) > 0 {
		rewardAmount = "160000000000000000"
	}
	return doc.EsBlock{
		BaseEsType:    &doc.BaseEsType{base58.Encode(block.Hash)},
		Timestamp:     time.Unix(0, block.Header.Timestamp),
		BlockNo:       block.Header.BlockNo,
		TxCount:       uint(len(block.Body.Txs)),
		Size:          int64(proto.Size(block)),
		RewardAccount: ns.encodeAndResolveAccount(block.Header.Consensus, block.Header.BlockNo),
		RewardAmount:  rewardAmount,
	}
}

// Internal names refer to special accounts that don't need to be resolved
func isInternalName(name string) bool {
	switch name {
	case
		"aergo.name",
		"aergo.system",
		"aergo.enterprise",
		"aergo.vault":
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

// bigIntToFloat takes a big.Int, divides it by 10^exp and returns the resulting float
// Note that this float is not precise. It can be used for sorting purposes
func bigIntToFloat(a *big.Int, exp int64) float32 {
	var y, e = big.NewInt(10), big.NewInt(exp)
	y.Exp(y, e, nil)
	z := new(big.Float).Quo(
		new(big.Float).SetInt(a),
		new(big.Float).SetInt(y),
	)
	f, _ := z.Float32()
	return f
}

// ConvTx converts Tx from RPC into Elasticsearch type
func (ns *Indexer) ConvTx(tx *types.Tx, blockNo uint64) doc.EsTx {
	account := ns.encodeAndResolveAccount(tx.Body.Account, blockNo)
	recipient := ns.encodeAndResolveAccount(tx.Body.Recipient, blockNo)
	amount := big.NewInt(0).SetBytes(tx.GetBody().Amount)
	doc := doc.EsTx{
		BaseEsType:  &doc.BaseEsType{base58.Encode(tx.Hash)},
		Account:     account,
		Recipient:   recipient,
		Amount:      amount.String(),
		AmountFloat: bigIntToFloat(amount, 18),
		Type:        fmt.Sprintf("%d", tx.Body.Type),
		Category:    category.DetectTxCategory(tx),
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

// ConvContractCreateTx creates document for token creation
func (ns *Indexer) ConvContractCreateTx(tx *types.Tx, txDoc doc.EsTx, receipt *types.Receipt) doc.EsToken {
	address := base58.Encode(receipt.ContractAddress)
	return doc.EsToken{
		BaseEsType: &doc.BaseEsType{address},
		TxId:       txDoc.GetID(),
		Type:       category.ARC1, // TODO
		Name:       "name",        // TODO
		Symbol:     "symbol",      // TODO
		Decimals:   18,            // TODO
		Supply:     "10101010",    // TODO
	}
}

// MaybeTokenCreation runs a heuristic to determine if tx might be creating a token
func (ns *Indexer) MaybeTokenCreation(tx *types.Tx) bool {
	txBody := tx.GetBody()
	isDeploy := len(txBody.GetRecipient()) == 0 && len(txBody.Payload) > 0
	if !isDeploy {
		return false
	}
	// We treat the payload (which is part bytecode, part ABI) as text
	payload := string(txBody.GetPayload())
	keywords := [...]string{"balanceOf", "transfer", "symbol"}
	for _, keyword := range keywords {
		if !strings.Contains(payload, keyword) {
			return false
		}
	}
	return true
}
