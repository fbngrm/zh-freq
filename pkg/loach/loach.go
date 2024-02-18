package loach

import (
	"encoding/json"
	"fmt"
	"os"
)

func NewFrequencyIndex(src string) ([]string, error) {
	lines, err := os.ReadFile(src)
	if err != nil {
		return nil, fmt.Errorf("could not open loach index source file: %w", err)
	}
	var data []string
	err = json.Unmarshal(lines, &data)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal loach index: %w", err)
	}
	return data, nil
}
