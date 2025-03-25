package frontend

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"image/color"
	"log"
	"math"
	"strconv"
)

type GraphWidget struct {
	widget.BaseWidget
	Vertices   map[int]*VertexWidget
	Edges      []*EdgeWidget
	IsEditable bool
	selected   []int
}

type VertexWidget struct {
	*canvas.Circle
	Text *canvas.Text
	ID   int
}

type EdgeWidget struct {
	*canvas.Line
	Start, End int
}

func NewGraphWidget() *GraphWidget {
	gw := &GraphWidget{
		Vertices:   make(map[int]*VertexWidget),
		Edges:      make([]*EdgeWidget, 0),
		IsEditable: true,
	}
	gw.ExtendBaseWidget(gw)

	// Создаем 6 вершин по умолчанию
	positions := []fyne.Position{
		{50, 100},  // 0
		{150, 50},  // 1
		{150, 150}, // 2
		{250, 50},  // 3
		{250, 150}, // 4
		{350, 100}, // 5
	}

	for i := 0; i < 6; i++ {
		circle := canvas.NewCircle(color.NRGBA{R: 255, A: 255})
		circle.Resize(fyne.NewSize(30, 30))
		circle.Move(positions[i])

		text := canvas.NewText(strconv.Itoa(i), color.White)
		text.TextSize = 14
		text.Move(positions[i].AddXY(10, 5))

		gw.Vertices[i] = &VertexWidget{
			Circle: circle,
			Text:   text,
			ID:     i,
		}
		log.Printf("Создана вершина %d в позиции %v", i, positions[i])
	}

	// Добавляем стандартные рёбра
	edges := []struct{ start, end int }{
		{0, 1}, {0, 2}, {1, 2},
		{1, 3}, {2, 4}, {3, 4},
		{3, 5}, {4, 5},
	}

	for _, e := range edges {
		gw.addEdge(e.start, e.end)
		log.Printf("Добавлено ребро %d-%d", e.start, e.end)
	}

	return gw
}

func (gw *GraphWidget) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(gw.render())
}

func (gw *GraphWidget) render() fyne.CanvasObject {
	objects := make([]fyne.CanvasObject, 0)
	for _, edge := range gw.Edges {
		objects = append(objects, edge.Line)
	}
	for _, vertex := range gw.Vertices {
		objects = append(objects, vertex.Circle, vertex.Text)
	}
	return container.NewWithoutLayout(objects...)
}

func (gw *GraphWidget) Tapped(ev *fyne.PointEvent) {
	if !gw.IsEditable {
		return
	}

	for _, v := range gw.Vertices {
		// Используем Position() из canvas.Circle
		pos := v.Position().Add(fyne.NewPos(v.Size().Width/2, v.Size().Height/2))
		if distance(pos, ev.Position) < 20 {
			gw.handleVertexTap(v.ID)
			return
		}
	}

	id := len(gw.Vertices)
	vertex := &VertexWidget{
		Circle: canvas.NewCircle(color.NRGBA{R: 0, G: 0, B: 255, A: 255}),
		ID:     id,
	}
	vertex.Resize(fyne.NewSize(20, 20))
	vertex.Move(ev.Position.SubtractXY(10, 10)) // Устанавливаем позицию через Move()
	gw.Vertices[id] = vertex
	gw.Refresh()
}

// Dragged Реализация интерфейса Draggable
func (gw *GraphWidget) Dragged(ev *fyne.DragEvent) {
	if !gw.IsEditable {
		return
	}

	for _, v := range gw.Vertices {
		pos := v.Position().Add(fyne.NewPos(10, 10)) // Получаем позицию через Position()
		if distance(pos, ev.Position) < 20 {
			newPos := ev.PointEvent.Position.SubtractXY(10, 10)
			v.Move(newPos) // Обновляем позицию через Move()
			gw.updateEdges(v.ID)
			gw.Refresh()
			break
		}
	}
}

func (gw *GraphWidget) handleVertexTap(id int) {
	if len(gw.selected) == 0 {
		gw.selected = append(gw.selected, id)
		gw.Vertices[id].FillColor = color.NRGBA{R: 255, G: 0, B: 0, A: 255}
	} else {
		start := gw.selected[0]
		end := id
		gw.addEdge(start, end)
		gw.Vertices[start].FillColor = color.NRGBA{R: 0, G: 0, B: 255, A: 255}
		gw.selected = gw.selected[:0]
	}
	gw.Refresh()
}

func (gw *GraphWidget) addEdge(start, end int) {
	line := canvas.NewLine(color.Black)
	line.StrokeWidth = 2
	line.Position1 = gw.Vertices[start].Position().AddXY(15, 15)
	line.Position2 = gw.Vertices[end].Position().AddXY(15, 15)

	gw.Edges = append(gw.Edges, &EdgeWidget{
		Line:  line,
		Start: start,
		End:   end,
	})
}

func distance(a, b fyne.Position) float32 {
	dx := a.X - b.X
	dy := a.Y - b.Y
	return float32(math.Sqrt(float64(dx*dx + dy*dy)))
}

func (gw *GraphWidget) updateEdges(vertexID int) {
	for _, edge := range gw.Edges {
		if edge.Start == vertexID {
			edge.Position1 = gw.Vertices[vertexID].Position().AddXY(10, 10)
		}
		if edge.End == vertexID {
			edge.Position2 = gw.Vertices[vertexID].Position().AddXY(10, 10)
		}
	}
}

func (gw *GraphWidget) ForEachEdge(fn func(*EdgeWidget)) {
	for _, edge := range gw.Edges {
		fn(edge)
	}
}
