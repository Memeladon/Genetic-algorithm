package frontend

import (
	"Genetic-algorithm/backend/genetic"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"image/color"
	"strconv"
)

type GraphWidget struct {
	widget.BaseWidget
	model     *genetic.GraphModel
	container *fyne.Container
	edges     []*canvas.Line
	vertices  []*canvas.Circle
}

func NewGraphWidget() *GraphWidget {
	gw := &GraphWidget{
		model:     genetic.NewGraphModel(6),
		container: container.NewWithoutLayout(),
		edges:     make([]*canvas.Line, 0),
		vertices:  make([]*canvas.Circle, 0),
	}
	gw.ExtendBaseWidget(gw)
	gw.ApplyModel()
	return gw
}

func (gw *GraphWidget) SetGraphModel(m *genetic.GraphModel) {
	gw.model = m
	gw.ApplyModel()
	gw.Refresh()
}

func (gw *GraphWidget) GetGraphModel() *genetic.GraphModel {
	return gw.model
}

func (gw *GraphWidget) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(gw.container)
}

func (gw *GraphWidget) ApplyModel() {
	gw.container.Objects = []fyne.CanvasObject{}
	gw.edges = gw.edges[:0]
	gw.vertices = gw.vertices[:0]

	// Рисуем рёбра
	for _, e := range gw.model.Edges {
		p1 := gw.model.Positions[e.U]
		p2 := gw.model.Positions[e.V]
		line := canvas.NewLine(color.NRGBA{R: 180, G: 180, B: 180, A: 255})
		line.StrokeWidth = 2
		line.Position1 = fyne.NewPos(float32(p1.X), float32(p1.Y))
		line.Position2 = fyne.NewPos(float32(p2.X), float32(p2.Y))
		gw.edges = append(gw.edges, line)
		gw.container.Add(line)
	}

	// Рисуем вершины поверх ребер
	for i, pos := range gw.model.Positions {
		circle := canvas.NewCircle(color.NRGBA{R: 50, G: 150, B: 250, A: 255})
		circle.StrokeWidth = 2
		circle.StrokeColor = color.White
		circle.Resize(fyne.NewSize(20, 20))
		circle.Move(fyne.NewPos(float32(pos.X-10), float32(pos.Y-10)))
		gw.vertices = append(gw.vertices, circle)
		gw.container.Add(circle)

		// Можно добавить текстовый слой поверх: номер вершины
		label := canvas.NewText(strconv.Itoa(i), color.White)
		label.TextSize = 12
		label.Move(fyne.NewPos(float32(pos.X-6), float32(pos.Y-8)))
		gw.container.Add(label)
	}
}

// Для подсветки рёбер в процессе алгоритма
func (gw *GraphWidget) updateEdgeColors(chrom genetic.Chromosome) {
	for idx, line := range gw.edges {
		if idx < len(chrom.Genes) && chrom.Genes[idx] {
			line.StrokeColor = color.NRGBA{R: 255, G: 80, B: 80, A: 255}
		} else {
			line.StrokeColor = color.NRGBA{R: 180, G: 180, B: 180, A: 255}
		}
		line.Refresh()
	}
}

// ForEachEdge применяет функцию к каждому ребру
func (gw *GraphWidget) ForEachEdge(f func(*canvas.Line)) {
	for _, line := range gw.edges {
		f(line)
	}
}

// Чтобы GUI вызывал подсветку:
func (mw *MainWindow) updateGraph(chrom genetic.Chromosome) {
	mw.GraphWidget.updateEdgeColors(chrom)
}
