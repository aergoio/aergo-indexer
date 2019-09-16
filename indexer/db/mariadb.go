package db

import (
	doc "github.com/aergoio/aergo-indexer/indexer/documents"
)

type MariaDbController struct {
	//mariaClient    *client.Client
}

func (c MariaDbController) Insert(document doc.DocType, params QueryParams) (uint64, error) {
	return 0, nil
}

func (c MariaDbController) InsertBulk(documentChannel chan doc.DocType, params QueryParams) (uint64, error) {
	return 0, nil
}
