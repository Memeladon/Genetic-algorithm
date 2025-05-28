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

	CrossoverType  *widget.RadioGroup
	MutationType   *widget.RadioGroup
	SelectionType  *widget.RadioGroup
	TournamentSize *widget.Entry
	OnStart        func()
	OnStop         func()
}

func NewControlsPanel() *ControlsPanel {
	cp := &ControlsPanel{
		PopulationSize:    widget.NewEntry(),
		Generations:       widget.NewEntry(),
		MutationRate:      widget.NewEntry(),
		CrossoverRate:     widget.NewEntry(),
		NumIslands:        widget.NewEntry(),
		MigrationInterval: widget.NewEntry(),
		EvolutionModel:    widget.NewRadioGroup([]string{"Classic", "Island", "Steady-State", "Memetic", "Combined"}, nil),

		CrossoverType:  widget.NewRadioGroup([]string{"Single-point", "Two-point", "Combined"}, nil),
		MutationType:   widget.NewRadioGroup([]string{"Classic", "Island", "Steady-State", "Conflict-Adaptive", "Augmenting-Path", "Combined"}, nil),
		SelectionType:  widget.NewRadioGroup([]string{"Tournament", "Roulette", "Rank"}, nil),
		TournamentSize: widget.NewEntry(),
	}
	cp.setDefaults()

	cp.StartBtn = widget.NewButton("Start", func() {
		if cp.OnStart != nil {
			cp.OnStart()
		}
	})
	cp.StopBtn = widget.NewButton("Stop", func() {
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

	cp.EvolutionModel.SetSelected("Classic")
	cp.CrossoverType.SetSelected("Single-point")
	cp.MutationType.SetSelected("Classic")
	cp.SelectionType.SetSelected("Tournament")
	cp.TournamentSize.SetText("3")
}

func (cp *ControlsPanel) GetParams() backend.Params {
	popSize, _ := strconv.Atoi(cp.PopulationSize.Text)
	gens, _ := strconv.Atoi(cp.Generations.Text)
	mutRate, _ := strconv.ParseFloat(cp.MutationRate.Text, 64)
	crossRate, _ := strconv.ParseFloat(cp.CrossoverRate.Text, 64)
	nIslands, _ := strconv.Atoi(cp.NumIslands.Text)
	migInt, _ := strconv.Atoi(cp.MigrationInterval.Text)
	tSize, _ := strconv.Atoi(cp.TournamentSize.Text)

	var model genetic.EvolutionModel
	switch cp.EvolutionModel.Selected {
	case "Classic":
		model = genetic.Classic
	case "Island":
		model = genetic.Island
	case "Steady-State":
		model = genetic.SteadyState
	case "Memetic":
		model = genetic.Memetic
	case "Combined":
		model = genetic.Combined
	}

	var cross genetic.CrossoverStrategy
	switch cp.CrossoverType.Selected {
	case "Single-point":
		cross = &genetic.SinglePoint{}
	case "Two-point":
		cross = &genetic.TwoPoint{}
	case "Combined":
		cross = &genetic.CombinedCrossover{}
	}

	var mut genetic.MutationStrategy
	switch cp.MutationType.Selected {
	case "Classic":
		mut = &genetic.ClassicMutationStrategy{}
	case "Island":
		mut = &genetic.IslandMutationStrategy{}
	case "Steady-State":
		mut = &genetic.SteadyStateMutationStrategy{}
	case "Conflict-Adaptive":
		mut = &genetic.ConflictAdaptiveMutationStrategy{}
	case "Augmenting-Path":
		mut = &genetic.AugmentingPathMutationStrategy{}
	case "Combined":
		mut = &genetic.CombinedMutationStrategy{Strategies: []genetic.MutationStrategy{
			&genetic.ClassicMutationStrategy{}, &genetic.IslandMutationStrategy{},
		}}
	}

	var sel genetic.SelectionStrategy
	switch cp.SelectionType.Selected {
	case "Tournament":
		sel = &genetic.TournamentSelectionStrategy{TournamentSize: tSize}
	case "Roulette":
		sel = &genetic.RouletteWheelSelectionStrategy{}
	case "Rank":
		sel = &genetic.RankSelectionStrategy{}
	}

	return backend.Params{
		EvolutionModel:    model,
		PopulationSize:    popSize,
		Generations:       gens,
		MutationRate:      mutRate,
		CrossoverRate:     crossRate,
		NumIslands:        nIslands,
		MigrationInterval: migInt,
		TournamentSize:    tSize,
		CrossoverStrategy: cross,
		MutationStrategy:  mut,
		SelectionStrategy: sel,
	}
}

func (cp *ControlsPanel) Render() fyne.CanvasObject {
	// Используем Accordion для сворачивания секций
	acc := widget.NewAccordion(
		widget.NewAccordionItem("Evolution Model & Basic", container.NewVBox(
			widget.NewLabel("Evolution Model:"), cp.EvolutionModel,
		)),
		widget.NewAccordionItem("Genetic Operators", container.NewVBox(
			widget.NewLabel("Crossover Type:"), cp.CrossoverType,
			widget.NewLabel("Mutation Type:"), cp.MutationType,
			widget.NewLabel("Selection Type:"), cp.SelectionType,
			widget.NewLabel("Tournament Size:"), cp.TournamentSize,
		)),
		widget.NewAccordionItem("Population & Generations", container.NewVBox(
			widget.NewLabel("Population Size:"), cp.PopulationSize,
			widget.NewLabel("Crossover Rate:"), cp.CrossoverRate,
			widget.NewLabel("Mutation Rate:"), cp.MutationRate,
			widget.NewLabel("Generations:"), cp.Generations,
		)),
		widget.NewAccordionItem("Island Parameters", container.NewVBox(
			widget.NewLabel("Num Islands:"), cp.NumIslands,
			widget.NewLabel("Migration Interval:"), cp.MigrationInterval,
		)),
	)
	btns := container.NewHBox(cp.StartBtn, cp.StopBtn)
	return container.NewBorder(nil, btns, nil, nil, acc)
}
