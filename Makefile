.PHONY: audio
src=./audio
dst=/home/f/.local/share/Anki2/User\ 1/collection.media
audio:
	mkdir -p $(dst)
	cp $(src)/*.mp3 $(dst)/
