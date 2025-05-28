package frontend

import (
	"Genetic-algorithm/backend"
	"Genetic-algorithm/backend/genetic"
	"errors"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
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

	controls := NewControlsPanel()
	graphWidget := NewGraphWidget()

	// Селектор предопределённых графов
	predefs := genetic.PredefinedGraphs()
	names := make([]string, 0, len(predefs))
	for k := range predefs {
		names = append(names, k)
	}
	presetSelect := widget.NewSelect(names, func(name string) {
		gm := predefs[name]
		graphWidget.SetGraphModel(gm)
	})
	presetSelect.PlaceHolder = "Select graph..."

	mw := &MainWindow{
		Window:      window,
		GraphWidget: graphWidget,
		Controls:    controls,
		Solver:      &backend.GASolver{UpdateChan: make(chan genetic.Chromosome)},
	}

	// Левая панель: селектор + граф
	left := container.NewBorder(presetSelect, nil, nil, nil, container.NewMax(graphWidget))

	// Правая прокручиваемая панель
	right := container.NewVScroll(controls.Render())
	right.SetMinSize(fyne.NewSize(300, 0))

	split := container.NewHSplit(left, right)
	split.SetOffset(0.75)

	window.SetContent(split)
	window.Resize(fyne.NewSize(1200, 800))

	mw.setupCallbacks()
	log.Println("MainWindow initialized")
	return mw
}

func (mw *MainWindow) setupCallbacks() {
	mw.Controls.OnStart = func() {
		mw.Controls.StartBtn.Disable()

		// Конвертация модели в genetic.Graph
		gm := mw.GraphWidget.GetGraphModel()
		if len(gm.Edges) == 0 {
			dialog.ShowError(errors.New("граф не содержит рёбер"), mw.Window)
			mw.Controls.StartBtn.Enable()
			return
		}
		//mw.GraphWidget.IsEditable = false

		params := mw.Controls.GetParams()
		graph := genetic.Graph{NumVertices: gm.NumVertices, Edges: gm.Edges}

		mw.Solver.Start(graph, params)

		// Обработка обновлений
		go func() {
			for chrom := range mw.Solver.UpdateChan {
				mw.updateGraph(chrom)
			}
			mw.Controls.StartBtn.Enable()
		}()
	}

	mw.Controls.OnStop = func() {
		mw.Solver.Stop()
		mw.Controls.StartBtn.Enable()
		//mw.GraphWidget.IsEditable = true
	}
}
