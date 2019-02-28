package models

import ()

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

func CopyObject(a Object) Object {
	b := Object{}
	b.C = a.C
	b.X = a.X
	b.Y = a.Y
	b.Expire = a.Expire
	b.Direction = a.Direction
	b.Username = a.Username
	b.Dead = a.Dead
	b.Bombs = a.Bombs
	return b
}
