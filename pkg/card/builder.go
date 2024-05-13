package card

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/fbngrm/zh-freq/pkg/audio"
	"github.com/fbngrm/zh-freq/pkg/cedict"
	"github.com/fbngrm/zh-freq/pkg/cjkvi"
	"github.com/fbngrm/zh-freq/pkg/components"
	"github.com/fbngrm/zh-freq/pkg/heisig"
	"github.com/fbngrm/zh-freq/pkg/hsk"
	"github.com/fbngrm/zh-mnemonics/mnemonic"
	"golang.org/x/exp/slog"
)

const idsSrc = "./pkg/heisig/heisig_decomp.json"
const dictSrc = "./pkg/heisig/traditional.txt"
const loachSrc = "./pkg/loach/loach_word_order.json"
const cjkviSrc = "./pkg/cjkvi/ids.txt"
const cedictSrc = "./pkg/cedict/cedict_1_0_ts_utf-8_mdbg.txt"
const frequencySrc = "./pkg/frequency/global_wordfreq.release_UTF-8.txt"

type Component struct {
	SimplifiedChinese string
	English           string
}

type DictEntry struct {
	Src          string
	English      string
	Pinyin       string
	Traditional  string
	MnemonicBase string
}

type Card struct {
	SimplifiedChinese  string
	TraditionalChinese string
	DictEntries        map[string]map[string]DictEntry // map[dict_name]map[pinyin]DictEntry
	Components         []Component
	Audio              string
	MnemonicBase       string
	Mnemonic           string
}

type Builder struct {
	HeisigDecomp     map[string][]string
	CJKVIDecomp      map[string][]string
	HeisigDict       map[string]heisig.Entry
	CedictDict       map[string][]cedict.Entry
	ComponentsDict   map[string]components.Component
	WordIndex        []string
	AudioDownloader  audio.Downloader
	MnemonicsBuilder *mnemonic.Builder
	HSKDict          map[string]hsk.Entry
}

func NewBuilder(audioDir, mnemonicsSrc, hskSrc string, numWords int) (*Builder, error) {
	heisigDecomp, err := heisig.NewDecompositionIndex(idsSrc)
	if err != nil {
		return nil, err
	}
	cjkviDecomp, err := cjkvi.NewDecompositionIndex(cjkviSrc)
	if err != nil {
		return nil, err
	}
	heisigDict, err := heisig.NewDict(dictSrc)
	if err != nil {
		return nil, err
	}
	cedictDict, err := cedict.NewDict(cedictSrc)
	if err != nil {
		return nil, err
	}
	componentsDict := components.NewDict()
	// index, err := index.NewMostFrequent(frequencySrc)
	// if err != nil {
	// 	return nil, err
	// }
	mnBuilder, err := mnemonic.NewBuilder(mnemonicsSrc)
	if err != nil {
		return nil, err
	}
	hskDict, err := hsk.NewDict(hskSrc)
	if err != nil {
		return nil, err
	}

	return &Builder{
		HeisigDecomp:   heisigDecomp,
		CJKVIDecomp:    cjkviDecomp,
		HeisigDict:     heisigDict,
		CedictDict:     cedictDict,
		ComponentsDict: componentsDict,
		// WordIndex:      index.GetMostFrequent(0, numWords),
		WordIndex: hsk.GetByLevel(hskDict, 1),
		AudioDownloader: audio.Downloader{
			AudioDir: audioDir,
		},
		MnemonicsBuilder: mnBuilder,
		HSKDict:          hskDict,
	}, nil
}

func (b *Builder) MustBuild() []*Card {
	cards := []*Card{}
	for _, word := range b.WordIndex {
		for _, hanzi := range word {
			// if not hanzi is already known
			cards = append(cards, b.getHanziCard(word, string(hanzi)))
		}
		if utf8.RuneCountInString(word) > 1 {
			cards = append(cards, b.getWordCard(word))
		}
	}
	return b.GetAudio(cards)
}

func (b *Builder) GetAudio(cards []*Card) []*Card {
	for _, c := range cards {
		filename, err := b.AudioDownloader.Fetch(
			context.Background(),
			c.SimplifiedChinese,
			c.SimplifiedChinese,
		)
		if err != nil {
			slog.Warn(fmt.Sprintf("download audio for %s: %v", c.SimplifiedChinese, err))
		}
		c.Audio = filename
	}
	return cards
}

func (b *Builder) getWordCard(word string) *Card {
	d, t, err := b.lookupDict(word)
	if err != nil {
		slog.Error(fmt.Sprintf("ignore word: %v", err))
	}

	return &Card{
		SimplifiedChinese:  word,
		TraditionalChinese: t,
		DictEntries:        d,
		Components:         b.getWordComponents(word),
	}
}

func (b *Builder) getHanziCard(word, hanzi string) *Card {
	entries, t, err := b.lookupDict(hanzi)
	if err != nil {
		slog.Error(fmt.Sprintf("ignore hanzi: %v", err))
	}

	mnemonicBase := ""
	for _, entry := range entries {
		for _, result := range entry {
			mnemonicBase = fmt.Sprintf("%s%s - %s<br>%s<br>", mnemonicBase, result.Src, result.Pinyin, result.MnemonicBase)
		}
	}

	return &Card{
		SimplifiedChinese:  hanzi,
		TraditionalChinese: t,
		DictEntries:        entries,
		Components:         b.getHanziComponents(hanzi),
		MnemonicBase:       mnemonicBase,
		Mnemonic:           b.MnemonicsBuilder.Lookup(hanzi),
	}
}

func (b *Builder) getWordComponents(word string) []Component {
	components := []Component{}
	for _, h := range word {
		s := string(h)
		entries, _, err := b.lookupDict(s)
		if err != nil {
			slog.Warn(fmt.Sprintf("get components for %s: %v", word, err))
		}
		e := []string{}
		for _, entry := range entries {
			for _, result := range entry {
				e = append(e, result.English)
			}
		}
		if len(e) == 0 {
			slog.Warn(fmt.Sprintf("component meaning is empty: %s", s))
		}
		components = append(components, Component{
			SimplifiedChinese: s,
			English:           strings.Join(e, ", "),
		})
	}
	return components
}

func (b *Builder) getHanziComponents(hanzi string) []Component {
	decomp := b.HeisigDecomp[hanzi]
	if len(decomp) == 0 {
		if d, ok := b.CJKVIDecomp[hanzi]; ok {
			decomp = d
		}
	}
	components := []Component{}
	if len(decomp) == 0 {
		// FIXME: try cjkvi decomp here
		slog.Warn(fmt.Sprintf("no components found: %s", hanzi))
	} else {
		for _, d := range decomp {
			if d == hanzi {
				continue
			}
			entries, _, err := b.lookupDict(d)
			if err != nil {
				slog.Warn(fmt.Sprintf("get components for %s: %v", hanzi, err))
			}
			e := []string{}
			for _, entry := range entries {
				for _, result := range entry {
					e = append(e, result.English)
				}
			}
			if len(e) == 0 {
				slog.Warn(fmt.Sprintf("component meaning is empty in heisig: %s", d))
			}
			components = append(components, Component{
				SimplifiedChinese: d,
				English:           strings.Join(e, ", "),
			})
		}
	}
	return components
}

func (b *Builder) lookupDict(word string) (map[string]map[string]DictEntry, string, error) {
	// map[dict_name]map[pinyin]DictEntry
	entries := map[string]map[string]DictEntry{}
	t := ""

	// lookup in HSK dict
	if h, ok := b.HSKDict[word]; ok {
		var m string
		var err error
		if utf8.RuneCountInString(word) == 1 {
			m, err = b.MnemonicsBuilder.GetBase(h.Pinyin)
			if err != nil {
				slog.Warn(fmt.Sprintf("hsk: get mnemonic base for: %s", h.Pinyin))
			}
		}
		r := map[string]DictEntry{}
		r[h.Pinyin] = DictEntry{
			Src:          "hsk",
			English:      h.Meaning,
			Pinyin:       h.Pinyin,
			MnemonicBase: m,
		}
		entries["hsk"] = r
	}

	// lookup in heisig dict
	if h, ok := b.HeisigDict[word]; ok {
		var m string
		var err error
		if utf8.RuneCountInString(word) == 1 {
			m, err = b.MnemonicsBuilder.GetBase(h.Pinyin)
			if err != nil {
				slog.Warn(fmt.Sprintf("heisig: get mnemonic base for: %s", h.Pinyin))
			}
		}
		r := map[string]DictEntry{}
		r[h.Pinyin] = DictEntry{
			Src:          "heisig",
			English:      h.Meaning,
			Pinyin:       h.Pinyin,
			MnemonicBase: m,
		}
		entries["heisig"] = r
		t = h.TraditionalChinese
	}

	// lookup in cedict
	if h, ok := b.CedictDict[word]; ok {
		r := map[string]DictEntry{}
		for _, hh := range h {
			if e, ok := r[hh.Readings]; ok {
				r[hh.Readings] = DictEntry{
					Src:          e.Src,
					English:      e.English + ", " + strings.Join(hh.Definitions, ", "),
					Pinyin:       e.Pinyin,
					MnemonicBase: e.MnemonicBase,
				}
				continue
			}
			var m string
			var err error
			if utf8.RuneCountInString(word) == 1 {
				m, err = b.MnemonicsBuilder.GetBase(hh.Readings)
				if err != nil {
					slog.Warn(fmt.Sprintf("cedict: get mnemonic base for: %s", hh.Readings))
				}
			}
			r[hh.Readings] = DictEntry{
				Src:          "cedict",
				English:      strings.Join(hh.Definitions, ", "),
				Pinyin:       hh.Readings,
				MnemonicBase: m,
			}
			t = hh.Traditional
		}
		entries["cedict"] = r
	}

	// lookup in components dict
	if h, ok := b.ComponentsDict[word]; ok {
		r := map[string]DictEntry{}
		r[""] = DictEntry{
			Src:     "components",
			English: h.Definition,
		}
		entries["components"] = r
	}

	if len(entries) == 0 {
		return nil, "", fmt.Errorf("lookup word: %s", word)
	}
	return entries, t, nil
}
