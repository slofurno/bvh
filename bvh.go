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

var mergecalls int = 0

func (a *AABB) Clone() *AABB {
	return &AABB{
		MinX: a.MinX,
		MaxX: a.MaxX,
		MinY: a.MinY,
		MaxY: a.MaxY}
}

func (a1 *AABB) Add(a2 *AABB) {
	a1.MinX = math.Min(a1.MinX, a2.MinX)
	a1.MaxX = math.Max(a1.MaxX, a2.MaxX)
	a1.MinY = math.Min(a1.MinY, a2.MinY)
	a1.MaxY = math.Max(a1.MaxY, a2.MaxY)
}

func (a1 *AABB) Merge(a2 *AABB) *AABB {
	mergecalls++
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
	Value  Bounded
}

func (n *Node) GetBounds() *AABB {
	return n.Bounds
}

var scannode int = 0

func Scan(n *Node, b Bounded) {

	scannode++

	if !n.GetBounds().Intersects(b.GetBounds()) {
		return
	}

	if n.Value != nil && n.Value != b && n.Value.GetBounds().Intersects(b.GetBounds()) {
		//*results = append(*results, n.Value)
		pairs[pairindex] = n.Value
		pairindex++
		return
	}

	if n.Right > 0 {
		Scan(&nodes[n.Left], b)
		Scan(&nodes[n.Right], b)
	} else if n.Left > 0 {
		Scan(&nodes[n.Left], b)

	}

}

var nodecount int = 0

func (n *Node) Insert(b Bounded) {
	n.Bounds.Add(b.GetBounds())

	//left + right both have value, pick optimal one to continue down
	if n.Left > 0 && n.Right > 0 {

		tryleft := nodes[n.Left].Bounds.Merge(b.GetBounds())
		tryright := nodes[n.Right].Bounds.Merge(b.GetBounds())

		leftratio := tryleft.Perimeter() / tryleft.Area()
		rightratio := tryright.Perimeter() / tryright.Area()

		if leftratio > rightratio {

			nodes[n.Left].Insert(b)
		} else {

			nodes[n.Right].Insert(b)
		}
		//right must be empty, insert new node there
	} else if n.Left > 0 {
		n.Right = nextnode
		nextnode++

		nodes[n.Right].Value = b
		nodes[n.Right].Bounds = b.GetBounds().Clone()

		//&Node{Value: b, Bounds: b.GetBounds()}
	} else if n.Value != nil {
		n.Left = nextnode
		nextnode++
		nodes[n.Left].Value = b
		nodes[n.Left].Bounds = b.GetBounds().Clone()

		n.Right = nextnode
		nextnode++
		nodes[n.Right].Value = n.Value
		nodes[n.Right].Bounds = n.Bounds.Clone()

		//n.Left = &Node{Value: b, Bounds: b.GetBounds()}
		//n.Right = &Node{Value: n.Value, Bounds: n.Bounds}
		nodecount += 2
		n.Value = nil
	} else {
		n.Value = b
	}
	/*
		n.Bounds = n.Bounds.Merge(b.GetBounds())
		if n.IsLeaf() {
			n.Left = &Node{Value: b, Bounds: b.GetBounds()}
			if n.Value != nil {
				n.Right = &Node{Value: n.Value, Bounds: n.Bounds}
				n.Value = nil
			}
			//n.Bounds = n.Left.Bounds.Merge(n.Right.Bounds)
			return
		}

		if n.Right == nil {
			n.Right = &Node{Value: b, Bounds: b.GetBounds()}

		} else {
			tryleft := n.Left.Bounds.Merge(b.GetBounds()).Area()
			tryright := n.Right.Bounds.Merge(b.GetBounds()).Area()

			if tryleft < tryright {
				n.Left.Insert(b)
			} else {
				n.Right.Insert(b)
			}
		}
	*/
}

func castNode(n Bounded) *Node {
	switch t := n.(type) {
	case *Node:
		return t
	default:
		panic("cant cast to node")
	}

}

var pairs = make([]Bounded, 20000)
var pairindex int = 0
var nodes = make([]Node, 100000)
var nextnode int = 0

func Init() {
	for i := 0; i < len(nodes); i++ {
		nodes[i] = Node{}
	}
}

func Clear() {
	for i := 0; i < nextnode; i++ {
		nodes[i].Value = nil
		nodes[i].Left = 0
		nodes[i].Right = 0

	}
	nextnode = 1
}

func main() {

	runtime.GOMAXPROCS(runtime.NumCPU())

	p := profile.Start(profile.CPUProfile, profile.ProfilePath("."), profile.NoShutdownHook)
	defer p.Stop()

	const maxsize int = 10000

	squares := []*Square{}
	width := float64(10)
	height := float64(10)

	for i := 0; i < 15000; i++ {

		x1 := float64(rand.Intn(maxsize))
		y1 := float64(rand.Intn(maxsize))
		square := NewSquare(x1, y1, x1+width, y1+height)
		//square := &Square{X1: x1, Y1: y1, X2: x1 + width, Y2: y1 + height}
		//root.Insert(square)
		squares = append(squares, square)
	}

	for i := 0; i < 10; i++ {
		start := time.Now()
		Clear()
		nodecount = 0
		collisions := 0
		scannode = 0
		pairindex = 0
		mergecalls = 0

		root := &Node{Bounds: &AABB{0, 0, float64(maxsize) + width, float64(maxsize) + height}}

		for _, s := range squares {
			root.Insert(s)
		}

		insertiondone := time.Since(start)

		for _, s := range squares {
			Scan(root, s)
		}

		collisions = pairindex

		elapsed := time.Since(start)
		fmt.Println("collisions:", collisions, "elapsed time:", elapsed.String(), "insert time:", insertiondone, "node count", nodecount, "nodes scanned", scannode, "merges:", mergecalls)

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
