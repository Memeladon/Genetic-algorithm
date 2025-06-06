package genetic

import (
	"fmt"
	"math"
	"math/rand"
)

// Point2D задаёт координаты в 2D-пространстве для рендеринга графа.
type Point2D struct {
	X, Y float64
}

// GraphModel представляет модель неориентированного графа с явными позициями и списком рёбер.
type GraphModel struct {
	NumVertices int
	Edges       []Edge
	Positions   []Point2D // len == NumVertices
}

// NewGraphModel создаёт пустую модель графа с n вершинами.
func NewGraphModel(n int) *GraphModel {
	return &GraphModel{
		NumVertices: n,
		Edges:       make([]Edge, 0),
		Positions:   make([]Point2D, n),
	}
}

func (gm *GraphModel) AddEdge(u, v int) {
	if u == v || u < 0 || v < 0 || u >= gm.NumVertices || v >= gm.NumVertices {
		return
	}
	// проверяем дубли
	for _, e := range gm.Edges {
		if (e.U == u && e.V == v) || (e.U == v && e.V == u) {
			return
		}
	}
	gm.Edges = append(gm.Edges, Edge{U: u, V: v})
}

// ToGraph конвертирует модель в Graph для запуска алгоритма.
func (gm *GraphModel) ToGraph() Graph {
	return Graph{NumVertices: gm.NumVertices, Edges: gm.Edges}
}

// PredefinedGraphs возвращает карту всех шаблонных графов с корректными Positions.
func PredefinedGraphs() map[string]*GraphModel {
	graphs := make(map[string]*GraphModel)

	//
	// 1) Tetrahedron (4)
	//
	k4 := NewGraphModel(4)
	for i := 0; i < 4; i++ {
		for j := i + 1; j < 4; j++ {
			k4.AddEdge(i, j)
		}
	}
	// Позиции вершин правильного тетраэдра (вид сверху на треугольную пирамиду)
	k4.Positions = []Point2D{
		{300, 100}, // top
		{200, 300}, // left
		{400, 300}, // right
		{300, 220}, // center
	}
	graphs["Tetrahedron (4)"] = k4

	//
	// 2) Cube (8) — 3D-гиперкуб в изометрии
	//
	cube := NewGraphModel(8)
	for i := 0; i < 8; i++ {
		for b := 0; b < 3; b++ {
			j := i ^ (1 << b)
			if j > i {
				cube.AddEdge(i, j)
			}
		}
	}
	cube.Positions = []Point2D{
		{200, 100}, // top-left
		{400, 100}, // top-right
		{200, 300}, // bottom-left
		{400, 300}, // bottom-right
		{250, 150}, // inner top-left
		{350, 150}, // inner top-right
		{250, 250}, // inner bottom-left
		{350, 250}, // inner bottom-right
	}
	graphs["Cube (8)"] = cube

	//
	// 3) Octahedron (6)
	//
	oct := NewGraphModel(6)
	opp := map[int]int{0: 1, 1: 0, 2: 3, 3: 2, 4: 5, 5: 4}
	for i := 0; i < 6; i++ {
		for j := i + 1; j < 6; j++ {
			if opp[i] != j {
				oct.AddEdge(i, j)
			}
		}
	}
	// Two triangles, one inverted over the other
	oct.Positions = []Point2D{
		{300, 100}, // top
		{200, 250}, // left
		{400, 250}, // right
		{250, 350}, // bottom left
		{350, 350}, // bottom right
		{300, 220}, // center
	}
	graphs["Octahedron (6)"] = oct

	//
	// 4) Icosahedron (12) — Schlegel diagram, 1:1 with reference image
	//
	{
		ico := NewGraphModel(12)
		// Node positions: top, outer pentagon, inner pentagon, center
		cx, cy := 300.0, 250.0
		R_outer := 170.0
		R_inner := 90.0
		// Top node
		ico.Positions[0] = Point2D{cx, cy - 210}
		// Outer pentagon (1-5)
		angles := []float64{-math.Pi / 2, -math.Pi/2 + 2*math.Pi/5, -math.Pi/2 + 4*math.Pi/5, -math.Pi/2 + 6*math.Pi/5, -math.Pi/2 + 8*math.Pi/5}
		for i := 0; i < 5; i++ {
			ico.Positions[1+i] = Point2D{
				X: cx + R_outer*math.Cos(angles[i]),
				Y: cy + R_outer*math.Sin(angles[i]),
			}
		}
		// Inner pentagon (6-10)
		for i := 0; i < 5; i++ {
			ico.Positions[6+i] = Point2D{
				X: cx + R_inner*math.Cos(angles[i]),
				Y: cy + R_inner*math.Sin(angles[i]),
			}
		}
		// Center node (11)
		ico.Positions[11] = Point2D{cx, cy}
		// Edges as in the reference image (0-based)
		// Top node to all outer pentagon
		ico.AddEdge(0, 1)
		ico.AddEdge(0, 2)
		ico.AddEdge(0, 3)
		ico.AddEdge(0, 4)
		ico.AddEdge(0, 5)
		// Outer pentagon
		ico.AddEdge(1, 2)
		ico.AddEdge(2, 3)
		ico.AddEdge(3, 4)
		ico.AddEdge(4, 5)
		ico.AddEdge(5, 1)
		// Outer to inner pentagon spokes
		ico.AddEdge(1, 6)
		ico.AddEdge(2, 7)
		ico.AddEdge(3, 8)
		ico.AddEdge(4, 9)
		ico.AddEdge(5, 10)
		// Inner pentagon
		ico.AddEdge(6, 7)
		ico.AddEdge(7, 8)
		ico.AddEdge(8, 9)
		ico.AddEdge(9, 10)
		ico.AddEdge(10, 6)
		// Inner pentagon to center
		ico.AddEdge(6, 11)
		ico.AddEdge(7, 11)
		ico.AddEdge(8, 11)
		ico.AddEdge(9, 11)
		ico.AddEdge(10, 11)
		// Star/diagonal connections
		ico.AddEdge(1, 8)
		ico.AddEdge(2, 9)
		ico.AddEdge(3, 10)
		ico.AddEdge(4, 6)
		ico.AddEdge(5, 7)
		graphs["Icosahedron (12)"] = ico
	}

	//
	// 5) Dodecahedron (20) — Schlegel diagram, 1:1 with reference image
	//
	{
		dod := NewGraphModel(20)
		// Node positions: 4 concentric pentagons, but order matches the image
		cx, cy := 300.0, 250.0
		radii := []float64{170, 130, 90, 50}
		angles := []float64{-math.Pi / 2, -math.Pi/2 + 2*math.Pi/5, -math.Pi/2 + 4*math.Pi/5, -math.Pi/2 + 6*math.Pi/5, -math.Pi/2 + 8*math.Pi/5}
		// Outer pentagon (1-5)
		for i := 0; i < 5; i++ {
			dod.Positions[i] = Point2D{
				X: cx + radii[0]*math.Cos(angles[i]),
				Y: cy + radii[0]*math.Sin(angles[i]),
			}
		}
		// 2nd pentagon (6-10)
		for i := 0; i < 5; i++ {
			dod.Positions[5+i] = Point2D{
				X: cx + radii[1]*math.Cos(angles[i]),
				Y: cy + radii[1]*math.Sin(angles[i]),
			}
		}
		// 3rd pentagon (11-15)
		for i := 0; i < 5; i++ {
			dod.Positions[10+i] = Point2D{
				X: cx + radii[2]*math.Cos(angles[i]),
				Y: cy + radii[2]*math.Sin(angles[i]),
			}
		}
		// innermost pentagon (16-20)
		for i := 0; i < 5; i++ {
			dod.Positions[15+i] = Point2D{
				X: cx + radii[3]*math.Cos(angles[i]),
				Y: cy + radii[3]*math.Sin(angles[i]),
			}
		}
		// Edges as in the image (0-based)
		// Outer pentagon
		dod.AddEdge(0, 1)
		dod.AddEdge(1, 2)
		dod.AddEdge(2, 3)
		dod.AddEdge(3, 4)
		dod.AddEdge(4, 0)
		// Spokes
		dod.AddEdge(0, 5)
		dod.AddEdge(1, 6)
		dod.AddEdge(2, 7)
		dod.AddEdge(3, 8)
		dod.AddEdge(4, 9)
		// 2nd pentagon
		dod.AddEdge(5, 6)
		dod.AddEdge(6, 7)
		dod.AddEdge(7, 8)
		dod.AddEdge(8, 9)
		dod.AddEdge(9, 5)
		// Spokes
		dod.AddEdge(5, 10)
		dod.AddEdge(6, 11)
		dod.AddEdge(7, 12)
		dod.AddEdge(8, 13)
		dod.AddEdge(9, 14)
		// 3rd pentagon
		dod.AddEdge(10, 11)
		dod.AddEdge(11, 12)
		dod.AddEdge(12, 13)
		dod.AddEdge(13, 14)
		dod.AddEdge(14, 10)
		// Spokes
		dod.AddEdge(10, 15)
		dod.AddEdge(11, 16)
		dod.AddEdge(12, 17)
		dod.AddEdge(13, 18)
		dod.AddEdge(14, 19)
		// innermost pentagon
		dod.AddEdge(15, 16)
		dod.AddEdge(16, 17)
		dod.AddEdge(17, 18)
		dod.AddEdge(18, 19)
		dod.AddEdge(19, 15)
		// Center cross connections (as in the image)
		dod.AddEdge(15, 17)
		dod.AddEdge(17, 19)
		dod.AddEdge(19, 16)
		dod.AddEdge(16, 18)
		dod.AddEdge(18, 15)
		graphs["Dodecahedron (20)"] = dod
	}

	//
	// === Большие графы ===
	//

	// 1) Grid
	for _, size := range []int{25, 50, 100} {
		var rows, cols int
		switch size {
		case 25:
			rows, cols = 5, 5
		case 50:
			rows, cols = 5, 10
		case 100:
			rows, cols = 10, 10
		}
		g := NewGraphModel(size)
		g.Positions = make([]Point2D, size)
		for r := 0; r < rows; r++ {
			for c := 0; c < cols; c++ {
				idx := r*cols + c
				// сетка
				if c+1 < cols {
					g.AddEdge(idx, idx+1)
				}
				if r+1 < rows {
					g.AddEdge(idx, idx+cols)
				}
				// позиция
				g.Positions[idx] = Point2D{
					X: 50 + float64(c)*50,
					Y: 50 + float64(r)*50,
				}
			}
		}
		graphs[fmt.Sprintf("Grid %dx%d (%d)", rows, cols, size)] = g
	}

	// 2) Cycle
	for _, n := range []int{25, 50, 100} {
		c := NewGraphModel(n)
		c.Positions = make([]Point2D, n)
		R := 200.0
		cx, cy := 300.0, 250.0
		for i := 0; i < n; i++ {
			c.AddEdge(i, (i+1)%n)
			theta := 2 * math.Pi * float64(i) / float64(n)
			c.Positions[i] = Point2D{
				X: cx + R*math.Cos(theta),
				Y: cy + R*math.Sin(theta),
			}
		}
		graphs[fmt.Sprintf("Cycle %d", n)] = c
	}

	// 3) Random
	rand.Seed(42)
	for _, n := range []int{25, 50, 100} {
		r := NewGraphModel(n)
		r.Positions = make([]Point2D, n)
		p := 0.1
		for i := 0; i < n; i++ {
			for j := i + 1; j < n; j++ {
				if rand.Float64() < p {
					r.AddEdge(i, j)
				}
			}
			// случайное расположение
			r.Positions[i] = Point2D{
				X: 50 + rand.Float64()*500,
				Y: 50 + rand.Float64()*400,
			}
		}
		graphs[fmt.Sprintf("Random %d", n)] = r
	}

	// === НОВЫЙ: очень большой граф ===

	// 3) Большой сетевой граф 25×40 = 1000
	{
		rows, cols := 25, 40
		size := rows * cols
		g := NewGraphModel(size)
		g.Positions = make([]Point2D, size)
		const dx, dy = 30.0, 30.0
		for r := 0; r < rows; r++ {
			for c := 0; c < cols; c++ {
				idx := r*cols + c
				if c+1 < cols {
					g.AddEdge(idx, idx+1)
				}
				if r+1 < rows {
					g.AddEdge(idx, idx+cols)
				}
				g.Positions[idx] = Point2D{
					X: 50 + float64(c)*dx,
					Y: 50 + float64(r)*dy,
				}
			}
		}
		graphs[fmt.Sprintf("Great grid %dx%d (%d)", rows, cols, size)] = g
	}

	// 4) Большой случайный граф 1000
	{
		n := 1000
		r := NewGraphModel(n)
		r.Positions = make([]Point2D, n)
		p := 0.01 // плотность связей
		for i := 0; i < n; i++ {
			r.Positions[i] = Point2D{
				X: 50 + rand.Float64()*700,
				Y: 50 + rand.Float64()*500,
			}
			for j := i + 1; j < n; j++ {
				if rand.Float64() < p {
					r.AddEdge(i, j)
				}
			}
		}
		graphs[fmt.Sprintf("Great random (%d edges)", n)] = r
	}

	return graphs
}

// dist2 возвращает квадрат расстояния между двумя 3D-точками.
func dist2(a, b [3]float64) float64 {
	dx := a[0] - b[0]
	dy := a[1] - b[1]
	dz := a[2] - b[2]
	return dx*dx + dy*dy + dz*dz
}
