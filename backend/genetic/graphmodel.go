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

// GraphModel представляет модель неориентированного графа
// с явными позициями и списком рёбер.
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
		{300, 100},
		{200, 300},
		{400, 300},
		{300, 200},
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
	cube.Positions = make([]Point2D, 8)
	for i := 0; i < 8; i++ {
		x3 := (1 - 2*float64((i>>0)&1))
		y3 := (1 - 2*float64((i>>1)&1))
		z3 := (1 - 2*float64((i>>2)&1))
		cube.Positions[i] = Point2D{
			X: 300 + 100*x3 + 50*z3,
			Y: 250 + 100*y3 + 50*z3,
		}
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
	oct.Positions = []Point2D{
		{300, 100}, {150, 250}, {300, 400},
		{450, 250}, {300, 50}, {300, 450},
	}
	graphs["Octahedron (6)"] = oct

	//
	// 4) Icosahedron (12)
	//
	{
		phi := (1 + math.Sqrt(5)) / 2
		pts := [][3]float64{
			{0, 1, phi}, {0, -1, phi}, {0, 1, -phi}, {0, -1, -phi},
			{1, phi, 0}, {-1, phi, 0}, {1, -phi, 0}, {-1, -phi, 0},
			{phi, 0, 1}, {-phi, 0, 1}, {phi, 0, -1}, {-phi, 0, -1},
		}
		ico := NewGraphModel(12)
		// строим рёбра по расстоянию
		minD := math.Inf(1)
		for i := 0; i < 12; i++ {
			for j := i + 1; j < 12; j++ {
				d := dist2(pts[i], pts[j])
				if d > 1e-6 && d < minD {
					minD = d
				}
			}
		}
		thr := minD * 1.01
		for i := 0; i < 12; i++ {
			for j := i + 1; j < 12; j++ {
				if dist2(pts[i], pts[j]) < thr {
					ico.AddEdge(i, j)
				}
			}
		}
		ico.Positions = make([]Point2D, 12)
		for i, p := range pts {
			ico.Positions[i] = Point2D{
				X: 300 + p[0]*80,
				Y: 250 + p[1]*80,
			}
		}
		graphs["Icosahedron (12)"] = ico
	}

	//
	// 5) Dodecahedron (20)
	//
	{
		phi := (1 + math.Sqrt(5)) / 2
		inv := 1.0 / phi
		pts := make([][3]float64, 0, 20)
		// куб: ±1,±1,±1
		for sx := -1.0; sx <= 1; sx += 2 {
			for sy := -1.0; sy <= 1; sy += 2 {
				for sz := -1.0; sz <= 1; sz += 2 {
					pts = append(pts, [3]float64{sx, sy, sz})
				}
			}
		}
		// остальные
		extra := [][3]float64{
			{0, inv, phi}, {0, -inv, phi}, {0, inv, -phi}, {0, -inv, -phi},
			{inv, phi, 0}, {-inv, phi, 0}, {inv, -phi, 0}, {-inv, -phi, 0},
			{phi, 0, inv}, {-phi, 0, inv}, {phi, 0, -inv}, {-phi, 0, -inv},
		}
		pts = append(pts, extra...)
		dod := NewGraphModel(20)
		// рёбра по минимальному расстоянию
		minD := math.Inf(1)
		for i := range pts {
			for j := i + 1; j < len(pts); j++ {
				if d := dist2(pts[i], pts[j]); d > 1e-6 && d < minD {
					minD = d
				}
			}
		}
		thr := minD * 1.01
		for i := range pts {
			for j := i + 1; j < len(pts); j++ {
				if dist2(pts[i], pts[j]) < thr {
					dod.AddEdge(i, j)
				}
			}
		}
		dod.Positions = make([]Point2D, 20)
		for i, p := range pts {
			dod.Positions[i] = Point2D{
				X: 300 + p[0]*60,
				Y: 250 + p[1]*60,
			}
		}
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
