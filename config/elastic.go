package config

import (
	"fmt"
	"os"

	elastic "github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

var es *elastic.Client

func CreateElasticConnection() {

	var err error
	var index string = "api-log"

	cfg := elastic.Config{
		Addresses: []string{os.Getenv("ELASTIC_HOST") + ":" + os.Getenv("ELASTIC_PORT")},
	}

	es, err = elastic.NewClient(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to elasticsearch: %v\n", err)
		os.Exit(1)
	}

	res, err := createIndices(index)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error to create initial index: %v\n", err)
		os.Exit(1)
	}

	fmt.Print("Elasticsearch: ")
	if res.Status() == "" {
		fmt.Print("Initial indices was not create because already exists\n")
	} else {
		fmt.Printf("%s indices created\n", index)
	}
	fmt.Println("Elasticsearch: connected")
}

func createIndices(index string) (*esapi.Response, error) {
	res, err := es.Indices.Create(index)
	if err != nil {
		return nil, err
	}

	if res.IsError() {
		return nil, err

	}

	return res, nil

}

func GetElasticConnection() *elastic.Client {
	return es
}
