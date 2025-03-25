package frontend

import (
	"Genetic-algorithm/backend"
	"Genetic-algorithm/backend/genetic"
	"errors"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"image/color"
	"log"
)

type MainWindow struct {
	App         fyne.App
	Window      fyne.Window
	GraphWidget *GraphWidget
	Controls    *ControlsPanel
	Solver      *backend.GASolver
}

func NewMainWindow(app fyne.App) *MainWindow {
	window := app.NewWindow("Genetic Algorithm Visualizer")

	mw := &MainWindow{
		Window:      window,
		GraphWidget: NewGraphWidget(),
		Controls:    NewControlsPanel(),
		Solver:      &backend.GASolver{UpdateChan: make(chan genetic.Chromosome)},
	}

	mw.setupCallbacks()
	mw.setupUI()
	return mw
}

func (mw *MainWindow) setupUI() {
	split := container.NewHSplit(
		mw.GraphWidget,
		mw.Controls.Render(),
	)
	split.Offset = 0.7
	mw.Window.SetContent(split)
	mw.Window.Resize(fyne.NewSize(1200, 800)) // Фиксируем размер окна

	log.Println("Окно инициализировано") // Добавляем лог
}

func (mw *MainWindow) setupCallbacks() {
	mw.Controls.OnStart = func() {
		if len(mw.GraphWidget.Edges) == 0 {
			dialog.ShowError(errors.New("граф не содержит рёбер"), mw.Window)
			return
		}
		mw.GraphWidget.IsEditable = false

		params := mw.Controls.GetParams()
		graph := convertGraph(mw.GraphWidget)

		mw.Solver.Start(graph, params)

		// Запускаем обработчик обновлений
		go func() {
			for chrom := range mw.Solver.UpdateChan {
				mw.updateGraph(chrom)
			}
		}()
	}

	mw.Controls.OnStop = func() {
		mw.Solver.Stop()
		mw.GraphWidget.IsEditable = true
	}
}

func (mw *MainWindow) updateGraph(chrom genetic.Chromosome) {
	// Сброс цвета всех рёбер
	mw.GraphWidget.ForEachEdge(func(e *EdgeWidget) {
		e.StrokeColor = color.Black
	})

	// Подсветка активных рёбер
	for i, gene := range chrom.Genes {
		if gene && i < len(mw.GraphWidget.Edges) {
			mw.GraphWidget.Edges[i].StrokeColor = color.NRGBA{R: 255, A: 255}
		}
	}
	mw.GraphWidget.Refresh()
}

func convertGraph(gw *GraphWidget) genetic.Graph {
	log.Println("Конвертация графа для алгоритма...")
	edges := make([]genetic.Edge, len(gw.Edges))

	for i, e := range gw.Edges {
		edges[i] = genetic.Edge{U: e.Start, V: e.End}
		log.Printf("Ребро %d: %d-%d", i, e.Start, e.End)
	}

	return genetic.Graph{
		NumVertices: len(gw.Vertices),
		Edges:       edges,
	}
}
