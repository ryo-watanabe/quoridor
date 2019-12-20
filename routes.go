package main

import (
	"fmt"
	"strconv"
)

type route struct {
	steps []position
	curr position
}

func (r *route)Steps() int {
	return len(r.steps)
}

func (r *route)IsLooped(pos position) bool {
	for _, s := range(r.steps) {
		if s.Equals(pos) {
			return true
		}
	}
	return false
}

func (r *route)Push(pos position) {
	r.steps = append(r.steps, pos)
	r.curr = pos
}

func (r *route)Copy() *route {
	var newR route
	newR.steps = append(newR.steps, r.steps...)
	newR.curr = r.curr
	return &newR
}

func shortestRoute(board *QuoridorBoard, from position, goalY, max int) (int, bool) {
	routesTree := make([]route, 0)

	// init
	var r route
	r.Push(from)
	routesTree = append(routesTree, r)
	cnt := 0

	for len(routesTree) > 0 {
		newRoutes := make([]route, 0)
		goalRoutes := make([]route, 0)
		for _, r := range(routesTree) {
			left := position{X:r.curr.X-1, Y:r.curr.Y}
			if isOnBoard(left, board) && !isBlocked(left, r.curr, board) && !r.IsLooped(left) {
				newR := r.Copy()
				newR.Push(left)
				newRoutes = append(newRoutes, *newR)
			}
			right := position{X:r.curr.X+1, Y:r.curr.Y}
			if isOnBoard(right, board) && !isBlocked(right, r.curr, board) && !r.IsLooped(right) {
				newR := r.Copy()
				newR.Push(right)
				newRoutes = append(newRoutes, *newR)
			}
			up := position{X:r.curr.X, Y:r.curr.Y-1}
			if isOnBoard(up, board) && !isBlocked(up, r.curr, board) && !r.IsLooped(up) {
				newR := r.Copy()
				newR.Push(up)
				if up.Y == goalY {
					goalRoutes = append(goalRoutes, *newR)
				} else {
					newRoutes = append(newRoutes, *newR)
				}
			}
			down := position{X:r.curr.X, Y:r.curr.Y+1}
			if isOnBoard(down, board) && !isBlocked(down, r.curr, board) && !r.IsLooped(down) {
				newR := r.Copy()
				newR.Push(down)
				if down.Y == goalY {
					goalRoutes = append(goalRoutes, *newR)
				} else {
					newRoutes = append(newRoutes, *newR)
				}
			}
		}
		routesTree = nil
		if len(goalRoutes) > 0 {
			routesTree = append(routesTree, goalRoutes...)
			break
		}
		cnt++
		if len(newRoutes) == 0 || cnt > max {
			break
		}
		routesTree = append(routesTree, newRoutes...)
	}

	//if len(routesTree) > 0 {
	//	routeStr := ""
	//	for _, pos := range(routesTree[0].steps) {
	//		routeStr += " -> " + strconv.Itoa(pos.Y) + ":" + strconv.Itoa(pos.X)
	//	}
	//	fmt.Println(routeStr)
	//} else {
	//	fmt.Println("No routes")
	//}

	return cnt, len(routesTree) > 0
}
