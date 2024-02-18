package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/fbngrm/zh-freq/pkg/anki"
	"github.com/fbngrm/zh-freq/pkg/audio"
	"github.com/fbngrm/zh-freq/pkg/card"
	"github.com/fbngrm/zh-freq/pkg/cedict"
	"github.com/fbngrm/zh-freq/pkg/cjkvi"
	"github.com/fbngrm/zh-freq/pkg/components"
	"github.com/fbngrm/zh-freq/pkg/frequency"
	"github.com/fbngrm/zh-freq/pkg/heisig"
	"github.com/fbngrm/zh-freq/pkg/template"
	"github.com/fbngrm/zh-mnemonics/mnemonic"
	"github.com/fbngrm/zh-mnemonics/pinyin"
	"golang.org/x/exp/slog"
)

const idsSrc = "./pkg/heisig/heisig_decomp.json"
const dictSrc = "./pkg/heisig/traditional.txt"
const loachSrc = "./pkg/loach/loach_word_order.json"
const cjkviSrc = "./pkg/cjkvi/ids.txt"
const cedictSrc = "./pkg/cedict/cedict_1_0_ts_utf-8_mdbg.txt"
const audioDir = "./audio"
const frequencySrc = "./pkg/frequency/global_wordfreq.release_UTF-8.txt"

func main() {
	heisigDecomp, err := heisig.NewDecompositionIndex(idsSrc)
	if err != nil {
		log.Fatal(err)
	}
	cjkviDecomp, err := cjkvi.NewDecompositionIndex(cjkviSrc)
	if err != nil {
		log.Fatal(err)
	}

	heisigDict, err := heisig.NewDict(dictSrc)
	if err != nil {
		log.Fatal(err)
	}
	cedictDict, err := cedict.NewDict(cedictSrc)
	if err != nil {
		fmt.Printf("could not init cedict: %v\n", err)
		os.Exit(1)
	}
	componentsDict := components.NewDict()

	index, err := frequency.NewWordIndex(frequencySrc)
	if err != nil {
		log.Fatal(err)
	}

	// loachIndex, err := loach.NewFrequencyIndex(loachSrc)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	words := index.GetMostFrequent(0, 5)

	builder := card.Builder{
		HeisigDecomp:       heisigDecomp,
		CJKVIDecomp:        cjkviDecomp,
		HeisigDict:         heisigDict,
		CedictDict:         cedictDict,
		ComponentsDict:     componentsDict,
		FrequencyWordIndex: words,
		AudioDownloader: audio.Downloader{
			AudioDir: audioDir,
		},
		MnemonicsBuilder: mnemonic.NewBuilder(pinyin.NewTable()),
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
		if err := anki.Export(deckname, modelname, front, back, c.MnemonicBase, ""); err != nil {
			slog.Warn(fmt.Sprintf("export anki card %s: %v", c.SimplifiedChinese, err))
		}
	}
}
