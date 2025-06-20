package db

import (
	"fmt"
	"goilerplate/config"
	"net/http"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/sirupsen/logrus"
)

func elasticDB(cfg *config.Config, log *logrus.Logger) *elasticsearch.Client {
	if cfg.Elasticsearch.Enabled {
		var err error
		es, err := elasticsearch.NewClient(elasticsearch.Config{
			Addresses: cfg.Elasticsearch.Addresses,
			Username:  cfg.Elasticsearch.Username,
			Password:  cfg.Elasticsearch.Password,
			Transport: &http.Transport{
				ResponseHeaderTimeout: 5 * time.Second,
			},
		})
		if err != nil {
			log.Fatalf("Error creating Elasticsearch client: %v", err)
		}

		res, err := es.Info()
		if err != nil {
			log.Fatalf("Error pinging Elasticsearch: %v", err)
		}
		if res.IsError() {
			log.Fatalf("Elasticsearch connection error: %s", res.String())
		}

		indices := []string{
			cfg.Elasticsearch.ApiIncomingLogIndex,
			cfg.Elasticsearch.ApiOutgoingLogIndex,
			cfg.Elasticsearch.ErrorLogIndex,
		}

		for _, indexName := range indices {
			if err := ensureIndexExists(es, indexName); err != nil {
				log.Fatalf("Failed to ensure index '%s' exists: %v", indexName, err)
			}
		}

		return es
	}

	return nil
}

func ensureIndexExists(es *elasticsearch.Client, indexName string) error {
	res, err := es.Indices.Exists([]string{indexName})
	if err != nil {
		return fmt.Errorf("error checking index existence: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusOK {
		return nil
	}

	res, err = es.Indices.Create(indexName)
	if err != nil {
		return fmt.Errorf("error creating index: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("elasticsearch error creating index: %s", res.String())
	}

	return nil
}
