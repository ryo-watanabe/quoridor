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
	NextPlayerEval int   `json:"nextPlayerEval"`
	NumCases int   `json:"numCases"`
	NumNextPlayerCases int   `json:"numNextPlayerCases"`
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
	return player - com

	// player route weight
	//return 100*(player - com) + player
}

type turn struct {
	comMove *position
	playerMove *position
	wall *wall
}

func (t *turn)Print() {
	if t.playerMove != nil {
		fmt.Printf("ply>%d/%d ", t.playerMove.X, t.playerMove.Y)
	}
	if t.comMove != nil {
		fmt.Printf("com>%d/%d ", t.comMove.X, t.comMove.Y)
	}
	if t.wall != nil {
		fmt.Printf("blk:%d/%d:%d/%d+%d/%d:%d/%d ",
			t.wall.Blockings[0][0].X, t.wall.Blockings[0][0].Y, t.wall.Blockings[0][1].X, t.wall.Blockings[0][1].Y,
			t.wall.Blockings[1][0].X, t.wall.Blockings[1][0].Y, t.wall.Blockings[1][1].X, t.wall.Blockings[1][1].Y)
	}
}

type boardCase struct {
	turn turn
	eval int
	prev *boardCase
	nextEval int
	com int
	player int
}

func (c *boardCase)appliedBorad(board *QuoridorBoard) *QuoridorBoard {
	var b *QuoridorBoard
	if c.prev != nil {
		b = c.prev.appliedBorad(board)
	} else {
		b = board.Copy()
	}
	if c.turn.comMove != nil {
		b.ComPos = *c.turn.comMove
	}
	if c.turn.playerMove != nil {
		b.PlayerPos = *c.turn.playerMove
	}
	if c.turn.wall != nil {
		b.Poles = append(b.Poles, c.turn.wall.Pole)
		b.Blockings = append(b.Blockings, c.turn.wall.Blockings...)
	}
	return b
}

func playout(board *QuoridorBoard) (bool, int, error) {
	max := maxRoute(board)
	cnt := 0
	routeSearch := 0
	for cnt < 100 {

		// Player turn
		moves := possiblePlayerMoves(board)
		for _, m := range(moves) {
			// player won
			if m.Y == 0 {
				return false, routeSearch, nil
			}
		}
		cases := make([]boardCase,0)
		for i := 0; i < len(moves); i++ {
			cases = append(cases, boardCase{turn:turn{playerMove:&moves[i],comMove:nil,wall:nil},eval:-10000,prev:nil})
		}
		if board.PlayerWalls > 0 {
			walls := possibleWalls(board)
			for i := 0; i < len(walls); i++ {
				cases = append(cases, boardCase{turn:turn{playerMove:nil,comMove:nil,wall:&walls[i]},eval:-10000,prev:nil})
			}
		}
		// Player evaluate
		bestEval := -1000
		bestIndices := make([]int, 0)
		for i := 0; i < len(cases); i++ {
			b := cases[i].appliedBorad(board)
			com, com_ok := shortestTreeRoute(b, b.ComPos, board.Dimension-1, max)
			player, player_ok := shortestTreeRoute(b, b.PlayerPos, 0, max)
			if com_ok && player_ok {
				cases[i].eval = eval(com, player)
				cases[i].com = com
				cases[i].player = player
			}
			if cases[i].eval > bestEval {
				bestEval = cases[i].eval
				bestIndices = make([]int, 0)
			}
			if cases[i].eval >= bestEval {
				bestIndices = append(bestIndices, i)
			}
		}
		routeSearch += len(cases)
		if len(bestIndices) == 0 {
			return false, routeSearch, fmt.Errorf("No possible move/wall in playout - Player evaluate")
		}
		// Random player
		index := bestIndices[rand.Intn(len(bestIndices))]
		if cases[index].turn.wall != nil {
			wall := cases[index].turn.wall
			board.Poles = append(board.Poles, wall.Pole)
			board.Blockings = append(board.Blockings, wall.Blockings...)
			board.PlayerWalls--;
		} else if cases[index].turn.playerMove != nil {
			board.PlayerPos = *cases[index].turn.playerMove
		} else {
			return false, routeSearch, fmt.Errorf("Compute error in playout - Random player")
		}

		//fmt.Printf("%03d PLY - ", cnt)
		//cases[index].turn.Print()
		//fmt.Printf("ply:%d com:%d selected from [", cases[index].player, cases[index].com)
		//for i := 0; i < len(bestIndices); i++ {
		//	cases[bestIndices[i]].turn.Print()
		//	fmt.Printf("ply:%d com:%d, ", cases[bestIndices[i]].player, cases[bestIndices[i]].com)
		//}
		//fmt.Printf("]\n")

		// Com turn
		moves = possibleComMoves(board)
		for _, m := range(moves) {
			// computer won
			if m.Y == board.Dimension-1 {
				return true, routeSearch, nil
			}
		}
		cases = make([]boardCase,0)
		for i := 0; i < len(moves); i++ {
			cases = append(cases, boardCase{turn:turn{playerMove:nil,comMove:&moves[i],wall:nil},eval:-10000,prev:nil})
		}
		if board.ComWalls > 0 {
			walls := possibleWalls(board)
			for i := 0; i < len(walls); i++ {
				cases = append(cases, boardCase{turn:turn{playerMove:nil,comMove:nil,wall:&walls[i]},eval:-10000,prev:nil})
			}
		}
		// Com evaluate
		bestEval = -1000
		bestIndices = make([]int, 0)
		for i := 0; i < len(cases); i++ {
			b := cases[i].appliedBorad(board)
			com, com_ok := shortestTreeRoute(b, b.ComPos, board.Dimension-1, max)
			player, player_ok := shortestTreeRoute(b, b.PlayerPos, 0, max)
			if com_ok && player_ok {
				cases[i].eval = eval(player, com)
				cases[i].com = com
				cases[i].player = player
			}
			if cases[i].eval > bestEval {
				bestEval = cases[i].eval
				bestIndices = make([]int, 0)
			}
			if cases[i].eval >= bestEval {
				bestIndices = append(bestIndices, i)
			}
		}
		if len(bestIndices) == 0 {
			return false, routeSearch, fmt.Errorf("No possible move/wall in playout - Com evaluate")
		}
		routeSearch += len(cases)
		// Random com
		index = bestIndices[rand.Intn(len(bestIndices))]
		if cases[index].turn.wall != nil {
			wall := cases[index].turn.wall
			board.Poles = append(board.Poles, wall.Pole)
			board.Blockings = append(board.Blockings, wall.Blockings...)
			board.ComWalls--;
		} else if cases[index].turn.comMove != nil {
			board.ComPos = *cases[index].turn.comMove
		} else {
			return false, routeSearch, fmt.Errorf("Compute error in playout - Random com")
		}

		//fmt.Printf("%03d COM - ", cnt)
		//cases[index].turn.Print()
		//fmt.Printf("ply:%d com:%d selected from [", cases[index].player, cases[index].com)
		//for i := 0; i < len(bestIndices); i++ {
		//	cases[bestIndices[i]].turn.Print()
		//	fmt.Printf("ply:%d com:%d, ", cases[bestIndices[i]].player, cases[bestIndices[i]].com)
		//}
		//fmt.Printf("]\n")

		cnt++
	}
	return false, routeSearch, fmt.Errorf("Reached limit in playout %d", cnt)
}

func prob_compute(ret *QuoridorResponse) error {

	max := maxRoute(ret.Board)

	// com turn
	cases := make([]boardCase,0)
	moves := possibleComMoves(ret.Board)
	for i := 0; i < len(moves); i++ {
		cases = append(cases, boardCase{turn:turn{playerMove:nil,comMove:&moves[i],wall:nil},eval:-10000,nextEval:0,prev:nil})
	}
	if ret.Board.ComWalls > 0 {
		walls := possibleWalls(ret.Board)
		for i := 0; i < len(walls); i++ {
			cases = append(cases, boardCase{turn:turn{playerMove:nil,comMove:nil,wall:&walls[i]},eval:-10000,nextEval:0,prev:nil})
		}
	}
	// evaluate by route
	bestEval := -1000
	bestIndices := make([]int, 0)
	for i := 0; i < len(cases); i++ {
		b := cases[i].appliedBorad(ret.Board)
		com, com_ok := shortestTreeRoute(b, b.ComPos, ret.Board.Dimension-1, max)
		player, player_ok := shortestTreeRoute(b, b.PlayerPos, 0, max)
		if com_ok && player_ok {
			cases[i].eval = eval(player, com)
			cases[i].com = com
			cases[i].player = player
		}
		if cases[i].eval > bestEval {
			bestEval = cases[i].eval
			bestIndices = make([]int, 0)
		}
		if cases[i].eval >= bestEval {
			bestIndices = append(bestIndices, i)
		}
	}
	if len(bestIndices) == 0 {
		return fmt.Errorf("No possible move/wall")
	}
	// evaluate by playouts
	routeSearch := 0
	playouts := 0
	for j := 0; j < 100; j++ {
		for i := 0; i < len(bestIndices); i++ {

			b := cases[bestIndices[i]].appliedBorad(ret.Board)

			//fmt.Printf("Playout %02d for: ", j)
			//cases[bestIndices[i]].turn.Print()
			//fmt.Printf("\n")

			win, search, err := playout(b.Copy())
			if err != nil {
				return err
			}
			if win {
				cases[bestIndices[i]].nextEval++
			}
			routeSearch += search
		}
		playouts++
		if routeSearch > 100000 {
			break
		}
	}
	bestPlayout := -1
	bestIndex := -1
	for i := 0; i < len(bestIndices); i++ {
		if cases[bestIndices[i]].nextEval > bestPlayout {
			bestPlayout = cases[bestIndices[i]].nextEval
			bestIndex = bestIndices[i]
		}
	}
	if bestIndex < 0 || playouts == 0 {
		return fmt.Errorf("No best playout")
	}
	// Set board
	if cases[bestIndex].turn.wall != nil {
		wall := cases[bestIndex].turn.wall
		ret.Board.Poles = append(ret.Board.Poles, wall.Pole)
		ret.Board.Blockings = append(ret.Board.Blockings, wall.Blockings...)
		ret.Board.ComWalls--;
	} else if cases[bestIndex].turn.comMove != nil {
		ret.Board.ComPos = *cases[bestIndex].turn.comMove
	} else {
		return fmt.Errorf("Prob Compute error")
	}
	// Set debug values
	ret.Evaluation = &QuoridorEvaluation{
		Eval: bestPlayout * 100 / playouts,
		NumCases: routeSearch,
	}
	return nil
}

func compute(ret *QuoridorResponse, careNext bool) error {

	max := maxRoute(ret.Board)

	// 1st com turn
	cases := make([]boardCase,0)
	moves := possibleComMoves(ret.Board)
	for i := 0; i < len(moves); i++ {
		cases = append(cases, boardCase{turn:turn{playerMove:nil,comMove:&moves[i],wall:nil},eval:-10000,nextEval:-10000,prev:nil})
	}
	if ret.Board.ComWalls > 0 {
		walls := possibleWalls(ret.Board)
		for i := 0; i < len(walls); i++ {
			cases = append(cases, boardCase{turn:turn{playerMove:nil,comMove:nil,wall:&walls[i]},eval:-10000,nextEval:-10000,prev:nil})
		}
	}

	// evaluate
	bestEvalFirst := -10000
	bestIndexFirst := -1
	numBestCases := 0
	for i := 0; i < len(cases); i++ {
		b := cases[i].appliedBorad(ret.Board)
		com, com_ok := shortestTreeRoute(b, b.ComPos, ret.Board.Dimension-1, max)
		player, player_ok := shortestTreeRoute(b, b.PlayerPos, 0, max)
		if com_ok && player_ok {
			cases[i].eval = eval(player, com)
			cases[i].com = com
			cases[i].player = player
		}
		if cases[i].eval > bestEvalFirst {
			bestEvalFirst = cases[i].eval
			bestIndexFirst = i
			numBestCases = 1
		} else if cases[i].eval == bestEvalFirst {
			numBestCases++
		}
	}
	if bestIndexFirst < 0 {
		return fmt.Errorf("No possible move/wall")
	}

	if !careNext {
		// Set board
		if cases[bestIndexFirst].turn.wall != nil {
			wall := cases[bestIndexFirst].turn.wall
			ret.Board.Poles = append(ret.Board.Poles, wall.Pole)
			ret.Board.Blockings = append(ret.Board.Blockings, wall.Blockings...)
			ret.Board.ComWalls--;
		} else if cases[bestIndexFirst].turn.comMove != nil {
			ret.Board.ComPos = *cases[bestIndexFirst].turn.comMove
		} else {
			return fmt.Errorf("Compute error")
		}

		// Set debug values
		ret.Evaluation = &QuoridorEvaluation{
			Eval: bestEvalFirst,
			NumCases: numBestCases,
		}
		return nil
	}

	// 2nd player turn
	newCases := make([]boardCase, 0)
	numBestCases = 0
	for i := 0; i < len(cases); i++ {
		if cases[i].eval < bestEvalFirst {
			continue
		}
		numBestCases++
		b := cases[i].appliedBorad(ret.Board)
		moves := possiblePlayerMoves(b)
		for j := 0; j < len(moves); j++ {
			newCases = append(newCases, boardCase{turn:turn{playerMove:&moves[j]},eval:-10000,nextEval:-10000,prev:&cases[i]})
		}
		if ret.Board.PlayerWalls > 0 {
			walls := possibleWalls(b)
			for j := 0; j < len(walls); j++ {
				newCases = append(newCases, boardCase{turn:turn{wall:&walls[j]},eval:-10000,nextEval:-10000,prev:&cases[i]})
			}
		}
	}

	// Evaluation
	for i := 0; i < len(newCases); i++ {
		b := newCases[i].appliedBorad(ret.Board)
		com, com_ok := shortestTreeRoute(b, b.ComPos, ret.Board.Dimension-1, max)
		player, player_ok := shortestTreeRoute(b, b.PlayerPos, 0, max)
		if com_ok && player_ok {
			newCases[i].eval = eval(com, player)
			if newCases[i].eval > newCases[i].prev.nextEval {
				newCases[i].prev.nextEval = newCases[i].eval
			}
		}
	}
	bestIndex := -1
	bestEval := 10000
	numNextBestCases := 0
	for index, c := range(cases) {
		if c.nextEval > -10000 && c.nextEval < bestEval {
			bestEval = c.nextEval
			bestIndex = index
			numNextBestCases = 1
		} else if c.nextEval > -10000 && c.nextEval == bestEval {
			numNextBestCases++
		}
	}
	if bestIndex < 0 {
		return fmt.Errorf("No best move/wall")
	}

	// Set board
	if cases[bestIndex].turn.wall != nil {
		wall := cases[bestIndex].turn.wall
		ret.Board.Poles = append(ret.Board.Poles, wall.Pole)
		ret.Board.Blockings = append(ret.Board.Blockings, wall.Blockings...)
		ret.Board.ComWalls--;
	} else if cases[bestIndex].turn.comMove != nil {
		ret.Board.ComPos = *cases[bestIndex].turn.comMove
	} else {
		return fmt.Errorf("Compute error")
	}

	// Set debug values
	ret.Evaluation = &QuoridorEvaluation{
		Eval: bestEvalFirst,
		NextPlayerEval: bestEval,
		NumCases: numBestCases,
		NumNextPlayerCases: numNextBestCases,
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

		// compute
		// return compute(ret, true)
		rand.Seed(time.Now().UnixNano())
		return prob_compute(ret)
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
