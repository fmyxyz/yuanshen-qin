package cmd

import (
	"log"

	_ "github.com/fmyxyz/yuanshen-qin/driver"
	"gitlab.com/gomidi/midi/v2"

	"gitlab.com/gomidi/midi/v2/smf"
)

func run(filePath string) {
	err := smf.ReadTracks(filePath).
		Only(midi.NoteOnMsg, midi.NoteOffMsg).
		Play(0)
	if err != nil {
		log.Fatal(err)
	}
}
