package db

import (
	doc "github.com/aergoio/aergo-indexer/indexer/documents"
)

type UpdateParams struct {
	IndexName string
	TypeName  string
	Upsert    bool
	Size      int
}

type QueryParams struct {
	IndexName    string
	TypeName     string
	From         int
	Size         int
	SortField    string
	SortAsc      bool
	SelectFields []string
	IntegerRange *IntegerRangeQuery
	StringMatch  *StringMatchQuery
}

type IntegerRangeQuery struct {
	Field string
	Min   uint64
	Max   uint64
}

type StringMatchQuery struct {
	Field string
	Value string
}

type DbController interface {
	Insert(document doc.DocType, params UpdateParams) (uint64, error)
	InsertBulk(documentChannel chan doc.DocType, params UpdateParams) (uint64, error)
	Delete(params QueryParams) (uint64, error)
	Count(params QueryParams) (int64, error)
	SelectOne(params QueryParams, unmarshal UnmarshalFunc) (doc.DocType, error)
	Scroll(params QueryParams, unmarshal UnmarshalFunc) ScrollInstance
}

type UnmarshalFunc = func([]byte) (doc.DocType, error)

type ScrollInstance interface {
	/*
		params QueryParams
		current    int
	*/
	Next() (doc.DocType, error)
}