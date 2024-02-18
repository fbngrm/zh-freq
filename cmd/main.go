package main

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/fbngrm/zh-freq/pkg/anki"
	"github.com/fbngrm/zh-freq/pkg/card"
	"github.com/fbngrm/zh-freq/pkg/template"
	"golang.org/x/exp/slog"
)

const audioDir = "./audio"
const mnemonicsSrc = "/home/f/Dropbox/notes/chinese/mnemonics/words.csv"

func main() {
	builder, err := card.NewBuilder(audioDir, mnemonicsSrc, 5)
	if err != nil {
		log.Fatal(err)
	}
	cards := builder.MustBuild()

	deckname := "vocab"
	modelname := "vocab"
	tmplProcessor := template.NewProcessor(
		deckname,
		filepath.Join(".", "tmpl"),
		[]string{"most frequent words"},
	)
	for _, c := range cards {
		front, back, err := tmplProcessor.Fill(c)
		if err != nil {
			log.Fatalf("generate template for card %s: %v", c.SimplifiedChinese, err)
		}
		if err := anki.Export(deckname, modelname, front, back, c.MnemonicBase, c.Mnemonic); err != nil {
			slog.Warn(fmt.Sprintf("export anki card %s: %v", c.SimplifiedChinese, err))
		}
	}
}
