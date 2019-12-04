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
)

type position struct {
	X int	`json:"x"`
	Y int	`json:"y"`
}

type wall struct {
	Pole position	`json:"pole"`
	Blocking [][]position	`json:"blocking"`
}

type QuoridorBoard struct {
	Dimension int	`json:"dimension"`
	Walls []wall	`json:"walls"`
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
		ComPos: position{X:0, Y:dim/2},
		PlayerPos: position{X:dim-1, Y:dim/2},
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
	if req.Action == "Dummy" {
		err := initBoard(req, ret)
		if err != nil {
			return err
		}
		newWall := wall{
			Pole: position{X:0, Y:1},
			Blocking: [][]position{
				{position{X:0, Y:0},position{X:0, Y:1}},
				{position{X:1, Y:0},position{X:1, Y:1}},
			},
		}
		ret.Board.Walls = append(ret.Board.Walls, newWall)
		newWall = wall{
			Pole: position{X:0, Y:2},
			Blocking: [][]position{
				{position{X:0, Y:1},position{X:1, Y:1}},
				{position{X:0, Y:2},position{X:1, Y:2}},
			},
		}
		ret.Board.Walls = append(ret.Board.Walls, newWall)
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

func main() {
	var contentsPath string
	flag.StringVar(&contentsPath, "contents-path", "/contents", "Set static contents path")

	// route handler
	http.HandleFunc("/api/", apiRequest)

	// route contents
	http.Handle("/", http.FileServer(http.Dir(contentsPath)))

	// do serve
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
