package main

import (
	//"fmt"
	//"strconv"
)

type tree struct {
	steps [][]position
	curr int
	goalY int
	max int
}

func newTree(start position, goalY int, max int) *tree {
	initSteps := make([][]position, 0)
	currStep := make([]position, 0)
	currStep = append(currStep, start)
	initSteps = append(initSteps, currStep)
	return &tree{
		steps: initSteps,
		curr: 0,
		goalY: goalY,
		max: max,
	}
}

func (t *tree)push(pos position) {
	t.steps[t.curr] = append(t.steps[t.curr], pos)
}

func (t *tree)next() (bool, bool) {
	if len(t.steps[t.curr]) == 0 || t.curr + 1 > t.max {
		return false, false
	}
	for _, p := range(t.steps[t.curr]) {
		if p.Y == t.goalY {
			return true, false
		}
	}
	nextStep := make([]position, 0)
	t.steps = append(t.steps, nextStep)
	t.curr++
	return false, true
}

func (t *tree)isLooped(pos position) bool {
	for _, s := range(t.steps) {
		for _, p := range(s) {
			if p.Equals(pos) {
				return true
			}
		}
	}
	return false
}

func (t *tree)activeSteps() []position {
	return t.steps[t.curr - 1]
}

func (t *tree)length() int {
	return t.curr
}

func shortestTreeRoute(board *QuoridorBoard, from position, goalY, max int) (int, bool) {
	routesTree := newTree(from, goalY, max)
	var gonext, goal bool

	goal, gonext = routesTree.next()
	for gonext {
		for _, s := range(routesTree.activeSteps()) {
			left := position{X:s.X-1, Y:s.Y}
			if isOnBoard(left, board) && !isBlocked(left, s, board) && !routesTree.isLooped(left) {
				routesTree.push(left)
			}
			right := position{X:s.X+1, Y:s.Y}
			if isOnBoard(right, board) && !isBlocked(right, s, board) && !routesTree.isLooped(right) {
				routesTree.push(right)
			}
			up := position{X:s.X, Y:s.Y-1}
			if isOnBoard(up, board) && !isBlocked(up, s, board) && !routesTree.isLooped(up) {
				routesTree.push(up)
			}
			down := position{X:s.X, Y:s.Y+1}
			if isOnBoard(down, board) && !isBlocked(down, s, board) && !routesTree.isLooped(down) {
				routesTree.push(down)
			}
		}
		goal, gonext = routesTree.next()
	}

	//fmt.Printf("TREE : length %d, goal %t\n", routesTree.length(), goal)
	//for index, step := range(routesTree.steps) {
	//	fmt.Printf("  step %3d :", index)
	//	for _, p := range(step) {
	//		fmt.Printf(" [%2d:%2d]", p.X, p.Y)
	//	}
	//	fmt.Printf("\n")
	//}

	return routesTree.length(), goal
}
