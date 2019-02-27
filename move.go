package main

import (
	"github.com/param108/MazeAttack/models"
	"github.com/param108/MazeAttack/screen"
)

type Response struct {
	done     chan<- int
	response int
}

func hitting(objectName string, X, Y int, objects []models.Object) bool {
	for _, obj := range objects {
		if obj.C == objectName && obj.X == X && obj.Y == Y {
			return true
		}
	}

	return false
}

func num_already_dead(objects []models.Object) int {
	num_dead := 0
	for _, obj := range objects {
		if obj.C == "HERO" && obj.Dead > 0 {
			num_dead++
		}
	}

	return 4 - num_dead
}

func explode(X, Y int, objects []models.Object) []models.Object {
	ret := []models.Object{}
	for _, obj := range objects {
		if obj.C == "HERO" {
			dist := ((X - obj.X) * (X - obj.X)) + ((Y - obj.Y) * (Y - obj.Y))
			if dist < 9 {
				ret = append(ret, models.Object{C: "HERO", X: obj.X, Y: obj.Y,
					Bombs:    obj.Bombs,
					Username: obj.Username,
					Dead:     num_already_dead(objects)})
			} else {
				ret = append(ret, obj)
			}
		} else {
			ret = append(ret, obj)
		}
	}
	return ret
}

func replace_bomb(bomb models.Object, newBomb models.Object, objects []models.Object) []models.Object {
	ret := []models.Object{}
	for _, obj := range objects {
		if obj.C == "BOMB" && obj.X == bomb.X && obj.Y == bomb.Y {
			ret = append(ret, newBomb)
		} else {
			ret = append(ret, obj)
		}
	}

	return ret
}

func replace_bullet(bullet models.Object, newBullet models.Object, objects []models.Object) []models.Object {
	ret := []models.Object{}
	for _, obj := range objects {
		if obj.C == "BULLET" && obj.X == bullet.X && obj.Y == bullet.Y {
			ret = append(ret, newBullet)
		} else {
			ret = append(ret, obj)
		}
	}

	return ret
}

func moveBomb(bomb models.Object, objects []models.Object) []models.Object {
	ret := []models.Object{}
	if bomb.Expire > 0 {
		if bomb.Expire > 1 {
			return replace_bomb(bomb, models.Object{C: "BOMB", X: bomb.X, Y: bomb.Y,
				Expire: bomb.Expire - 1, Username: bomb.Username}, objects)
		}

		ret = explode(bomb.X, bomb.Y, objects)
		return replace_bomb(bomb, models.Object{C: "BOMB", X: bomb.X, Y: bomb.Y, Expire: 0,
			Username: bomb.Username}, ret)
	}
	return objects
}

func moveBullet(bullet models.Object, objects []models.Object) []models.Object {
	newX := bullet.X
	newY := bullet.Y
	switch bullet.Direction {
	case "UP":
		newY--
	case "DOWN":
		newY++
	case "RIGHT":
		newX++
	case "LEFT":
		newX--
	}

	if bullet.Dead > 0 {
		return objects
	}

	for _, obj := range objects {
		if obj.C == "HERO" && obj.X == newX && obj.Y == newY {
			objects = replace_hero(obj.Username,
				models.Object{C: "HERO", X: obj.X,
					Y: obj.Y, Bombs: obj.Bombs,
					Username: obj.Username, Dead: num_already_dead(objects)},
				objects)
		}
	}

	if newX <= 0 || newX >= 50 || newY >= 50 || newY <= 0 ||
		hitting("WALL", newX, newY, objects) ||
		hitting("BOMB", newX, newY, objects) ||
		hitting("BOMBSHED", newX, newY, objects) ||
		hitting("BULLET", newX, newY, objects) ||
		hitting("HERO", newX, newY, objects) {
		return replace_bullet(bullet, models.Object{Username: bullet.Username, C: "BULLET",
			X: bullet.X, Y: bullet.Y, Dead: 1, Direction: bullet.Direction}, objects)
	}

	return replace_bullet(bullet, models.Object{C: "BULLET", X: newX, Y: newY, Dead: 0,
		Username: bullet.Username, Direction: bullet.Direction}, objects)

}

func moveHero(done chan<- Response, heroMoved chan<- int, hero models.Object,
	objects []models.Object, dir string) models.Object {
	newX := hero.X
	newY := hero.Y
	switch dir {
	case "UP":
		newY--
	case "DOWN":
		newY++
	case "RIGHT":
		newX++
	case "LEFT":
		newX--
	}

	response := Response{heroMoved, -1}

	if hero.Dead > 0 {
		return hero
	}

	if newX <= 0 || newX >= 50 || newY >= 50 || newY <= 0 {
		response.response = -1
		done <- response
		return hero
	}

	if hitting("WALL", newX, newY, objects) {
		response.response = -1
		done <- response
		return hero
	}

	if hitting("HERO", newX, newY, objects) {
		response.response = -1
		done <- response
		return hero
	}

	if hitting("BOMB", newX, newY, objects) {
		response.response = -1
		done <- response
		return hero
	}

	// walking into a bullet is fine as bullets move after Heros.
	if hitting("BOMBSHED", newX, newY, objects) {
		response.response = 0
		done <- response
		return models.Object{C: "HERO", X: hero.X, Y: hero.Y, Bombs: 1,
			Username: hero.Username}
	}

	response.response = 0
	done <- response
	return models.Object{C: "HERO", X: newX, Y: newY, Bombs: hero.Bombs,
		Username: hero.Username}
}

func find_hero(username string, objects []models.Object) models.Object {
	for _, obj := range objects {
		if obj.C == "HERO" && obj.Username == username {
			return obj
		}
	}
	panic("Failed to find hero")
}

func can_place_bomb(X, Y int, objects []models.Object) bool {
	if X >= 50 || Y >= 50 || X <= 0 || Y <= 0 {
		return false
	}
	for _, obj := range objects {
		if obj.X == X && obj.Y == Y {
			return false
		}
	}
	return true
}

func replace_hero(username string, newHero models.Object, objects []models.Object) []models.Object {
	ret := []models.Object{}
	for _, obj := range objects {
		if obj.C == "HERO" && obj.Username == username {
			ret = append(ret, newHero)
		} else {
			ret = append(ret, obj)
		}
	}

	return ret

}

func move(moves []models.Move, objects []models.Object) []models.Object {
	responseChan := make(chan Response, 100)
	for _, m := range moves {
		hero := find_hero(m.Username, objects)
		if hero.Dead == 0 {
			if m.Move == "MOVE" {
				objects = replace_hero(m.Username,
					moveHero(responseChan, m.Done, hero, objects, m.Direction),
					objects)
			} else if m.Move == "FIRE" {
				objects = append(objects, models.Object{C: "BULLET", X: hero.X,
					Y:        hero.Y,
					Username: hero.Username, Direction: m.Direction})
				responseChan <- Response{m.Done, 0}
			} else if m.Move == "PLACE" {
				if hero.Bombs > 0 {
					if can_place_bomb(hero.X+1, hero.Y, objects) {
						objects = append(objects, models.Object{C: "BOMB",
							X: hero.X + 1, Y: hero.Y,
							Username: hero.Username,
							Expire:   5})
						objects = replace_hero(hero.Username,
							models.Object{C: "HERO", X: hero.X,
								Y: hero.Y, Bombs: 0,
								Username: hero.Username}, objects)
						responseChan <- Response{m.Done, 0}
					} else {
						responseChan <- Response{m.Done, -1}
					}
				}
			}
		}
	}

	for _, obj := range objects {
		switch obj.C {
		case "BOMB":
			objects = moveBomb(obj, objects)
		case "BULLET":
			objects = moveBullet(obj, objects)
		}
	}

	// update objects now
	GlobalObjectList = objects
	Moves = []models.Move{}
	close(responseChan)
	for response := range responseChan {
		response.done <- response.response
	}
	return objects
}

func Convert(objects []models.Object) []screen.Object {
	ret := []screen.Object{}
	for _, obj := range objects {
		switch obj.C {
		case "WALL":
			ret = append(ret, screen.NewWall(obj.X, obj.Y))
		case "HERO":
			if obj.Dead == 0 {
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
