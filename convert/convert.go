package convert

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

type inputFormatType interface {
	isLineByLine() bool
	convert(data []byte) (interface{}, error)
	init(args []string) error
}

type outputFormatType interface {
	isLineByLine() bool
	convert(data interface{}) ([]byte, error)
	init(args []string) error
}

var inputFormats = map[string]inputFormatType{
	"json":  jsonIn{},
	"jsonl": jsonlIn{},
	"regex": regexIn{},
	//"yaml" : yamlIn{},
	//toml" : tomlIn{},
	//"xml" : xmlIn{},
	//"csv" : csvIn{},
	//"ini" : iniIn{},
	//"json5" : json5In{},
	//hcl" : hclIn{},
	//"properties" : propertiesIn{},
	//"dotenv" : dotenvIn{},
}

var outputFormats = map[string]outputFormatType{
	"json":  jsonOut{},
	"jsonl": jsonlOut{},
}

// Convert converts data from one format to another.
func Convert(from string, fromArgs []string, to string, toArgs []string, in, out *os.File) error {
	inputFormat, ok := inputFormats[from]
	if !ok {
		return fmt.Errorf("invalid input format: %s", from)
	}
	err := inputFormat.init(fromArgs)
	if err != nil {
		return fmt.Errorf("error initializing input format: %w", err)
	}
	outputFormat, ok := outputFormats[to]
	if !ok {
		return fmt.Errorf("invalid output format: %s", to)
	}
	err = outputFormat.init(toArgs)
	if err != nil {
		return fmt.Errorf("error initializing output format: %w", err)
	}
	if inputFormat.isLineByLine() && outputFormat.isLineByLine() {
		scanner := bufio.NewScanner(in)
		for scanner.Scan() {
			lineData, err := inputFormat.convert([]byte(scanner.Text()))
			if err != nil {
				return fmt.Errorf("error converting line in: %w", err)
			}
			outputData, err := outputFormat.convert(lineData)
			if err != nil {
				return fmt.Errorf("error converting line out: %w", err)
			}
			_, err = out.Write(outputData)
		}
	} else if inputFormat.isLineByLine() {
		scanner := bufio.NewScanner(in)
		var lines []interface{}
		for scanner.Scan() {
			lineData, err := inputFormat.convert([]byte(scanner.Text()))
			if err != nil {
				return fmt.Errorf("error converting line in: %w", err)
			}
			lines = append(lines, lineData)
		}
		outputData, err := outputFormat.convert(lines)
		if err != nil {
			return fmt.Errorf("error converting lines out: %w", err)
		}
		_, err = out.Write(outputData)
	} else if outputFormat.isLineByLine() {
		inputData, err := io.ReadAll(in)
		if err != nil {
			return fmt.Errorf("error reading input: %w", err)
		}
		lines, err := inputFormat.convert(inputData)
		if err != nil {
			return fmt.Errorf("error converting input: %w", err)
		}
		// TODO check if is array, if not error
		for _, line := range lines.([]interface{}) {
			outputData, err := outputFormat.convert(line)
			if err != nil {
				return fmt.Errorf("error converting line out: %w", err)
			}
			_, err = out.Write(outputData)
		}
	} else {
		// simple conversion
		input, err := io.ReadAll(in)
		if err != nil {
			return fmt.Errorf("error reading input: %w", err)
		}
		inputData, err := inputFormat.convert(input)
		if err != nil {
			return fmt.Errorf("error converting input: %w", err)
		}
		outputData, err := outputFormat.convert(inputData)
		if err != nil {
			return fmt.Errorf("error converting output: %w", err)
		}
		_, err = out.Write(outputData)
	}
	return nil
}
