package midi

import (
	"fmt"
	"io"
	"time"

	"github.com/hypebeast/go-osc/osc"
)

const (
	NoteOff   = 8
	NoteOn    = 9
	Volume    = 11
	PitchBend = 14
)

type Message struct {
	Time     time.Time
	Address  string
	Channel  byte
	Kind     byte
	Note     byte
	Velocity byte
}

func (m Message) String() string {
	return fmt.Sprintf(
		"%s %2v | %3v | %3v | %3v -- %v",
		m.Time.Format("15:04:05.000"),
		m.Channel,
		m.Kind,
		m.Note,
		m.Velocity,
		m.Address,
	)
}

func Parse(msg []byte) Message {
	return Message{
		Time:     time.Now(),
		Channel:  msg[0] & 15,
		Kind:     msg[0] >> 4,
		Note:     msg[1],
		Velocity: msg[2],
		Address:  fmt.Sprintf("/midi/%v", msg[0]&15),
	}
}

func Listen(f io.Reader, ch chan<- Message) (err error) {
	defer close(ch)

	for {
		msg := make([]byte, 3)
		_, err := f.Read(msg)
		if err != nil {
			return err
		}
		ch <- Parse(msg)
	}
}

type StreamFunc func(in chan Message) (out chan Message)

func Rename(from, to string) StreamFunc {
	return func(in chan Message) (out chan Message) {
		out = make(chan Message)
		go func() {
			defer close(out)

			for m := range in {
				if m.Address == from {
					m.Address = to
				}
				out <- m
			}
		}()
		return
	}
}

func Log(name string) StreamFunc {
	return func(in chan Message) (out chan Message) {
		out = make(chan Message)
		go func() {
			for m := range in {
				fmt.Println(name, m)
				out <- m
			}
			close(out)
		}()
		return out
	}
}

func Tee(in chan Message, n int) (outs []chan Message) {
	outs = make([]chan Message, n)
	for i := 0; i < n; i++ {
		outs[i] = make(chan Message)
	}
	go func() {
		for m := range in {
			for _, o := range outs {
				o <- m
			}
		}
		for _, o := range outs {
			close(o)
		}
	}()
	return
}
func Null(in chan Message) (out chan Message) {
	go func() {
		for range in {
		}
	}()
	return
}
func SendTo(target int, args ...interface{}) StreamFunc {
	cl := osc.NewClient("", target)
	return func(in chan Message) (out chan Message) {
		out = make(chan Message)
		go func() {
			for m := range in {
				var x []interface{}
				x = append(x, args...)
				x = append(x,
					float32(m.Note%12)/12,
					float32(1), //1-float32(m.Note)/127,
					float32(m.Velocity)/127,
				)
				err := cl.Send(osc.NewMessage(m.Address, x...))
				if err != nil {
					fmt.Println(err)
				}
			}
			close(out)
		}()
		return
	}
}
func IfNoteBetween(low, high byte, f StreamFunc) StreamFunc {
	return func(in chan Message) (out chan Message) {
		match := make(chan Message)
		discard := make(chan Message)
		go func() {
			for m := range in {
				if m.Note >= low && m.Note <= high {
					match <- m
					fmt.Println(m)
				} else {
					discard <- m
				}
			}
			close(match)
			close(discard)
		}()
		go func() {
			flush := f(match)
			for range flush {
			}
		}()

		return discard
	}
}

func OnlyKind(k byte) StreamFunc {
	return func(in chan Message) (out chan Message) {
		out = make(chan Message)

		go func() {
			for m := range in {
				if m.Kind == k {
					out <- m
				}
			}
			close(out)
		}()
		return out
	}
}

func Chain(in chan Message, fs ...StreamFunc) (out chan Message) {
	out = in
	for _, f := range fs {
		out = f(out)
	}
	return
}
