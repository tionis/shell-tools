package convert

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"tasadar.net/tionis/shell-tools/convert/regex2json"
)

type regexIn struct {
	expressions []*regex2json.Expression
	logger      *log.Logger
}

func (r regexIn) isLineByLine() bool {
	return true
}

func (r regexIn) convert(data []byte) (interface{}, error) {
	// TODO implement this
	return nil, nil
}

func (r regexIn) init(args []string) error {
	if len(args) != 1 {
		return errors.New("invalid number of arguments")
	}
	compile, err := regexp.Compile(args[0])
	if err != nil {
		return fmt.Errorf("error compiling regex: %w", err)
	}
	r.expressions, err = regex2json.CompileExpressions(compile)
	if err != nil {
		return fmt.Errorf("error compiling expressions: %w", err)
	}
	return nil
}
