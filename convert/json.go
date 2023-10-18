package convert

import "encoding/json"

type jsonIn struct {
}

func (j jsonIn) isLineByLine() bool {
	return false
}

func (j jsonIn) convert(data []byte) (interface{}, error) {
	var result interface{}
	err := json.Unmarshal(data, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (j jsonIn) init(args []string) error {
	return nil
}

type jsonOut struct {
}

func (j jsonOut) isLineByLine() bool {
	return false
}

func (j jsonOut) convert(data interface{}) ([]byte, error) {
	return json.Marshal(data)
}

func (j jsonOut) init(args []string) error {
	return nil
}
