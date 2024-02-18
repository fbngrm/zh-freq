package frequency

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"golang.org/x/exp/slog"
)

type WordIndex struct {
	path  string
	Words []string
}

func NewWordIndex(src string) (*WordIndex, error) {
	c := WordIndex{
		path: src,
	}
	if err := c.init(); err != nil {
		return nil, err
	}
	return &c, nil
}

func (i *WordIndex) init() error {
	file, err := os.Open(i.path)
	if err != nil {
		return err
	}
	defer file.Close()

	byteOrderMarkAsString := string('\uFEFF')
	scanner := bufio.NewScanner(file)
	index := []string{}
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, ":")
		if len(parts) != 2 {
			slog.Warn(fmt.Sprintf("word frequency index, line to short: %s", line))
			continue
		}

		s := parts[0]
		if strings.HasPrefix(s, byteOrderMarkAsString) {
			s = strings.TrimPrefix(s, byteOrderMarkAsString)
		}
		index = append(index, s)
	}
	i.Words = index

	return scanner.Err()
}

func (wi *WordIndex) GetExamplesForHanzi(hanzi string, count int) []string {
	examples := []string{}
	for _, w := range wi.Words {
		if !strings.Contains(w, hanzi) {
			continue
		}
		examples = append(examples, w)
		if len(examples) == count {
			return examples
		}
	}
	return examples
}

func (wi *WordIndex) GetMostFrequent(from, to int) []string {
	known := []string{}
	mostFreq := []string{}
	for _, w := range wi.Words[from:to] {
		for _, c := range w {
			cc := strings.TrimSpace(string(c))
			if !contains(known, cc) {
				mostFreq = append(mostFreq, cc)
				known = append(known, cc)
			}
		}
		if !contains(known, strings.TrimSpace(w)) {
			mostFreq = append(mostFreq, w)
			known = append(known, w)
		}
	}
	return mostFreq
}

func contains(s []string, target string) bool {
	for _, val := range s {
		if val == target {
			return true
		}
	}
	return false
}
