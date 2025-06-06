package frontend

import (
	"Genetic-algorithm/backend"
	"Genetic-algorithm/backend/genetic"
	"errors"
	"log"
	"sort"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

type MainWindow struct {
	App          fyne.App
	Window       fyne.Window
	GraphWidget  *GraphWidget
	Controls     *ControlsPanel
	Solver       *backend.GASolver
	PresetSelect *widget.Select // Добавляем сохранение селектора
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
	// Sort names for stable order
	sort.Strings(names)
	presetSelect := widget.NewSelect(names, func(name string) {
		gm := predefs[name]
		graphWidget.SetGraphModel(gm)
	})
	presetSelect.PlaceHolder = "Select graph..."

	// Select the first graph by default
	if len(names) > 0 {
		presetSelect.SetSelected(names[0])
		graphWidget.SetGraphModel(predefs[names[0]])
	}

	// Сохраняем presetSelect в структуре
	mw := &MainWindow{
		Window:       window,
		GraphWidget:  graphWidget,
		Controls:     controls,
		Solver:       &backend.GASolver{UpdateChan: make(chan genetic.Chromosome)},
		PresetSelect: presetSelect, // Сохраняем селектор
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
		mw.Controls.StopBtn.Enable()

		// Конвертация модели в genetic.Graph
		gm := mw.GraphWidget.GetGraphModel()
		if len(gm.Edges) == 0 {
			dialog.ShowError(errors.New("граф не содержит рёбер"), mw.Window)
			mw.Controls.StartBtn.Enable()
			mw.Controls.StopBtn.Disable()
			return
		}

		params := mw.Controls.GetParams()
		graph := genetic.Graph{NumVertices: gm.NumVertices, Edges: gm.Edges}

		// Получаем имя текущего графа
		graphName := "Custom"
		if mw.PresetSelect.Selected != "" {
			graphName = mw.PresetSelect.Selected
		}

		mw.Solver.UpdateChan = make(chan genetic.Chromosome)
		mw.Solver.Done = make(chan struct{})

		// Передаем три аргумента
		mw.Solver.Start(graph, params, graphName)

		// Обработка обновлений
		go func() {
			for chrom := range mw.Solver.UpdateChan {
				mw.updateGraph(chrom)
			}
			<-mw.Solver.Done

			// После завершения: подсветить только лучшее паросочетание
			if len(mw.Solver.Results) > 0 {
				// Найти результат с максимальным BestFitness
				bestRes := mw.Solver.Results[0]
				for _, r := range mw.Solver.Results {
					if r.BestFitness > bestRes.BestFitness {
						bestRes = r
					}
				}
				bestIndices := make(map[int]struct{})
				for _, idx := range bestRes.BestMatchingEdges {
					bestIndices[idx] = struct{}{}
				}
				mw.GraphWidget.updateEdgeColorsBestOnly(bestIndices)
			}
			mw.Controls.StartBtn.Enable()
			mw.Controls.StopBtn.Disable()
			mw.Window.Canvas().Refresh(mw.Controls.StartBtn)
			mw.Window.Canvas().Refresh(mw.Controls.StopBtn)
		}()
	}

	mw.Controls.OnStop = func() {
		mw.Solver.Stop()
		mw.Controls.StartBtn.Enable()
		mw.Controls.StopBtn.Disable()
		mw.Window.Canvas().Refresh(mw.Controls.StartBtn)
		mw.Window.Canvas().Refresh(mw.Controls.StopBtn)
	}

	mw.Controls.OnPlot = func() {
		if len(mw.Solver.Results) == 0 {
			dialog.ShowError(errors.New("нет данных для построения графиков"), mw.Window)
			return
		}

		go func() {
			if err := mw.Solver.PlotResults(); err != nil {
				dialog.ShowError(err, mw.Window)
			} else {
				dialog.ShowInformation(
					"Графики построены",
					"Графики сохранены в папку plots_<timestamp>",
					mw.Window,
				)
			}
		}()
	}
}
