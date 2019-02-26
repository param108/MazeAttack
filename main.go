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

var GlobalObjectList []models.Object
var CurrentUserID = 0

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
	User string
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
var mutex *sync.Mutex

func createPositionMap() {
	posmap = map[string]*Pos{}
}

func main() {
	maxx = 51
	maxy = 50
	rand.Seed(time.Now().UTC().UnixNano())
	mutex = &sync.Mutex{}

	createPositionMap()

	maze = createMaze()
	checkMaze()
	printMaze(maze)
	http.HandleFunc("/create/", createUser)
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

func checkMaze() {
	for j := 0; j < maxy; j++ {
		success := false
		for i := 0; i < maxx; i++ {
			if maze[j][i] == " " {
				success = true
				break
			}
			if !success {
				maze[j][i] = " "
			}
		}
	}
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

func getPlace() string {
	ch := rand.Int() % 2
	switch ch {
	case 0:
		return " "
	case 1:
		return "#"
	}
	return " "
}

func createMaze() [][]string {
	ret := [][]string{}
	for j := 0; j < maxy; j++ {
		row := []string{}
		for i := 0; i < maxx; i++ {
			if i%2 == 0 {
				row = append(row, " ")
				continue
			}
			row = append(row, getPlace())
		}
		ret = append(ret, row)
	}

	// place dest.
	ch := rand.Int() % maxy

	ret[ch][maxx-1] = "E"
	posmap["BADDIE"] = &Pos{maxx - 1, ch}
	return ret
}

func printMaze(maze [][]string) {
	for j := 0; j < maxy; j++ {
		for i := 0; i < maxx; i++ {
			fmt.Print(maze[j][i])
		}
		fmt.Print("\n")
	}
}
