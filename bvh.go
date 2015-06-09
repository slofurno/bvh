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

func SquareRatio(a1 *AABB, a2 *AABB) float64 {

	minX := math.Min(a1.MinX, a2.MinX)
	maxX := math.Max(a1.MaxX, a2.MaxX)
	minY := math.Min(a1.MinY, a2.MinY)
	maxY := math.Max(a1.MaxY, a2.MaxY)

	width := maxX - minX
	height := maxY - minY

	p := 2*width + 2*height
	a := width * height

	return p / a

}

func (a1 *AABB) Merge(a2 *AABB) *AABB {
	return &AABB{
		MinX: math.Min(a1.MinX, a2.MinX),
		MaxX: math.Max(a1.MaxX, a2.MaxX),
		MinY: math.Min(a1.MinY, a2.MinY),
		MaxY: math.Max(a1.MaxY, a2.MaxY)}
}

func (a *AABB) Perimeter() float64 {
	dy := a.MaxY - a.MinY
	dx := a.MaxX - a.MinX

	return 2*dx + 2*dy
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
	Bounds *AABB
	Left   int
	Right  int
	Parent int
	Value  Bounded
	Next   int
}

func (n *Node) IsLeaf() bool {
	if n.Left >= 0 {
		return false
	}
	return true
}

func (n *Node) GetBounds() *AABB {
	return n.Bounds
}

var scannode int = 0

func Scan(n int, b Bounded) {

	if n == -1 {
		return
	}

	if !nodes[n].GetBounds().Intersects(b.GetBounds()) {
		return
	}

	if nodes[n].Value != nil && nodes[n].Value != b && nodes[n].Value.GetBounds().Intersects(b.GetBounds()) {
		//*results = append(*results, n.Value)
		pairs[pairindex] = nodes[n].Value
		pairindex++
		return
	}

	Scan(nodes[n].Left, b)
	Scan(nodes[n].Right, b)

}

func Insert(index int) {

	if rootindex == -1 {
		rootindex = index
		nodes[index].Parent = -1
		return
	}

	bounds := nodes[index].GetBounds()
	current := rootindex

	for nodes[current].IsLeaf() == false {

		left := nodes[current].Left
		right := nodes[current].Right
		/*
			tryleft := nodes[left].Bounds.Merge(bounds)
			tryright := nodes[right].Bounds.Merge(bounds)

			leftratio := tryleft.Perimeter() / tryleft.Area()
			rightratio := tryright.Perimeter() / tryright.Area()
		*/

		leftratio := SquareRatio(nodes[left].Bounds, bounds)
		rightratio := SquareRatio(nodes[right].Bounds, bounds)

		if leftratio > rightratio {
			current = left
		} else {
			current = right
		}
	}

	parent := nodes[current].Parent

	newnode := getNode()
	nodes[newnode].Parent = parent
	nodes[newnode].Left = current
	nodes[newnode].Right = index
	nodes[current].Parent = newnode
	nodes[index].Parent = newnode

	if parent != -1 {

		if nodes[parent].Left == current {
			nodes[parent].Left = newnode
		} else {
			nodes[parent].Right = newnode
		}

	} else {
		rootindex = newnode
	}

	current = nodes[index].Parent

	for current != -1 {

		left := nodes[current].Left
		right := nodes[current].Right

		nodes[current].Bounds = nodes[left].Bounds.Merge(nodes[right].Bounds)
		current = nodes[current].Parent
	}

}

var pairs = make([]Bounded, 100000)
var pairindex int = 0
var nodes = make([]Node, 100000)
var nodecount int = 0
var nextnode int = 0
var rootindex = -1

func getNode() int {
	nextindex := nextnode
	nextnode = nodes[nextnode].Next
	nodes[nextindex].Parent = -1
	nodes[nextindex].Left = -1
	nodes[nextindex].Right = -1

	return nextindex
}

func putNode(index int) {
	nodes[index].Next = nextnode
	nextnode = index
}

func initNodes() {
	nextnode = 0
	rootindex = -1
	for i := 0; i < len(nodes); i++ {
		nodes[i].Next = i + 1
	}

}

func main() {

	runtime.GOMAXPROCS(runtime.NumCPU())
	p := profile.Start(profile.CPUProfile, profile.ProfilePath("."), profile.NoShutdownHook)
	defer p.Stop()

	const maxsize int = 1000

	squares := []*Square{}
	width := float64(2)
	height := float64(2)

	for i := 0; i < 9000; i++ {

		x1 := float64(rand.Intn(maxsize))
		y1 := float64(rand.Intn(maxsize))
		square := NewSquare(x1, y1, x1+width, y1+height)
		squares = append(squares, square)
	}

	for i := 0; i < 10; i++ {
		initNodes()

		start := time.Now()
		collisions := 0
		scannode = 0
		pairindex = 0

		for _, s := range squares {
			node := getNode()
			nodes[node].Value = s
			nodes[node].Bounds = s.GetBounds()
			Insert(node)
		}

		insertiondone := time.Since(start)

		for _, s := range squares {
			Scan(rootindex, s)
		}

		collisions = pairindex

		elapsed := time.Since(start)
		fmt.Println("collisions:", collisions, "elapsed time:", elapsed.String(), "insert time:", insertiondone)

	}

	for i := 0; i < 1; i++ {
		start := time.Now()
		collisions := 0
		scannode = 0

		for j := 1; j < len(squares); j++ {
			for k := 0; k < j; k++ {
				scannode++
				if squares[j] != squares[k] && squares[j].GetBounds().Intersects(squares[k].GetBounds()) {
					collisions++
				}
			}
		}

		elapsed := time.Since(start)
		fmt.Println("collisions:", collisions, "elapsed time:", elapsed.String(), "nodes scanned", scannode)
	}

}
