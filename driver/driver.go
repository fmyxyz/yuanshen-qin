package driver

import (
	"fmt"
	"time"

	"github.com/go-vgo/robotgo"
	"gitlab.com/gomidi/midi/v2/drivers"
	"gitlab.com/gomidi/midi/v2/smf"
)

/*
八度	音符编号
 	C	C#	D	D#	E	F	F#	G	G#	A	A#	B
-1	0	1	2	3	4	5	6	7	8	9	10	11
0	12	13	14	15	16	17	18	19	20	21	22	23
1	24	25	26	27	28	29	30	31	32	33	34	35
2	36	37	38	39	40	41	42	43	44	45	46	47
3	48	49	50	51	52	53	54	55	56	57	58	59
4	60	61	62	63	64	65	66	67	68	69	70	71
5	72	73	74	75	76	77	78	79	80	81	82	83
6	84	85	86	87	88	89	90	91	92	93	94	95
7	96	97	98	99	100	101	102	103	104	105	106	107
8	108	109	110	111	112	113	114	115	116	117	118	119
9	120	121	122	123	124	125	126	127
*/

var pianoKey = []string{
	"C-", "C-#", "D-", "D-#", "E-", "F-", "F-#", "G-", "G-#", "A-", "A-#", "B-",
	"C0", "C0#", "D0", "D0#", "E0", "F0", "F0#", "G0", "G0#", "A0", "A0#", "B0",
	"C1", "C1#", "D1", "D1#", "E1", "F1", "F1#", "G1", "G1#", "A1", "A1#", "B1",
	"C2", "C2#", "D2", "D2#", "E2", "F2", "F2#", "G2", "G2#", "A2", "A2#", "B2",
	"C3", "C3#", "D3", "D3#", "E3", "F3", "F3#", "G3", "G3#", "A3", "A3#", "B3",
	"C4", "C4#", "D4", "D4#", "E4", "F4", "F4#", "G4", "G4#", "A4", "A4#", "B4",
	"C5", "C5#", "D5", "D5#", "E5", "F5", "F5#", "G5", "G5#", "A5", "A5#", "B5",
	"C6", "C6#", "D6", "D6#", "E6", "F6", "F6#", "G6", "G6#", "A6", "A6#", "B6",
	"C7", "C7#", "D7", "D7#", "E7", "F7", "F7#", "G7", "G7#", "A7", "A7#", "B7",
	"C8", "C8#", "D8", "D8#", "E8", "F8", "F8#", "G8", "G8#", "A8", "A8#", "B8",
	"C9", "C9#", "D9", "D9#", "E9", "F9", "F9#", "G9",
}

var keys = map[string]string{
	"C3": "z", "C3#": "zx", "D3": "x", "D3#": "xc", "E3": "c", "F3": "v", "F3#": "vb", "G3": "b", "G3#": "bn", "A3": "n", "A3#": "nm", "B3": "m",
	"C4": "a", "C4#": "as", "D4": "s", "D4#": "ss", "E4": "d", "F4": "f", "F4#": "fg", "G4": "g", "G4#": "gh", "A4": "h", "A4#": "hj", "B4": "j",
	"C5": "q", "C5#": "qw", "D5": "w", "D5#": "ww", "E5": "e", "F5": "r", "F5#": "rt", "G5": "t", "G5#": "ty", "A5": "y", "A5#": "yu", "B5": "u",
}

func init() {
	drv := New("键盘自动按键Drv")
	drivers.Register(drv)
}

type Driver struct {
	in            *in
	out           *out
	name          string
	last          time.Time
	stopListening bool
	rd            *drivers.Reader
}

func New(name string) drivers.Driver {
	d := &Driver{name: name}
	d.in = &in{name: name + "-in", driver: d, number: 0}
	d.out = &out{name: name + "-out", driver: d, number: 0}
	d.last = time.Now()
	return d
}

func (f *Driver) String() string               { return f.name }
func (f *Driver) Close() error                 { return nil }
func (f *Driver) Ins() ([]drivers.In, error)   { return []drivers.In{f.in}, nil }
func (f *Driver) Outs() ([]drivers.Out, error) { return []drivers.Out{f.out}, nil }

type in struct {
	number int
	name   string
	isOpen bool
	driver *Driver
}

func (f *in) String() string          { return f.name }
func (f *in) Number() int             { return f.number }
func (f *in) IsOpen() bool            { return f.isOpen }
func (f *in) Underlying() interface{} { return nil }

func (f *in) Listen(onMsg func([]byte, int32), conf drivers.ListenConfig) (func(), error) {
	f.driver.last = time.Now()

	stopper := func() {
		f.driver.stopListening = true
	}

	f.driver.rd = drivers.NewReader(conf, onMsg)
	return stopper, nil
}

func (f *in) Close() error {
	if !f.isOpen {
		return nil
	}
	f.isOpen = false
	return nil
}

func (f *in) Open() error {
	if f.isOpen {
		return nil
	}
	f.isOpen = true
	return nil
}

type out struct {
	number int
	name   string
	isOpen bool
	driver *Driver
}

func (f *out) Number() int             { return f.number }
func (f *out) IsOpen() bool            { return f.isOpen }
func (f *out) String() string          { return f.name }
func (f *out) Underlying() interface{} { return nil }

func (f *out) Close() error {
	if !f.isOpen {
		return nil
	}
	f.isOpen = false
	return nil
}

func (f *out) Send(bt []byte) error {
	if !f.isOpen {
		return drivers.ErrPortClosed
	}

	if f.driver.stopListening {
		return nil
	}

	m := smf.Message(bt)
	var channel, key, velocity uint8

	if m.GetNoteOn(&channel, &key, &velocity) {
		fmt.Printf("channel:%v, key:%v, velocity:%v\n", channel, key, velocity)
		if velocity == 0 {
			return nil
		}
		name := pianoKey[key]
		s := keys[name]
		for _, v := range []rune(s) {
			go robotgo.KeyPress(string(v))
		}
	}

	return nil
}

func (f *out) Open() error {
	if f.isOpen {
		return nil
	}
	f.isOpen = true
	return nil
}
