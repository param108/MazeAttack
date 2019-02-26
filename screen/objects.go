package screen

import ()

type Object struct {
	C rune
	X int
	Y int
}

func NewObject(X, Y int, C rune) Object {
	return Object{C, X, Y}
}

func NewHero(X, Y int, username string) Object {
	return Object{([]rune(username))[0], X, Y}
}

func NewWall(X, Y int) Object {
	return Object{rune('#'), X, Y}
}

func NewBombShed(X, Y int) Object {
	return Object{rune('@'), X, Y}
}

func NewBomb(X, Y int) Object {
	return Object{rune('='), X, Y}
}

func NewBullet(X, Y int) Object {
	return Object{rune('*'), X, Y}
}
