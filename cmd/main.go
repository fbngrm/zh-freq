package main

import (
	"log"
	"path/filepath"
	"time"

	"github.com/fbngrm/zh-freq/pkg/anki"
	"github.com/fbngrm/zh-freq/pkg/card"
	"github.com/fbngrm/zh-freq/pkg/template"
	"golang.org/x/exp/slog"
)

const mnemonicsSrc = "/home/f/Dropbox/notes/chinese/mnemonics/words.csv"

func main() {
	builder, err := card.NewBuilder(mnemonicsSrc)
	if err != nil {
		log.Fatal(err)
	}
	cards := builder.MustBuild()

	deckname := "chinese::hsk1"
	modelname := "vocab"
	tmplProcessor := template.NewProcessor(
		deckname,
		filepath.Join(".", "tmpl"),
		[]string{"most frequent words"},
	)
	succs := 0
	errs := map[string]int{}
	for _, c := range cards {
		front, back, err := tmplProcessor.Fill(c)
		if err != nil {
			log.Fatalf("generate template for card %s: %v", c.SimplifiedChinese, err)
		}
		time.Sleep(10 + time.Millisecond)
		if err := anki.Export(deckname, modelname, front, back, c.MnemonicBase, c.Mnemonic); err != nil {
			if c, ok := errs[err.Error()]; ok {
				errs[err.Error()] = c + 1
				continue
			}
			errs[err.Error()] = 1
			continue
		}
		succs++
	}
	slog.Info("success", "notes added", succs)
	for e, c := range errs {
		slog.Info("failed", e, c)
	}
}
