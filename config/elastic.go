package config

import (
	"fmt"
	"os"
	"strings"

	elastic "github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

var es *elastic.Client

func CreateElasticConnection() *elastic.Client {

	var err error

	app := GetAppVariable()

	cfg := elastic.Config{
		Addresses: []string{app.ElasticHost.(string) + ":" + app.ElasticPort.(string)},
	}

	es, err = elastic.NewClient(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to elasticsearch: %v\n", err)
		os.Exit(1)
	}

	err = apiLog()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create index %s: %v\n", "api-log", err)
		os.Exit(1)
	}

	err = errorLog()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create index %s: %v\n", "error-log", err)
		os.Exit(1)
	}

	err = curlLog()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create index %s: %v\n", "curl-log", err)
		os.Exit(1)
	}

	err = appLog()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create index %s: %v\n", "app-log", err)
		os.Exit(1)
	}

	fmt.Println("Elasticsearch: connected")
	fmt.Println()

	return GetElasticConnection()
}

func apiLog() error {
	index := "api-log"
	mapping := `{
		"mappings": {
			"properties": {
				"timestamp": {
					"type": "date",
					"fields": {
						"keyword": {
							"type": "keyword"
						}
					}
				},
				"request_id": {
					"type": "text",
					"fields": {
						"keyword": {
							"type": "keyword"
						}
					}
				}
			}
		}
	}`

	if _, err := createIndices(index, mapping); err != nil {
		return fmt.Errorf("error create index %s: %v", index, err)
	}

	return nil
}

func errorLog() error {
	index := "error-log"
	mapping := `{
		"mappings": {
			"properties": {
				"timestamp": {
					"type": "date",
					"fields": {
						"keyword": {
							"type": "keyword"
						}
					}
				},
				"request_id": {
					"type": "text",
					"fields": {
						"keyword": {
							"type": "keyword"
						}
					}
				},
				"activity": {
					"type": "text",
					"fields": {
						"keyword": {
							"type": "keyword"
						}
					}
				}
			}
		}
	}`

	if _, err := createIndices(index, mapping); err != nil {
		return fmt.Errorf("error create index %s: %v", index, err)
	}

	return nil
}

func curlLog() error {
	index := "curl-log"
	mapping := `{
		"mappings": {
			"properties": {
				"timestamp": {
					"type": "date",
					"fields": {
						"keyword": {
							"type": "keyword"
						}
					}
				}
			}
		}
	}`

	if _, err := createIndices(index, mapping); err != nil {
		return fmt.Errorf("error create index %s: %v", index, err)
	}

	return nil
}

func appLog() error {
	index := "app-log"
	mapping := `{
		"mappings": {
			"properties": {
				"timestamp": {
					"type": "date",
					"fields": {
						"keyword": {
							"type": "keyword"
						}
					}
				}
			}
		}
	}`

	if _, err := createIndices(index, mapping); err != nil {
		return fmt.Errorf("error create index %s: %v", index, err)
	}

	return nil
}

func createIndices(index string, mapping string) (*esapi.Response, error) {

	res, err := es.Indices.Create(index, func(a *esapi.IndicesCreateRequest) {
		a.Body = strings.NewReader(mapping)
	})

	if err != nil {
		return nil, fmt.Errorf("error %v", err)
	}

	if res.StatusCode == 400 {
		// fmt.Printf("%s not created because already exists\n", index)
	} else {
		fmt.Print("Elasticsearch creating index: ")
		fmt.Printf("%s created\n", index)
	}

	return res, nil
}

func GetElasticConnection() *elastic.Client {
	return es
}
