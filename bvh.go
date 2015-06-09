package main

import (
	"fmt"
	"github.com/pkg/profile"
	"math"
	"math/rand"
	"runtime"
	"time"
)

type AABB struct {
	MinX float64
	MinY float64
	MaxX float64
	MaxY float64
}

func (a1 *AABB) Merge(a2 *AABB) *AABB {
	return &AABB{
		MinX: math.Min(a1.MinX, a2.MinX),
		MaxX: math.Max(a1.MaxX, a2.MaxX),
		MinY: math.Min(a1.MinY, a2.MinY),
		MaxY: math.Max(a1.MaxY, a2.MaxY)}
}

func (a *AABB) Area() float64 {
	dy := a.MaxY - a.MinY
	dx := a.MaxX - a.MinX
	return dy * dx
}

func (a1 *AABB) Intersects(a2 *AABB) bool {

	if a1.MinX > a2.MaxX || a1.MaxX < a2.MinX || a1.MinY > a2.MaxY || a1.MaxY < a2.MinY {
		return false
	}
	return true
}

type Square struct {
	Bounds *AABB
}

func (s *Square) GetBounds() *AABB {
	return s.Bounds
}

func NewSquare(x1 float64, y1 float64, x2 float64, y2 float64) *Square {
	bounds := &AABB{
		MinX: math.Min(x1, x2),
		MaxX: math.Max(x1, x2),
		MinY: math.Min(y1, y2),
		MaxY: math.Max(y1, y2)}
	return &Square{Bounds: bounds}
}

type Bounded interface {
	GetBounds() *AABB
}

type Node struct {
	IsBranch bool
	Bounds   *AABB
	Children []Bounded
}

func (n *Node) GetBounds() *AABB {
	return n.Bounds
}

func (n *Node) IsLeaf() bool {
	return !n.IsBranch
}

func (n *Node) Scan(b Bounded) []Bounded {

	results := []Bounded{}

	if !n.IsLeaf() {

		for _, child := range n.Children {
			if child.GetBounds().Intersects(b.GetBounds()) {
				results = append(results, castNode(child).Scan(b)...)
			}
		}

	} else {
		for _, child := range n.Children {
			if b != child && child.GetBounds().Intersects(b.GetBounds()) {
				results = append(results, child)
			}
		}
	}

	return results
}

func (n *Node) Insert(b Bounded) {

	if len(n.Children) > 0 {
		n.Bounds = n.Bounds.Merge(b.GetBounds())
	} else {
		n.Bounds = b.GetBounds()
	}

	if n.IsLeaf() {
		if len(n.Children) < 2 {
			n.Children = append(n.Children, b)
			return
		} else {
			children := n.Children
			n.Children = []Bounded{}
			n0 := &Node{}
			n1 := &Node{}

			n0.Insert(children[0])
			n1.Insert(children[1])
			n.Children = append(n.Children, n0, n1)

			n.IsBranch = true
		}
	}

	var minBounding float64 = 999999
	minIndex := -1

	for index, branch := range n.Children {
		bounds := branch.GetBounds().Merge(b.GetBounds()).Area()
		if bounds < minBounding {
			minBounding = bounds
			minIndex = index
		}
	}

	castNode(n.Children[minIndex]).Insert(b)
}

/*
func (n *Node) Scan(b Bounded) []Bounded {

	results := []Bounded{}

	if n.IsLeaf() {

	}
}
*/
func castNode(n Bounded) *Node {
	switch t := n.(type) {
	case *Node:
		return t
	default:
		panic("cant cast to node")
	}

}

func main() {

	runtime.GOMAXPROCS(runtime.NumCPU())

	p := profile.Start(profile.CPUProfile, profile.ProfilePath("."), profile.NoShutdownHook)
	defer p.Stop()

	squares := []*Square{}
	width := float64(10)
	height := float64(10)

	for i := 0; i < 5000; i++ {

		x1 := float64(rand.Intn(1000))
		y1 := float64(rand.Intn(1000))
		square := NewSquare(x1, y1, x1+width, y1+height)
		//square := &Square{X1: x1, Y1: y1, X2: x1 + width, Y2: y1 + height}
		//root.Insert(square)
		squares = append(squares, square)
	}

	for i := 0; i < 100; i++ {
		start := time.Now()

		collisions := 0

		root := &Node{Bounds: &AABB{0, 0, 1050, 1050}}

		for _, s := range squares {
			root.Insert(s)
		}

		insertiondone := time.Since(start)

		for _, s := range squares {
			collisions += len(root.Scan(s))
		}

		elapsed := time.Since(start)
		fmt.Println("collisions:", collisions, "elapsed time:", elapsed.String(), "insert time:", insertiondone)
	}

	/*
		for i := 0; i < 10; i++ {
			start := time.Now()
			collisions := 0

			for j := 1; j < len(squares); j++ {
				for k := 0; k < j; k++ {
					if squares[j] != squares[k] && squares[j].GetBounds().Intersects(squares[k].GetBounds()) {
						collisions++
					}
				}
			}

			elapsed := time.Since(start)
			fmt.Println("collisions:", collisions, "elapsed time:", elapsed.String())
		}
	*/

	/*
		broad := root.Scan(test)

		fmt.Println("collisions:")
		for _, square := range broad {
			fmt.Println(square)
		}

		fmt.Println("manual collisions detected", col)
		fmt.Println("bvh collisons detected", len(broad))
	*/
}

func printBorder(n *Node) {

	fmt.Println(n.GetBounds())

	if !n.IsLeaf() {
		for _, child := range n.Children {
			printBorder(castNode(child))
		}
	}

}
