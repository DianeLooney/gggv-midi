package main

import (
	"flag"
	"os"

	"github.com/dianelooney/gggv-midi/pkg/midi"
)

var inst = flag.String("in", "/dev/midi2", "Input instrument")
var out = flag.Int("out", 4202, "Output osc endpoint")

func main() {
	flag.Parse()

	f, err := os.Open("/dev/midi2")
	if err != nil {
		panic(err)
	}

	ch0 := make(chan midi.Message)
	go midi.Listen(f, ch0)

	out := midi.Chain(
		ch0,
		midi.OnlyKind(midi.NoteOn),
		midi.Rename("/midi/0", "/source.shader/set/uniform1f"),
		//midi.SendTo(4200, "frag68", "u76"),
		midi.IfNoteBetween(36, 47, midi.SendTo(4200, "frag68", "u76")),
		midi.IfNoteBetween(48, 59, midi.SendTo(4200, "frag68", "u75")),
		midi.IfNoteBetween(60, 71, midi.SendTo(4200, "frag68", "u76")),
		midi.Log(""),
	)
	for range out {
	}
}
