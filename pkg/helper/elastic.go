package helper

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/elastic/go-elasticsearch/v8"
)

type elasticClient struct {
	Ctx    context.Context
	Client *elasticsearch.Client
}

func NewElastic(ctx context.Context, es *elasticsearch.Client) *elasticClient {
	return &elasticClient{
		Ctx:    ctx,
		Client: es,
	}
}

func (ec *elasticClient) Post(indexName string, document map[string]any) error {
	body, err := json.Marshal(document)
	if err != nil {
		return fmt.Errorf("failed to marshal document: %w", err)
	}

	res, err := ec.Client.Index(
		indexName,
		bytes.NewReader(body),
		ec.Client.Index.WithContext(ec.Ctx),
	)
	if err != nil {
		return fmt.Errorf("failed to index document: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("elasticsearch error indexing document: %s", res.String())
	}

	return nil
}

func (ec *elasticClient) Put(indexName string, id string, document map[string]interface{}) error {
	body, err := json.Marshal(document)
	if err != nil {
		return fmt.Errorf("failed to marshal document: %w", err)
	}

	res, err := ec.Client.Index(
		indexName,
		bytes.NewReader(body),
		ec.Client.Index.WithDocumentID(id),
		ec.Client.Index.WithContext(ec.Ctx),
		ec.Client.Index.WithOpType("index"), // ensure it's PUT-like, not create-only
	)
	if err != nil {
		return fmt.Errorf("failed to PUT document: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("elasticsearch error in PUT: %s", res.String())
	}

	return nil
}

func (ec *elasticClient) Search(indexName string, query map[string]interface{}) ([]byte, error) {
	body, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query: %w", err)
	}

	res, err := ec.Client.Search(
		ec.Client.Search.WithContext(ec.Ctx),
		ec.Client.Search.WithIndex(indexName),
		ec.Client.Search.WithBody(bytes.NewReader(body)),
		ec.Client.Search.WithTrackTotalHits(true),
	)
	if err != nil {
		return nil, fmt.Errorf("search request failed: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("elasticsearch error in search: %s", res.String())
	}

	result, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read search response: %w", err)
	}

	return result, nil
}
