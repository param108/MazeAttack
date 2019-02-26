package main

import (
	"encoding/json"
	"fmt"
	"github.com/param108/MazeAttack/models"
	"io/ioutil"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

var CurrentUserID = 0
var mutex *sync.Mutex
var ServerSecret string
var UserSecrets = make(map[string]string)
var Moves = []models.Move{}
var GlobalObjectList []models.Object

type Profile struct {
	Name    string
	Hobbies []string
}

var maxx int
var maxy int

type Pos struct {
	X int
	Y int
}

type MoveMessage struct {
	Move string
	User string
}

type MoveResponse struct {
	Error   string
	Message string
	X       int
	Y       int
	BaddieX int
	BaddieY int
}

type CreateMessage struct {
	User         string
	ServerSecret string
}

type CreateResponse struct {
	Error   string
	Message string
	X       int
	Y       int
	BaddieX int
	BaddieY int
}

var posmap map[string]*Pos
var maze [][]string

func createPositionMap() {
	posmap = map[string]*Pos{}
}

func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	mutex = &sync.Mutex{}

	createMaze()
	http.HandleFunc("/login/", createUser)
	http.HandleFunc("/move/", moveUser)
	http.ListenAndServe(":3000", nil)
}

func createUser(w http.ResponseWriter, r *http.Request) {

	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	mutex.Lock()
	defer mutex.Unlock()
	msg := CreateMessage{}
	// Unmarshal
	err = json.Unmarshal(b, &msg)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	if _, ok := posmap[msg.User]; ok {
		response := CreateResponse{}
		response.Message = "Already Exists"
		response.Error = "true"
		output, _ := json.Marshal(response)
		http.Error(w, string(output), 409)
		return
	}
	posmap[msg.User] = &Pos{0, rand.Int() % maxy}
	success := CreateResponse{}
	success.Error = "false"
	success.Message = "Success"
	success.X = posmap[msg.User].X
	success.Y = posmap[msg.User].Y
	success.BaddieX = posmap["BADDIE"].X
	success.BaddieY = posmap["BADDIE"].Y
	fmt.Println("created user", msg.User)
	output, err := json.Marshal(success)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("content-type", "application/json")
	w.Write(output)

}

/*func move(move string, user string) bool {
	if userPos, ok := posmap[user]; ok {
		switch move {
		case "UP":
			if userPos.Y == 0 {
				return true
			}

			if strings.Compare(maze[userPos.Y-1][userPos.X], "#") == 0 {
				//Dont move just return
				return true
			}
			posmap[user].Y = posmap[user].Y - 1
		case "DOWN":
			if userPos.Y == maxy-1 {
				return true
			}
			if strings.Compare(maze[userPos.Y+1][userPos.X], "#") == 0 {
				//Dont move just return
				return true
			}
			posmap[user].Y = posmap[user].Y + 1
		case "RIGHT":
			if userPos.X == maxx-1 {
				return true
			}
			if strings.Compare(maze[userPos.Y][userPos.X+1], "#") == 0 {
				//Dont move just return
				return true
			}
			posmap[user].X = posmap[user].X + 1
		case "LEFT":
			if userPos.X == 0 {
				return true
			}
			if strings.Compare(maze[userPos.Y][userPos.X-1], "#") == 0 {
				//Dont move just return
				return true
			}
			posmap[user].X = posmap[user].X - 1
		default:
			return false
		}

		if posmap[user].X == posmap["BADDIE"].X && posmap[user].Y == posmap["BADDIE"].Y {
			fmt.Println(user, " has won")
		}
		return true
	}
	return false
}*/

func moveUser(w http.ResponseWriter, r *http.Request) {
	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	mutex.Lock()
	defer mutex.Unlock()

	// Unmarshal
	msg := MoveMessage{}
	err = json.Unmarshal(b, &msg)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	/*if !move(msg.Move, msg.User) {
		http.Error(w, "Invalid Move", 500)
		return
	}*/
	success := MoveResponse{}
	success.Error = "false"
	success.Message = "Success"
	success.X = posmap[msg.User].X
	success.Y = posmap[msg.User].Y
	success.BaddieX = posmap["BADDIE"].X
	success.BaddieY = posmap["BADDIE"].Y
	fmt.Println("moved user", msg.User)

	output, err := json.Marshal(success)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("content-type", "application/json")
	w.Write(output)

}

func foo(w http.ResponseWriter, r *http.Request) {
	profile := Profile{"Alex", []string{"snowboarding", "programming"}}

	js, err := json.Marshal(profile)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func is_spot_empty(X, Y int) bool {
	for _, obj := range GlobalObjectList {
		if X == obj.X && Y == obj.Y {
			return false
		}
	}
	return true
}

func createMaze() {
	for i := 0; i < 15; i++ {
		newWall := models.Object{C: "WALL", X: (rand.Int() % 50) + 1, Y: (rand.Int() % 50) + 1}
		for !is_spot_empty(newWall.X, newWall.Y) {
			newWall = models.Object{C: "WALL", X: (rand.Int() % 50) + 1, Y: (rand.Int() % 50) + 1}
		}
		GlobalObjectList = append(GlobalObjectList, newWall)
	}

	newBombShed := models.Object{C: "BOMBSHED", X: (rand.Int() % 50) + 1, Y: (rand.Int() % 50) + 1}
	for !is_spot_empty(newBombShed.X, newBombShed.Y) {
		newBombShed = models.Object{C: "BOMBSHED", X: (rand.Int() % 50) + 1, Y: (rand.Int() % 50) + 1}
	}
	GlobalObjectList = append(GlobalObjectList, newBombShed)

}
