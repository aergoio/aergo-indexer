package category

import (
	"strings"

	"github.com/aergoio/aergo-indexer/indexer/transaction"
	"github.com/aergoio/aergo-indexer/types"
)

// TxCategory is a user-friendly categorization of a transaction
type TxCategory string

// Categories
const (
	None       TxCategory = ""
	Payload    TxCategory = "payload"
	Call       TxCategory = "call"
	Governance TxCategory = "governance"
	System     TxCategory = "system"
	Staking    TxCategory = "staking"
	Voting     TxCategory = "voting"
	Name       TxCategory = "name"
	NameCreate TxCategory = "namecreate"
	NameUpdate TxCategory = "nameupdate"
	Enterprise TxCategory = "enterprise"
	Conf       TxCategory = "conf"
	Cluster    TxCategory = "cluster"
	Deploy     TxCategory = "deploy"
	Redeploy   TxCategory = "redeploy"
)

// TxCategories is the list of available categories in order of increasing weight
var TxCategories = []TxCategory{None, Payload, Call, Governance, System, Staking, Voting, Name, NameCreate, NameUpdate, Enterprise, Conf, Cluster, Deploy, Redeploy}

// DetectTxCategory by performing a cascade of checks with fallbacks
func DetectTxCategory(tx *types.Tx) TxCategory {
	txBody := tx.GetBody()
	txType := txBody.GetType()
	txRecipient := string(txBody.GetRecipient())
	/* Not merged yet
	if txType == types.TxType_REDEPLOY {
		return Redeploy
	} */
	if txRecipient == "" && len(txBody.Payload) > 0 {
		return Deploy
	}
	if txRecipient == "aergo.enterprise" {
		txCallName, err := transaction.GetCallName(tx)
		if err == nil {
			txCallName = strings.ToLower(txCallName)
			if strings.HasSuffix(txCallName, "cluster") {
				return Cluster
			}
			if strings.HasSuffix(txCallName, "conf") {
				return Conf
			}
		}
		return Enterprise
	}
	if txRecipient == "aergo.name" {
		txCallName, err := transaction.GetCallName(tx)
		if err == nil {
			txCallName = strings.ToLower(txCallName)
			if strings.HasSuffix(txCallName, "updatename") {
				return NameUpdate
			}
			if strings.HasSuffix(txCallName, "createname") {
				return NameCreate
			}
		}
		return Name
	}
	if txRecipient == "aergo.system" {
		txCallName, err := transaction.GetCallName(tx)
		if err == nil {
			txCallName = strings.ToLower(txCallName)
			if strings.HasSuffix(txCallName, "stake") || strings.HasSuffix(txCallName, "unstake") {
				return Staking
			}
			if strings.HasSuffix(txCallName, "vote") || strings.HasSuffix(txCallName, "votebp") || strings.HasSuffix(txCallName, "proposal") {
				return Voting
			}
		}
		return System
	}
	if txType == types.TxType_GOVERNANCE {
		return Governance
	}
	txCallName, err := transaction.GetCallName(tx)
	if err == nil && txCallName != "" {
		return Call
	}
	if len(txBody.Payload) > 0 {
		return Payload
	}
	return None
}
