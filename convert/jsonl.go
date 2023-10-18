package convert

import "encoding/json"

type jsonlIn struct {
}

func (j jsonlIn) isLineByLine() bool {
	return true
}

func (j jsonlIn) convert(data []byte) (interface{}, error) {
	var result interface{}
	err := json.Unmarshal(data, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (j jsonlIn) init(_ []string) error {
	return nil
}

type jsonlOut struct {
}

func (j jsonlOut) isLineByLine() bool {
	return true
}

func (j jsonlOut) convert(data interface{}) ([]byte, error) {
	return json.Marshal(data)
}

func (j jsonlOut) init(_ []string) error {
	return nil
}
