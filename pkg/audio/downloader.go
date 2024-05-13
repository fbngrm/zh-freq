package audio

import (
	"context"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	texttospeech "cloud.google.com/go/texttospeech/apiv1"
	"cloud.google.com/go/texttospeech/apiv1/texttospeechpb"
)

type Downloader struct {
	AudioDir string
}

// we support 4 different voices only
var voices = []*texttospeechpb.VoiceSelectionParams{
	{
		LanguageCode: "cmn-CN",
		Name:         "cmn-CN-Wavenet-C",
		SsmlGender:   texttospeechpb.SsmlVoiceGender_MALE,
	},
	{
		LanguageCode: "cmn-CN",
		Name:         "cmn-CN-Wavenet-A",
		SsmlGender:   texttospeechpb.SsmlVoiceGender_FEMALE,
	},
	{
		LanguageCode: "cmn-CN",
		Name:         "cmn-TW-Wavenet-C",
		SsmlGender:   texttospeechpb.SsmlVoiceGender_MALE,
	},
	{
		LanguageCode: "cmn-CN",
		Name:         "cmn-TW-Wavenet-A",
		SsmlGender:   texttospeechpb.SsmlVoiceGender_FEMALE,
	},
}

func (p *Downloader) GetVoices(speakers map[string]struct{},
) map[string]*texttospeechpb.VoiceSelectionParams {
	v := make(map[string]*texttospeechpb.VoiceSelectionParams)
	var i int
	for speaker := range speakers {
		v[speaker] = voices[i]
		i++
	}
	return v
}

// download audio file from google text-to-speech api if it doesn't exist in cache dir.
func (p *Downloader) Fetch(ctx context.Context, query, filename string) (string, error) {
	if err := os.MkdirAll(p.AudioDir, os.ModePerm); err != nil {
		return "", err
	}
	filename = "most_freq_" + filename + ".mp3"
	lessonPath := filepath.Join(p.AudioDir, filename)
	if _, err := os.Stat(lessonPath); err == nil {
		return filename, nil
	}

	resp, err := fetch(ctx, query, nil)
	if err != nil {
		return "", err
	}

	// The resp's AudioContent is binary.
	err = ioutil.WriteFile(lessonPath, resp.AudioContent, os.ModePerm)
	if err != nil {
		return "", err
	}

	return filename, nil
}

// uses a random voice if param voice is nil
func fetch(ctx context.Context, query string, voice *texttospeechpb.VoiceSelectionParams) (*texttospeechpb.SynthesizeSpeechResponse, error) {
	time.Sleep(100 * time.Millisecond)
	client, err := texttospeech.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	if voice == nil {
		rand.Seed(time.Now().UnixNano()) // initialize global pseudo random generator
		voice = voices[rand.Intn(len(voices))]
	}
	// perform the text-to-speech request on the text input with the selected
	// voice parameters and audio file type.
	req := texttospeechpb.SynthesizeSpeechRequest{
		// set the text input to be synthesized.
		Input: &texttospeechpb.SynthesisInput{
			InputSource: &texttospeechpb.SynthesisInput_Text{Text: query},
		},
		// build the voice request, select the language code ("en-US") and the SSML
		// voice gender ("neutral").
		Voice: voice,
		// select the type of audio file you want returned.
		AudioConfig: &texttospeechpb.AudioConfig{
			AudioEncoding: texttospeechpb.AudioEncoding_MP3,
			SpeakingRate:  0.9,
		},
	}
	return client.SynthesizeSpeech(ctx, &req)
}
