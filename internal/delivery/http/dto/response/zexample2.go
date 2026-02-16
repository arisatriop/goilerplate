package dtoresponse

type Example2Response struct {
	SomethingField1 string   `json:"somethingField1"`
	SomethingField2 string   `json:"somethingField2"`
	SomethingElse   []string `json:"somethingElse"`
}
