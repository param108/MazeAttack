package main

import (
	"encoding/json"
	"fmt"
	"github.com/param108/MazeAttack/models"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
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
	Direction    string
	User         string
	UserSecret   string
	ServerSecret string
}

type MoveResponse struct {
	Objects []models.Object
}

type CreateMessage struct {
	User         string
	ServerSecret string
}

type CreateResponse struct {
	UserSecret string
}

var posmap map[string]*Pos
var maze [][]string

func createPositionMap() {

	posmap = map[string]*Pos{}
}

func loginMonitor(in []chan int, out chan<- int) {
	for i := 0; i < len(in); i++ {
		<-in[i]
	}

	for i := 0; i < len(in); i++ {
		out <- i
	}
}

var loginMonitorInputs = []chan int{}
var loginMonitorOutput = make(chan int)

func setupLoginMonitor() {
	for i := 0; i < 4; i++ {
		loginMonitorInputs = append(loginMonitorInputs, make(chan int))
	}

	go loginMonitor(loginMonitorInputs, loginMonitorOutput)
}
func main() {

	if len(os.Args) != 3 {
		fmt.Printf("%v <secret key> <port>", os.Args[0])
		return
	}

	ServerSecret = os.Args[1]

	setupLoginMonitor()

	rand.Seed(time.Now().UTC().UnixNano())
	mutex = &sync.Mutex{}

	createMaze()
	http.HandleFunc("/login/", createUser)
	http.HandleFunc("/move/", moveUser)
	http.ListenAndServe(":"+os.Args[2], nil)
}

func ServerAuth(auth string) bool {
	return auth == ServerSecret
}

func randomString(length int) string {
	digits := "ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890_"
	ret := ""
	for i := 0; i < length; i++ {
		ret = ret + string(digits[rand.Int()%len(digits)])
	}

	return ret
}

func createUser(w http.ResponseWriter, r *http.Request) {

	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	mutex.Lock()
	msg := CreateMessage{}
	// Unmarshal
	err = json.Unmarshal(b, &msg)
	if err != nil {
		http.Error(w, err.Error(), 500)
		mutex.Unlock()
		return
	}

	if !ServerAuth(msg.ServerSecret) {
		http.Error(w, "Forbidden", 403)
		mutex.Unlock()
		return
	}

	if CurrentUserID >= 4 {
		http.Error(w, "everyone logged in already", 403)
		mutex.Unlock()
		return
	}

	if _, ok := UserSecrets[msg.User]; ok {
		http.Error(w, "already logged in", 409)
		mutex.Unlock()
		return
	}

	UserSecrets[msg.User] = randomString(10)
	myID := CurrentUserID
	CurrentUserID++
	mutex.Unlock()

	// wait for 4 to login
	loginMonitorInputs[myID] <- myID
	<-loginMonitorOutput

	success := CreateResponse{}
	success.UserSecret = UserSecrets[msg.User]
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

	// Unmarshal
	msg := MoveMessage{}
	err = json.Unmarshal(b, &msg)
	if err != nil {
		http.Error(w, err.Error(), 500)
		mutex.Unlock()
		return
	}

	if !ServerAuth(msg.ServerSecret) {
		http.Error(w, "Forbidden", 403)
		mutex.Unlock()
		return
	}

	if userSecret, ok := UserSecrets[msg.User]; ok {
		if userSecret != msg.UserSecret {
			http.Error(w, "Forbidden", 403)
			mutex.Unlock()
			return
		}
	} else {
		http.Error(w, "Forbidden", 403)
		mutex.Unlock()
		return
	}

	for _, mv := range Moves {
		if mv.Username == msg.User {
			// already moved this turn
			http.Error(w, "Forbidden", 403)
			mutex.Unlock()
			return
		}
	}

	done := make(chan int)
	Moves = append(Moves, models.Move{msg.User, "MOVE", msg.Direction, done})

	mutex.Unlock()

	rc := <-done
	if rc < 0 {
		http.Error(w, err.Error(), 409)
		return
	}

	success := MoveResponse{}
	success.Objects = GlobalObjectList

	fmt.Println("moved user", msg.User)

	output, err := json.Marshal(success)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("content-type", "application/json")
	w.Write(output)

}

func fire(w http.ResponseWriter, r *http.Request) {
	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	mutex.Lock()

	// Unmarshal
	msg := MoveMessage{}
	err = json.Unmarshal(b, &msg)
	if err != nil {
		http.Error(w, err.Error(), 500)
		mutex.Unlock()
		return
	}

	if !ServerAuth(msg.ServerSecret) {
		http.Error(w, "Forbidden", 403)
		mutex.Unlock()
		return
	}

	if userSecret, ok := UserSecrets[msg.User]; ok {
		if userSecret != msg.UserSecret {
			http.Error(w, "Forbidden", 403)
			mutex.Unlock()
			return
		}
	} else {
		http.Error(w, "Forbidden", 403)
		mutex.Unlock()
		return
	}

	for _, mv := range Moves {
		if mv.Username == msg.User {
			// already moved this turn
			http.Error(w, "Forbidden", 403)
			mutex.Unlock()
			return
		}
	}

	done := make(chan int)
	Moves = append(Moves, models.Move{msg.User, "MOVE", msg.Direction, done})

	mutex.Unlock()

	rc := <-done
	if rc < 0 {
		http.Error(w, err.Error(), 500)
		return
	}
	success := MoveResponse{}
	success.Objects = GlobalObjectList

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
