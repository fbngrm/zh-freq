package card

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/fbngrm/zh-freq/pkg/audio"
	"github.com/fbngrm/zh-freq/pkg/cedict"
	"github.com/fbngrm/zh-freq/pkg/components"
	"github.com/fbngrm/zh-freq/pkg/heisig"
	"github.com/fbngrm/zh-mnemonics/mnemonic"
	"golang.org/x/exp/slog"
)

type Component struct {
	SimplifiedChinese string
	English           string
}

type DictData struct {
	Src         string
	English     string
	Pinyin      string
	Traditional string
}

type Card struct {
	SimplifiedChinese  string
	TraditionalChinese string
	DictData           []DictData
	Components         []Component
	Audio              string
}

type Builder struct {
	HeisigDecomp       map[string][]string
	CJKVIDecomp        map[string][]string
	HeisigDict         map[string]heisig.Entry
	CedictDict         map[string][]cedict.Entry
	ComponentsDict     map[string]components.Component
	FrequencyWordIndex []string
	AudioDownloader    audio.Downloader
	MnemonicsBuilder   *mnemonic.Builder
}

func (b *Builder) MustBuild() []*Card {
	cards := []*Card{}
	for _, word := range b.FrequencyWordIndex {
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
		DictData:           d,
		Components:         b.getWordComponents(word),
	}
}

func (b *Builder) getHanziCard(word, hanzi string) *Card {
	d, t, err := b.lookupDict(hanzi)
	if err != nil {
		slog.Error(fmt.Sprintf("ignore hanzi: %v", err))
	}

	return &Card{
		SimplifiedChinese:  hanzi,
		TraditionalChinese: t,
		DictData:           d,
		Components:         b.getHanziComponents(hanzi),
	}
}

func (b *Builder) getWordComponents(word string) []Component {
	components := []Component{}
	for _, h := range word {
		s := string(h)
		r, _, err := b.lookupDict(s)
		if err != nil {
			slog.Warn(fmt.Sprintf("get components for word %s: %v", word, err))
		}

		e := []string{}
		for _, result := range r {
			e = append(e, result.English)
		}
		if len(e) == 0 {
			slog.Warn(fmt.Sprintf("component meaning is empty in heisig: %s", s))
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
		slog.Warn(fmt.Sprintf("no components found: %s", hanzi))
	} else {
		for _, d := range decomp {
			if d == hanzi {
				continue
			}
			r, _, err := b.lookupDict(d)
			if err != nil {
				slog.Warn(fmt.Sprintf("get components for %s: %v", hanzi, err))
			}
			e := []string{}
			for _, result := range r {
				e = append(e, result.English)
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

func (b *Builder) lookupDict(word string) ([]DictData, string, error) {
	r := []DictData{}
	t := ""
	if h, ok := b.HeisigDict[word]; ok {
		r = append(r, DictData{
			Src:     "heisig",
			English: h.Meaning,
			Pinyin:  h.Pinyin,
		})
		t = h.TraditionalChinese
	}
	if h, ok := b.CedictDict[word]; ok {
		if len(h) > 1 {
			slog.Warn(fmt.Sprintf("found several cedict dict entries for word: %s", word))
		}
		for _, hh := range h {
			r = append(r, DictData{
				Src:     "cedict",
				English: strings.Join(hh.Definitions, ", "),
				Pinyin:  hh.Readings,
			})
			t = hh.Traditional
		}
	}
	if h, ok := b.ComponentsDict[word]; ok {
		r = append(r, DictData{
			Src:     "components",
			English: h.Definition,
		})
	}

	if len(r) == 0 {
		return nil, "", fmt.Errorf("no entry found for %s", word)
	}
	return r, t, nil
}
