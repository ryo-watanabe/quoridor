package main

import (
	"fmt"
	"net/http"
	"encoding/json"
	"os"
	"flag"
	//"strconv"
	"io"
	"bufio"
	"crypto/subtle"
	"math/rand"
	"time"
)

type QuoridorRequest struct {
	Action string	`json:"action"`
	Board *QuoridorBoard	`json:"board,omitempty"`
}

type QuoridorResponse struct {
	Status string	`json:"status"`
	Board *QuoridorBoard	`json:"board,omitempty"`
	Message string	`json:"message,omitempty"`
}

func action(req *QuoridorRequest, ret *QuoridorResponse) error {
	if req.Action == "Init" {
		err := initBoard(req, ret)
		if err != nil {
			return err
		}
		return nil
	}
	if req.Action == "Com" {
		ret.Board = req.Board

		// player won
		if ret.Board.PlayerPos.Y == 0 {
			ret.Message = "Player won"
			return nil
		}

		// computer won
		moves := possibleMoves(ret.Board)
		for _, m := range(moves) {
			if m.Y == ret.Board.Dimension-1 {
				ret.Board.ComPos = m
				ret.Message = "Com won"
				return nil
			}
		}

		walls := possibleWalls(ret.Board)

		max := maxRoute(ret.Board)

		bestMoveEval := -100
		bestMoveIndex := -1
		for index, m := range(moves) {
			com, com_ok := shortestRoute(ret.Board, m, ret.Board.Dimension-1, max)
			player, player_ok := shortestRoute(ret.Board, ret.Board.PlayerPos, 0, max)
			if !com_ok || !player_ok {
				continue
			}
			if player - com > bestMoveEval {
				bestMoveEval = player - com
				bestMoveIndex = index
			}
		}

		bestWallEval := -100
		bestWallIndex := -1
		if ret.Board.ComWalls > 0 {
			for index, w := range(walls) {
				testboard := ret.Board.Copy()
				testboard.Poles = append(testboard.Poles, w.Pole)
				testboard.Blockings = append(testboard.Blockings, w.Blockings...)

				com, com_ok := shortestRoute(testboard, testboard.ComPos, testboard.Dimension-1, max)
				player, player_ok := shortestRoute(testboard, testboard.PlayerPos, 0, max)
				if !com_ok || !player_ok {
					continue
				}
				if player - com > bestWallEval {
					bestWallEval = player - com
					bestWallIndex = index
				}
			}
		}

		if bestWallIndex < 0 && bestMoveIndex < 0 {
			return fmt.Errorf("No possible move/wall")
		}
		if bestWallIndex >= 0 && (bestMoveIndex < 0 || bestMoveEval < bestWallEval) {
			wall := walls[bestWallIndex]
			ret.Board.Poles = append(ret.Board.Poles, wall.Pole)
			ret.Board.Blockings = append(ret.Board.Blockings, wall.Blockings...)
			ret.Board.ComWalls--;
		} else {
			ret.Board.ComPos = moves[bestMoveIndex]
		}
		return nil
	}
	if req.Action == "Rand" {
		ret.Board = req.Board
		rand.Seed(time.Now().UnixNano())
		moves := possibleMoves(ret.Board)
		walls := possibleWalls(ret.Board)
		if rand.Intn(2) == 0 && len(moves) > 0 {
			ret.Board.ComPos = moves[rand.Intn(len(moves))]
		} else if len(walls) > 0 {
			wall := walls[rand.Intn(len(walls))]
			ret.Board.Poles = append(ret.Board.Poles, wall.Pole)
			ret.Board.Blockings = append(ret.Board.Blockings, wall.Blockings...)
		}
		return nil
	}
	if req.Action == "Dummy" {
		err := initBoard(req, ret)
		if err != nil {
			return err
		}
		ret.Board.Poles = append(ret.Board.Poles, position{X:0, Y:0})
		ret.Board.Blockings = append(ret.Board.Blockings, []position{position{X:0, Y:0},position{X:0, Y:1}})
		ret.Board.Blockings = append(ret.Board.Blockings, []position{position{X:1, Y:0},position{X:1, Y:1}})

		ret.Board.Poles = append(ret.Board.Poles, position{X:0, Y:1})
		ret.Board.Blockings = append(ret.Board.Blockings, []position{position{X:0, Y:1},position{X:1, Y:1}})
		ret.Board.Blockings = append(ret.Board.Blockings, []position{position{X:0, Y:2},position{X:1, Y:2}})

		ret.Board.ComPos = position{X:1, Y:2}
		ret.Board.PlayerPos = position{X:3, Y:0}

		return nil
	}

	return fmt.Errorf("Invalid action %s", req.Action)
}

func apiRequest(w http.ResponseWriter, r *http.Request) {
	ret := QuoridorResponse{Status:"OK", Message:""}
	request := ""

	// JSON return
	defer func() {
		// result
		outjson,err := json.Marshal(ret)
		if err != nil {
			fmt.Println(err) //TODO: change to log
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, string(outjson))
	}()

	// type check
	if r.Method != "POST" {
		ret.Status = "NG"
		ret.Message = "Not POST method"
		return
	}

	// request body
	rb := bufio.NewReader(r.Body)
	for {
		s, err := rb.ReadString('\n')
		request = request + s
		if err == io.EOF { break }
	}

	// JSON parse
	var req QuoridorRequest
	b := []byte(request)
	err := json.Unmarshal(b, &req)
	if err != nil {
		ret.Status = "NG"
		ret.Message = "JSON parse error."
		return
	}

	// do action
	err = action(&req, &ret)
	if err != nil {
		ret.Status = "NG"
		ret.Message = err.Error()
		return
	}

}

func BasicAuth(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok || subtle.ConstantTimeCompare([]byte(user), []byte(username)) != 1 || subtle.ConstantTimeCompare([]byte(pass), []byte(password)) != 1 {
			w.Header().Set("WWW-Authenticate", `Basic realm="`+realm+`"`)
			w.WriteHeader(401)
			w.Write([]byte("Unauthorised.\n"))
			return
		}
		handler(w, r)
	}
}

var (
	realm = "Please enter your username and password"
	contentsPath, username, password string
)

func main() {
	flag.StringVar(&contentsPath, "contents-path", "/contents", "Set static contents path")
	flag.StringVar(&username, "username", "user", "Basic auth username")
	flag.StringVar(&password, "password", "pass", "Basic auth password")
	flag.Parse()

	// route handler
	http.HandleFunc("/api/", BasicAuth(apiRequest))

	// route contents
	http.HandleFunc("/", BasicAuth(func(w http.ResponseWriter, r *http.Request) {
		http.FileServer(http.Dir(contentsPath)).ServeHTTP(w, r)
	}))

	// do serve
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
