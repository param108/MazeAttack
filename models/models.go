package models

import (
	"github.com/param108/MazeAttack/screen"
)

type Move struct {
	Username  string
	Move      string   // can be MOVE or FIRE
	Direction string   // which direction
	Done      chan int // Who is waiting ?
}

type Object struct {
	C         string
	X         int
	Y         int
	Expire    int    // number of ticks for the bomb to expire
	Direction string // direction of bullet
	Username  string
	Dead      int //If HERO is dead then this will store the place.
	Bombs     int // number of bombs this HERO has
}

func Convert(objects []Object) []screen.Object {
	ret := []screen.Object{}
	for _, obj := range objects {
		switch obj.C {
		case "WALL":
			ret = append(ret, screen.NewWall(obj.X, obj.Y))
		case "HERO":
			if obj.Dead != 0 {
				ret = append(ret, screen.NewHero(obj.X, obj.Y, obj.Username))
			}
		case "BOMBSHED":
			ret = append(ret, screen.NewBombShed(obj.X, obj.Y))
		case "BOMB":
			if obj.Expire > 0 {
				ret = append(ret, screen.NewBomb(obj.X, obj.Y))
			}
		case "BULLET":
			if obj.Dead == 0 {
				ret = append(ret, screen.NewBullet(obj.X, obj.Y))
			}
		}
	}
	return ret
}
