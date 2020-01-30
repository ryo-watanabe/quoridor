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

type QuoridorEvaluation struct {
	Eval int   `json:"eval"`
	PlayerSteps int   `json:"playerSteps"`
	ComSteps int   `json:"comSteps"`
	LookForward int   `json:"lookForward"`
	NumCases int   `json:"numCases"`
}

type QuoridorRequest struct {
	Action string	`json:"action"`
	Board *QuoridorBoard	`json:"board,omitempty"`
}

type QuoridorResponse struct {
	Status string	`json:"status"`
	Board *QuoridorBoard	`json:"board,omitempty"`
	Message string	`json:"message,omitempty"`
	Evaluation *QuoridorEvaluation   `json:"evaluation"`
}

func eval(player, com int) int {
	// simple:
	//return player - com

	// player route weight
	return 100*(player - com) + player
}

type turn struct {
	comMove *position
	playerMove *position
	wall *wall
}

type boardCase struct {
	turns []turn
	eval int
}

func (c *boardCase)appliedBorad(board *QuoridorBoard) *QuoridorBoard {
	b := board.Copy()
	for _, t := range(c.turns) {
		if t.comMove != nil {
			b.ComPos = *t.comMove
		}
		if t.playerMove != nil {
			b.PlayerPos = *t.playerMove
		}
		if t.wall != nil {
			b.Poles = append(b.Poles, t.wall.Pole)
			b.Blockings = append(b.Blockings, t.wall.Blockings...)
		}
	}
	return b
}

func compute(ret *QuoridorResponse, maxCases int) error {

	// 1st com turn
	cases := make([]boardCase,0)
	moves := possibleComMoves(ret.Board)
	for _, m := range(moves) {
		t := make([]turn, 0)
		t = append(t, turn{comMove:&m})
		cases = append(cases, boardCase{turns:t})
	}
	walls := possibleWalls(ret.Board)
	for _, w := range(walls) {
		t := make([]turn, 0)
		t = append(t, turn{wall:&w})
		cases = append(cases, boardCase{turns:t})
	}

	// 2nd player turn, and more...
	nextturn := true
	lookForward := 1
	newCases := make([]boardCase, 0)
	for nextturn {
		for _, c := range(cases) {
			b := c.appliedBorad(ret.Board)
			if lookForward % 2 != 0 {
				moves := possiblePlayerMoves(b)
				for _, m := range(moves) {
					t := make([]turn, 0)
					t = append(t, c.turns...)
					t = append(t, turn{playerMove:&m})
					newCases = append(newCases, boardCase{turns:t})
				}
			} else {
				moves := possibleComMoves(b)
				for _, m := range(moves) {
					t := make([]turn, 0)
					t = append(t, c.turns...)
					t = append(t, turn{comMove:&m})
					newCases = append(newCases, boardCase{turns:t})
				}
			}
			walls := possibleWalls(b)
			for _, w := range(walls) {
				t := make([]turn, 0)
				t = append(t, c.turns...)
				t = append(t, turn{wall:&w})
				newCases = append(newCases, boardCase{turns:t})
			}
		}
		if len(newCases) < maxCases {
			cases = make([]boardCase, 0)
			cases = append(cases, newCases...)
			lookForward++
		} else {
			nextturn = false
		}
		newCases = make([]boardCase, 0)
	}

	// Evaluation
	bestIndex := -1
	bestEval := -10000
	bestPlayerSteps := -1
	bestComSteps := -1
	max := maxRoute(ret.Board)
	for index, c := range(cases) {
		b := c.appliedBorad(ret.Board)
		com, com_ok := shortestTreeRoute(b, b.ComPos, ret.Board.Dimension-1, max)
		player, player_ok := shortestTreeRoute(b, b.PlayerPos, 0, max)
		if com_ok && player_ok {
			c.eval = eval(player, com)
			if c.eval > bestEval {
				bestEval = c.eval
				bestIndex = index
				bestPlayerSteps = player
				bestComSteps = com
			}
		}
	}
	if bestIndex < 0 {
		return fmt.Errorf("No possible move/wall")
	}

	// Set board
	if cases[bestIndex].turns[0].wall != nil {
		wall := cases[bestIndex].turns[0].wall
		ret.Board.Poles = append(ret.Board.Poles, wall.Pole)
		ret.Board.Blockings = append(ret.Board.Blockings, wall.Blockings...)
		ret.Board.ComWalls--;
	} else if cases[bestIndex].turns[0].comMove != nil {
		ret.Board.ComPos = *cases[bestIndex].turns[0].comMove
	} else {
		return fmt.Errorf("Compute error")
	}

	// Set debug values
	ret.Evaluation = &QuoridorEvaluation{
		Eval: bestEval,
		PlayerSteps: bestPlayerSteps,
		ComSteps: bestComSteps,
		LookForward: lookForward,
		NumCases: len(cases),
	}
	return nil
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
			ret.Status = "PLY"
			return nil
		}

		// computer won
		moves := possibleComMoves(ret.Board)
		for _, m := range(moves) {
			if m.Y == ret.Board.Dimension-1 {
				ret.Board.ComPos = m
				ret.Message = "Com won"
				ret.Status = "COM"
				return nil
			}
		}

		/*
		walls := possibleWalls(ret.Board)

		max := maxRoute(ret.Board)

		bestMoveEval := -10000
		bestMoveIndex := -1
		bestMovePlayerSteps := -1
		bestMoveComSteps := -1
		for index, m := range(moves) {
			com, com_ok := shortestTreeRoute(ret.Board, m, ret.Board.Dimension-1, max)
			player, player_ok := shortestTreeRoute(ret.Board, ret.Board.PlayerPos, 0, max)
			if !com_ok || !player_ok {
				continue
			}
			if eval(player, com) > bestMoveEval {
				bestMoveEval = eval(player, com)
				bestMoveIndex = index
				bestMovePlayerSteps = player
				bestMoveComSteps = com
			}
		}

		bestWallEval := -10000
		bestWallIndex := -1
		bestWallPlayerSteps := -1
		bestWallComSteps := -1
		if ret.Board.ComWalls > 0 {
			for index, w := range(walls) {
				testboard := ret.Board.Copy()
				testboard.Poles = append(testboard.Poles, w.Pole)
				testboard.Blockings = append(testboard.Blockings, w.Blockings...)

				com, com_ok := shortestTreeRoute(testboard, testboard.ComPos, testboard.Dimension-1, max)
				player, player_ok := shortestTreeRoute(testboard, testboard.PlayerPos, 0, max)
				if !com_ok || !player_ok {
					continue
				}
				if eval(player, com) > bestWallEval {
					bestWallEval = eval(player, com)
					bestWallIndex = index
					bestWallPlayerSteps = player
					bestWallComSteps = com
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
			ret.Evaluation = &QuoridorEvaluation{
				Eval: bestWallEval,
				PlayerSteps: bestWallPlayerSteps,
				ComSteps: bestWallComSteps,
				BestMoveEval: bestMoveEval,
				BestWallEval: bestWallEval,
			}
		} else {
			ret.Board.ComPos = moves[bestMoveIndex]
			ret.Evaluation = &QuoridorEvaluation{
				Eval: bestMoveEval,
				PlayerSteps: bestMovePlayerSteps,
				ComSteps: bestMoveComSteps,
				BestMoveEval: bestMoveEval,
				BestWallEval: bestWallEval,
			}
		}
		return nil
		*/

		return compute(ret, 5000)
	}
	if req.Action == "Rand" {
		ret.Board = req.Board
		rand.Seed(time.Now().UnixNano())
		moves := possibleComMoves(ret.Board)
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
