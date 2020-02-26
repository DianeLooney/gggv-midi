package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/dianelooney/gggv-midi/pkg/midi"
	"github.com/hypebeast/go-osc/osc"
)

var inst = flag.String("in", "/dev/midi1", "Input instrument")
var out = flag.Int("out", 4202, "Output osc endpoint")

func main() {
	flag.Parse()

	f, err := os.Open(*inst)
	if err != nil {
		panic(err)
	}

	ch0 := make(chan midi.Message)
	go midi.Listen(f, ch0)

	cl := osc.NewClient("", 4200)
	for m := range ch0 {
		switch m.Kind {
		case midi.Volume:
			cl.Send(osc.NewMessage(
				"/source.shader/set.global/uniform1f",
				fmt.Sprintf("midiVolume%v", m.Note),
				float32(m.Velocity)/127,
			))
		case midi.NoteOff:
		case midi.NoteOn:
			cl.Send(osc.NewMessage(
				"/source.shader/set.global/uniform1f",
				fmt.Sprintf("midiHSV%v", m.Note/12),
				float32(m.Note%12)/12,
				float32(1),
				float32(m.Velocity)/127,
			))
			cl.Send(osc.NewMessage(
				"/source.shader/set.global/uniform1f",
				fmt.Sprintf("midiV%v", m.Note/12),
				float32(m.Velocity)/127,
			))
		}
	}
}
