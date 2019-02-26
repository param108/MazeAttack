package screen

import (
	"github.com/nsf/termbox-go"
)

type Screen struct {
	objects []Object
}

func NewScreen() (*Screen, error) {
	if !termbox.IsInit {
		err := termbox.Init()
		if err != nil {
			return nil, err
		}
	}
	return &Screen{[]Object{}}, nil
}

func (scr *Screen) Update(newPositions []Object) {
	if termbox.IsInit {
		for _, obj := range scr.objects {
			termbox.SetCell(obj.X, obj.Y, ' ', termbox.Attribute(0),
				termbox.Attribute(0))
		}

		scr.objects = []Object{}
		for _, obj := range newPositions {
			termbox.SetCell(obj.X, obj.Y, obj.C, termbox.Attribute(0),
				termbox.Attribute(0))
			scr.objects = append(scr.objects, NewObject(obj.X, obj.Y, obj.C))
		}

		termbox.Flush()
	}
}

func (scr *Screen) Destroy() {
	if termbox.IsInit {
		termbox.Close()
	}
}
