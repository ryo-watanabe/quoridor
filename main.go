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

type position struct {
	X int	`json:"x"`
	Y int	`json:"y"`
}

func (p *position)Equals(pos position) bool {
	return (p.X == pos.X && p.Y == pos.Y)
}

type wall struct {
	Pole position	`json:"poles"`
	Blockings [][]position	`json:"blockings"`
}

type QuoridorBoard struct {
	Dimension int	`json:"dimension"`
	Poles []position	`json:"poles"`
	Blockings [][]position	`json:"blockings"`
	PlayerPos position	`json:"playerPos"`
	ComPos position	`json:"comPos"`
}

type QuoridorRequest struct {
	Action string	`json:"action"`
	Board *QuoridorBoard	`json:"board,omitempty"`
}

type QuoridorResponse struct {
	Status string	`json:"status"`
	Board *QuoridorBoard	`json:"board,omitempty"`
	Message string	`json:"message,omitempty"`
}

func (b *QuoridorBoard)Copy() *QuoridorBoard {
	var board QuoridorBoard
	board = *b
	board.Poles = make([]position,0)
	board.Poles = append(board.Poles, b.Poles...)
	board.Blockings = make([][]position,0)
	for _, block := range(b.Blockings) {
		newBlock := make([]position,0)
		newBlock = append(newBlock, block...)
		board.Blockings = append(board.Blockings, newBlock)
	}
	return &board
}

func maxRoute(board *QuoridorBoard) int {
	return board.Dimension*board.Dimension
}

func isBlocked(a, b position, board *QuoridorBoard) bool {
	for _, block := range(board.Blockings) {
		if (a.Equals(block[0]) && b.Equals(block[1])) ||
		   (a.Equals(block[1]) && b.Equals(block[0])) {
			return true
		}
	}
	return false
}

func isOnBoard(pos position, board *QuoridorBoard) bool {
	if pos.X < 0 || pos.X >= board.Dimension || pos.Y < 0 || pos.Y >= board.Dimension {
		return false
	}
	return true
}

func possibleMovesFrom(pos position, board *QuoridorBoard) []position {
	moves := make([]position, 0)
	// Right
	moveright := position{X:pos.X+1, Y:pos.Y}
	if moveright.X < board.Dimension && !isBlocked(pos, moveright, board) {
		moves = append(moves, moveright)
	}
	// Left
	moveleft := position{X:pos.X-1, Y:pos.Y}
	if moveleft.X > 0 && !isBlocked(pos, moveleft, board) {
		moves = append(moves, moveleft)
	}
	// Up
	moveup := position{X:pos.X, Y:pos.Y-1}
	if moveup.Y > 0 && !isBlocked(pos, moveup, board) {
		moves = append(moves, moveup)
	}
	// Down
	movedown := position{X:pos.X, Y:pos.Y+1}
	if movedown.Y < board.Dimension && !isBlocked(pos, movedown, board) {
		moves = append(moves, movedown)
	}
	return moves
}

func possibleMoves(board *QuoridorBoard) []position {
	moves := make([]position, 0)
	comMoves := possibleMovesFrom(board.ComPos, board)
	for _, m := range(comMoves) {
		// Jump over player pos
		if m.Equals(board.PlayerPos) {
			playerMoves := possibleMovesFrom(m, board)
			for _, pm := range(playerMoves) {
				if !pm.Equals(board.ComPos) {
					moves = append(moves, pm)
				}
			}
		} else {
			moves = append(moves, m)
		}
	}
	return moves
}

func isPossible(w *wall, board *QuoridorBoard) bool {
	for _, builtPole := range(board.Poles) {
		if builtPole.Equals(w.Pole) {
			return false
		}
	}
	if isBlocked(w.Blockings[0][0], w.Blockings[0][1], board) {
		return false
	}
	if isBlocked(w.Blockings[1][0], w.Blockings[1][1], board) {
		return false
	}
	return true
}

func possibleWalls(board *QuoridorBoard) []wall {
	walls := make([]wall, 0)
	for i := 0; i < board.Dimension - 1; i++ {
		for j := 0; j < board.Dimension - 1; j++ {
			// vertical
			vw := wall{
				Pole: position{Y:i, X:j},
				Blockings: [][]position{
					{position{Y:i, X:j},position{Y:i, X:j+1}},
					{position{Y:i+1, X:j},position{Y:i+1, X:j+1}},
				},
			}
			if isPossible(&vw, board) {
				walls = append(walls, vw)
			}
			// horizontal
			hw := wall{
				Pole: position{Y:i, X:j},
				Blockings: [][]position{
					{position{Y:i, X:j},position{Y:i+1, X:j}},
					{position{Y:i, X:j+1},position{Y:i+1, X:j+1}},
				},
			}
			if isPossible(&hw, board) {
				walls = append(walls, hw)
			}
		}
	}
	return walls
}

func initBoard(req *QuoridorRequest, ret *QuoridorResponse) error {
	dim := 5
	if req.Board != nil && req.Board.Dimension != 0 {
		if req.Board.Dimension > 4  && req.Board.Dimension < 14 {
			dim = req.Board.Dimension
		} else {
			return fmt.Errorf("Invalid dimension %d", req.Board.Dimension)
		}
	}
	ret.Board = &QuoridorBoard{
		Dimension: dim,
		ComPos: position{Y:0, X:dim/2},
		PlayerPos: position{Y:dim-1, X:dim/2},
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

		if bestWallIndex < 0 && bestMoveIndex < 0 {
			return fmt.Errorf("No possible move/wall")
		}
		if bestWallIndex >= 0 && (bestMoveIndex < 0 || bestMoveEval < bestWallEval) {
			wall := walls[bestWallIndex]
			ret.Board.Poles = append(ret.Board.Poles, wall.Pole)
			ret.Board.Blockings = append(ret.Board.Blockings, wall.Blockings...)
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
