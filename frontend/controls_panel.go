package frontend

import (
	"Genetic-algorithm/backend"
	"Genetic-algorithm/backend/genetic"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"strconv"
)

type ControlsPanel struct {
	StartBtn          *widget.Button
	StopBtn           *widget.Button
	PopulationSize    *widget.Entry
	Generations       *widget.Entry
	MutationRate      *widget.Entry
	CrossoverRate     *widget.Entry
	NumIslands        *widget.Entry
	MigrationInterval *widget.Entry
	EvolutionModel    *widget.RadioGroup
	OnStart           func()
	OnStop            func()
}

func NewControlsPanel() *ControlsPanel {
	cp := &ControlsPanel{
		PopulationSize:    widget.NewEntry(),
		Generations:       widget.NewEntry(),
		MutationRate:      widget.NewEntry(),
		CrossoverRate:     widget.NewEntry(),
		NumIslands:        widget.NewEntry(),
		MigrationInterval: widget.NewEntry(),
		EvolutionModel:    widget.NewRadioGroup([]string{"Classic", "Island", "Combined"}, nil),
	}

	cp.EvolutionModel.SetSelected("Classic")
	cp.setDefaults()

	cp.StartBtn = widget.NewButton("Старт", func() {
		if cp.OnStart != nil {
			cp.OnStart()
		}
	})
	cp.StopBtn = widget.NewButton("Стоп", func() {
		if cp.OnStop != nil {
			cp.OnStop()
		}
	})

	return cp
}

func (cp *ControlsPanel) setDefaults() {
	cp.PopulationSize.SetText("100")
	cp.Generations.SetText("100")
	cp.MutationRate.SetText("0.05")
	cp.CrossoverRate.SetText("0.8")
	cp.NumIslands.SetText("4")
	cp.MigrationInterval.SetText("10")
}

func (cp *ControlsPanel) GetParams() backend.Params {
	popSize, _ := strconv.Atoi(cp.PopulationSize.Text)
	generations, _ := strconv.Atoi(cp.Generations.Text)
	mutRate, _ := strconv.ParseFloat(cp.MutationRate.Text, 64)
	crossoverRate, _ := strconv.ParseFloat(cp.CrossoverRate.Text, 64)
	numIslands, _ := strconv.Atoi(cp.NumIslands.Text)
	migrationInterval, _ := strconv.Atoi(cp.MigrationInterval.Text)

	var model genetic.EvolutionModel
	switch cp.EvolutionModel.Selected {
	case "Classic":
		model = genetic.Classic
	case "Island":
		model = genetic.Island
	case "Combined":
		model = genetic.Combined
	}

	return backend.Params{
		EvolutionModel:    model,
		PopulationSize:    popSize,
		Generations:       generations,
		MutationRate:      mutRate,
		CrossoverRate:     crossoverRate,
		NumIslands:        numIslands,
		MigrationInterval: migrationInterval,
	}
}

func (cp *ControlsPanel) Render() fyne.CanvasObject {
	return container.NewVBox(
		widget.NewLabel("Модель эволюции:"),
		cp.EvolutionModel,
		widget.NewLabel("Размер популяции:"),
		cp.PopulationSize,
		widget.NewLabel("Поколения:"),
		cp.Generations,
		widget.NewLabel("Вероятность мутации:"),
		cp.MutationRate,
		widget.NewLabel("Вероятность кроссовера:"),
		cp.CrossoverRate,
		widget.NewLabel("Количество островов:"),
		cp.NumIslands,
		widget.NewLabel("Интервал миграции:"),
		cp.MigrationInterval,
		container.NewHBox(cp.StartBtn, cp.StopBtn),
	)
}
