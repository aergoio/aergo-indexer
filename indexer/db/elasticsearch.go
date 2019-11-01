package db

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	doc "github.com/aergoio/aergo-indexer/indexer/documents"
	"github.com/olivere/elastic"
)

// ElasticsearchDbController implements DbController
type ElasticsearchDbController struct {
	Client *elastic.Client
}

// NewElasticClient creates a new instance of elastic.Client
func NewElasticClient(esURL string) (*elastic.Client, error) {
	url := esURL
	if !strings.HasPrefix(url, "http") {
		url = fmt.Sprintf("http://%s", url)
	}
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	httpClient := &http.Client{Transport: tr}
	client, err := elastic.NewClient(
		elastic.SetHttpClient(httpClient),
		elastic.SetURL(url),
		elastic.SetHealthcheckTimeoutStartup(30*time.Second),
		elastic.SetSniff(false),
	)
	if err != nil {
		return nil, err
	}
	return client, nil
}

// NewElasticsearchDbController creates a new instance of ElasticsearchDbController
func NewElasticsearchDbController(esURL string) (*ElasticsearchDbController, error) {
	client, err := NewElasticClient(esURL)
	if err != nil {
		return nil, err
	}
	return &ElasticsearchDbController{
		Client: client,
	}, nil
}

func getFirstError(res *elastic.BulkResponse) error {
	if res.Errors {
		for _, v := range res.Items {
			for action, item := range v {
				if item.Error != nil {
					resJSON, _ := json.Marshal(item.Error)
					return fmt.Errorf("%s %s (%s): %s", action, item.Type, item.Id, string(resJSON))
				}
			}
		}
	}
	return nil
}

// Insert inserts a single document using the updata params
// It returns the number of inserted documents (1) or an error
func (esdb *ElasticsearchDbController) Insert(document doc.DocType, params UpdateParams) (uint64, error) {
	ctx := context.Background()
	_, err := esdb.Client.Index().Index(params.IndexName).Type(params.TypeName).Id(document.GetID()).BodyJson(document).Do(ctx)
	if err != nil {
		return 0, err
	}
	return 1, nil
}

// InsertBulk inserts documents arriving in documentChannel in bulk using the updata params
// It returns the number of inserted documents or an error
func (esdb *ElasticsearchDbController) InsertBulk(documentChannel chan doc.DocType, params UpdateParams) (uint64, error) {
	ctx := context.Background()
	var total uint64
	bulk := esdb.Client.Bulk().Index(params.IndexName).Type(params.TypeName)

	begin := time.Now()
	commitBulk := func() error {
		res, err := bulk.Do(ctx)
		dur := time.Since(begin).Seconds()
		pps := int64(float64(total) / dur)
		logger.Info().Int("chunkSize", params.Size).Uint64("total", total).Int64("perSecond", pps).Str("indexName", params.IndexName).Msg("Comitted bulk chunk")
		if err == nil {
			err = getFirstError(res)
		}
		if err != nil {
			return err
		}
		return nil
	}
	for d := range documentChannel {
		atomic.AddUint64(&total, 1)
		if params.Upsert {
			bulk.Add(elastic.NewBulkIndexRequest().Id(d.GetID()).Doc(d))
		} else {
			bulk.Add(elastic.NewBulkUpdateRequest().Id(d.GetID()).Doc(d).DocAsUpsert(true))
		}
		if bulk.NumberOfActions() >= params.Size {
			err := commitBulk()
			if err != nil {
				return total, err
			}
		}

		select {
		default:
		case <-ctx.Done():
			return total, ctx.Err()
		}
	}

	// Commit the final batch before exiting
	if bulk.NumberOfActions() > 0 {
		err := commitBulk()
		if err != nil {
			return total, err
		}
	}
	return total, nil
}

// Delete removes documents specified by the query params
func (esdb *ElasticsearchDbController) Delete(params QueryParams) (uint64, error) {
	var query *elastic.RangeQuery
	if params.IntegerRange != nil {
		query = elastic.NewRangeQuery(params.IntegerRange.Field).From(params.IntegerRange.Min).To(params.IntegerRange.Max)
	}
	if params.StringMatch != nil {
		return 0, errors.New("Delete is not imlemented for string matches")
	}

	ctx := context.Background()
	res, err := esdb.Client.DeleteByQuery().Index(params.IndexName).Query(query).Do(ctx)
	if err != nil {
		return 0, err
	}
	return uint64(res.Deleted), nil
}

// Count returns the number of indexed documents
func (esdb *ElasticsearchDbController) Count(params QueryParams) (int64, error) {
	ctx := context.Background()
	return esdb.Client.Count(params.IndexName).Do(ctx)
}

// SelectOne selects a single document
func (esdb *ElasticsearchDbController) SelectOne(params QueryParams, createDocument CreateDocFunction) (doc.DocType, error) {
	ctx := context.Background()
	query := elastic.NewMatchAllQuery()
	res, err := esdb.Client.Search().Index(params.IndexName).Query(query).Sort(params.SortField, params.SortAsc).From(params.From).Size(1).Do(ctx)
	if err != nil {
		return nil, err
	}
	if res == nil || res.TotalHits() == 0 || len(res.Hits.Hits) == 0 {
		return nil, nil
	}
	// Unmarshall document
	hit := res.Hits.Hits[0]
	document := createDocument()
	if err := json.Unmarshal(*hit.Source, document); err != nil {
		return nil, err
	}
	document.SetID(hit.Id)
	if err != nil {
		return nil, err
	}
	return document, nil
}

// UpdateAlias updates an alias with a new index name and delete stale indices
func (esdb *ElasticsearchDbController) UpdateAlias(aliasName string, indexName string) error {
	ctx := context.Background()
	svc := esdb.Client.Alias()
	res, err := esdb.Client.Aliases().Index("_all").Do(ctx)
	if err != nil {
		return err
	}
	indices := res.IndicesByAlias(aliasName)
	if len(indices) > 0 {
		// Remove old aliases
		for _, indexName := range indices {
			svc.Remove(indexName, aliasName)
		}
	}
	// Add new alias
	svc.Add(indexName, aliasName)
	_, err = svc.Do(ctx)
	// Delete old indices
	if len(indices) > 0 {
		for _, indexName := range indices {
			esdb.Client.DeleteIndex(indexName).Do(ctx)
		}
	}
	return err
}

// GetExistingIndexPrefix checks for existing indices and returns the prefix, if any
func (esdb *ElasticsearchDbController) GetExistingIndexPrefix(aliasName string, documentType string) (bool, string, error) {
	ctx := context.Background()
	res, err := esdb.Client.Aliases().Index("_all").Do(ctx)
	if err != nil {
		return false, "", err
	}
	indices := res.IndicesByAlias(aliasName)
	if len(indices) > 0 {
		indexNamePrefix := strings.TrimRight(indices[0], documentType)
		return true, indexNamePrefix, nil
	}
	return false, "", nil
}

// CreateIndex creates index according to documentType definition
func (esdb *ElasticsearchDbController) CreateIndex(indexName string, documentType string) error {
	ctx := context.Background()
	createIndex, err := esdb.Client.CreateIndex(indexName).BodyString(doc.EsMappings[documentType]).Do(ctx)
	if err != nil {
		return err
	}
	if !createIndex.Acknowledged {
		return errors.New("CreateIndex not acknowledged")
	}
	return nil
}

// Scroll creates a new scroll instance with the specified query and unmarshal function
func (esdb *ElasticsearchDbController) Scroll(params QueryParams, createDocument CreateDocFunction) ScrollInstance {
	fsc := elastic.NewFetchSourceContext(true).Include(params.SelectFields...)
	scroll := esdb.Client.Scroll(params.IndexName).Type(params.TypeName).Size(params.Size).Sort(params.SortField, params.SortAsc).FetchSourceContext(fsc)
	return &EsScrollInstance{
		scrollService:  scroll,
		ctx:            context.Background(),
		createDocument: createDocument,
	}
}

// EsScrollInstance is an instance of a scroll for ES
type EsScrollInstance struct {
	scrollService  *elastic.ScrollService
	result         *elastic.SearchResult
	current        int
	currentLength  int
	ctx            context.Context
	createDocument CreateDocFunction
}

// Next returns the next document of a scroll or io.EOF
func (scroll *EsScrollInstance) Next() (doc.DocType, error) {
	// Load next part of scroll
	if scroll.result == nil || scroll.current >= scroll.currentLength {
		result, err := scroll.scrollService.Do(scroll.ctx)
		if err != nil {
			return nil, err // returns io.EOF when scroll is done
		}
		scroll.result = result
		scroll.current = 0
		scroll.currentLength = len(result.Hits.Hits)
	}

	// Return next document
	if scroll.current < scroll.currentLength {
		doc := scroll.result.Hits.Hits[scroll.current]
		scroll.current++

		unmarshalled := scroll.createDocument()
		if err := json.Unmarshal(*doc.Source, unmarshalled); err != nil {
			return nil, err
		}
		unmarshalled.SetID(doc.Id)
		return unmarshalled, nil
	}

	return nil, io.EOF
}
