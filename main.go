package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/param108/MazeAttack/models"
	"github.com/param108/MazeAttack/screen"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"
)

var CurrentUserID = 0
var mutex *sync.Mutex
var ServerSecret string
var UserSecrets = make(map[string]string)
var Moves = []models.Move{}
var GlobalObjectList []models.Object
var ServerScreen *screen.Screen
var ServerWaitingForPlayers = true

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
	ServerWaitingForPlayers = false
	for i := 0; i < len(in); i++ {
		out <- i
	}

}

var loginMonitorInputs = []chan int{}
var loginMonitorOutput = make(chan int)
var logfile = "logfile.txt"

func log(s string) {
	f, err := os.OpenFile(logfile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err == nil {
		f.Write([]byte(s))
		f.Close()
	} else {
		Server.Shutdown(context.TODO())
		fmt.Println("HELP", err.Error())
		panic(err.Error())
	}

}

func setupLoginMonitor() {
	for i := 0; i < 4; i++ {
		loginMonitorInputs = append(loginMonitorInputs, make(chan int))
	}

	go loginMonitor(loginMonitorInputs, loginMonitorOutput)
}

func everyonesDead() bool {
	count := 0
	for _, obj := range GlobalObjectList {
		if obj.C == "HERO" && obj.Dead == 0 {
			count++
		}
	}
	return count <= 1
}

func UIFramesTick(delay int) {
	ticker := time.NewTicker(time.Duration(delay) * time.Second)

	for _ = range ticker.C {
		log("Tick\n")
		if ServerWaitingForPlayers {
			fmt.Println("Waiting For Players")
			fmt.Println("Number Joined: ", len(UserSecrets))
			for k := range UserSecrets {
				fmt.Println("Joined User: ", k)
			}
			continue
		}

		if ServerScreen == nil {
			ServerScreen, _ = screen.NewScreen()
		}

		mutex.Lock()
		move(Moves, GlobalObjectList)
		screenObjectList := Convert(GlobalObjectList)
		log(fmt.Sprint(GlobalObjectList))
		ServerScreen.Update(screenObjectList)
		if everyonesDead() {
			if ServerScreen != nil {
				ServerScreen.Destroy()
			}
			Server.Shutdown(context.TODO())
			break
		}
		mutex.Unlock()
	}

}

var Server *http.Server

func waitForInterrupt(c chan os.Signal) {
	<-c
	if ServerScreen != nil {
		ServerScreen.Destroy()
	}
	Server.Shutdown(context.TODO())
}

func waitForQuit() {
	ServerScreen.WaitForQuit()
	if ServerScreen != nil {
		ServerScreen.Destroy()
	}
	Server.Shutdown(context.TODO())

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
	Server = &http.Server{Addr: ":" + os.Args[2]}

	createMaze()
	http.HandleFunc("/login/", createUser)
	http.HandleFunc("/move/", moveUser)
	http.HandleFunc("/fire/", fire)
	http.HandleFunc("/place_bomb/", placeBomb)

	c := make(chan os.Signal, 1)

	signal.Notify(c, os.Interrupt)
	go waitForInterrupt(c)
	go waitForQuit()
	go UIFramesTick(1)
	Server.ListenAndServe()
	fmt.Println("Username", "Position")
	for _, obj := range GlobalObjectList {
		if obj.C == "HERO" {
			fmt.Println(obj.Username, obj.Dead)
		}
	}
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
	placeUser(msg.User)
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

func isUserDead(username string) bool {
	for _, obj := range GlobalObjectList {
		if obj.C == "HERO" && obj.Username == username {
			if obj.Dead > 0 {
				return true
			}
		}
	}
	return false
}

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

	if isUserDead(msg.User) {
		http.Error(w, "You are already dead", 403)
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
		http.Error(w, "Invalid", 409)
		return
	}

	success := MoveResponse{}
	success.Objects = GlobalObjectList

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

	if isUserDead(msg.User) {
		http.Error(w, "You are already dead", 403)
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

	for _, obj := range GlobalObjectList {
		if obj.Username == msg.User && obj.C == "BULLET" && obj.Dead == 0 {
			// already fired a bullet
			http.Error(w, "Forbidden", 403)
			mutex.Unlock()
			return
		}
	}
	done := make(chan int)
	Moves = append(Moves, models.Move{msg.User, "FIRE", msg.Direction, done})
	log("FIRE:" + msg.Direction)
	mutex.Unlock()

	rc := <-done
	if rc < 0 {
		http.Error(w, "Could not fire", 409)
		return
	}
	success := MoveResponse{}
	success.Objects = GlobalObjectList

	output, err := json.Marshal(success)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("content-type", "application/json")
	w.Write(output)

}

func placeBomb(w http.ResponseWriter, r *http.Request) {
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

	if isUserDead(msg.User) {
		http.Error(w, "You are already dead", 403)
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
	Moves = append(Moves, models.Move{msg.User, "PLACE", msg.Direction, done})

	mutex.Unlock()

	rc := <-done
	if rc < 0 {
		http.Error(w, "Could not place bomb", 409)
		return
	}
	success := MoveResponse{}
	success.Objects = GlobalObjectList

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

func isSpotEmpty(X, Y int) bool {
	if X <= 0 || X >= 50 || Y <= 0 || Y >= 50 {
		return false
	}
	for _, obj := range GlobalObjectList {
		if X == obj.X && Y == obj.Y {
			return false
		}
	}
	return true
}

func placeUser(username string) {
	newUser := models.Object{C: "HERO", Username: username, X: (rand.Int() % 50) + 1,
		Y: (rand.Int() % 50) + 1}
	for !isSpotEmpty(newUser.X, newUser.Y) {
		newUser = models.Object{C: "HERO", Username: username, X: (rand.Int() % 50) + 1,
			Y: (rand.Int() % 50) + 1}
	}
	GlobalObjectList = append(GlobalObjectList, newUser)
}

func createMaze() {
	for i := 0; i < 15; i++ {
		newWall := models.Object{C: "WALL", X: (rand.Int() % 50) + 1, Y: (rand.Int() % 50) + 1}
		for !isSpotEmpty(newWall.X, newWall.Y) {
			newWall = models.Object{C: "WALL", X: (rand.Int() % 50) + 1, Y: (rand.Int() % 50) + 1}
		}
		GlobalObjectList = append(GlobalObjectList, newWall)
	}

	newBombShed := models.Object{C: "BOMBSHED", X: (rand.Int() % 50) + 1, Y: (rand.Int() % 50) + 1}
	for !isSpotEmpty(newBombShed.X, newBombShed.Y) {
		newBombShed = models.Object{C: "BOMBSHED", X: (rand.Int() % 50) + 1, Y: (rand.Int() % 50) + 1}
	}
	GlobalObjectList = append(GlobalObjectList, newBombShed)

}
