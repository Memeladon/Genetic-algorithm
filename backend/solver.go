package backend

import (
	"Genetic-algorithm/backend/genetic"
	"fmt"
	"github.com/fatih/color"
	"log"
)

type Params struct {
	EvolutionModel    genetic.EvolutionModel
	CrossoverStrategy genetic.CrossoverStrategy
	SelectionStrategy genetic.SelectionStrategy
	MutationStrategy  genetic.MutationStrategy
	PopulationSize    int
	Generations       int
	MutationRate      float64
	CrossoverRate     float64
	NumIslands        int
	MigrationInterval int
	TournamentSize    int // Новый параметр для турнирной селекции
}

type GASolver struct {
	GA           *genetic.Algorithm
	IsRunning    bool
	BestSolution genetic.Chromosome
	UpdateChan   chan genetic.Chromosome
}

func (s *GASolver) Start(graph genetic.Graph, params Params) {
	s.IsRunning = true
	go func() {
		ga, err := genetic.NewGeneticAlgorithm(
			&graph,
			params.EvolutionModel,
			params.CrossoverStrategy,
			params.SelectionStrategy,
			params.MutationStrategy,
			params.PopulationSize,
			params.Generations,
			params.MutationRate,
			params.CrossoverRate,
			params.NumIslands,
			params.MigrationInterval,
		)
		if err != nil {
			log.Println(err)
			s.IsRunning = false
			return
		}

		// Безопасно пытаемся вычислить MaxMatching, ловя панику
		var target int
		func() {
			defer func() {
				if r := recover(); r != nil {
					color.Yellow("MaxMatching failed (%v), disabling early stop", r)
					target = 0
				}
			}()
			target = genetic.MaxMatching(ga.Graph)
		}()
		if target > 0 {
			color.Cyan("Target (max matching) = %d", target)
		} else {
			color.Yellow("Target unknown or error, running all %d generations", ga.Generations)
		}

		// Инициализация популяции
		ga.InitializePopulation()
		color.Cyan("Start GA: model=%v, popSize=%d, gens=%d",
			ga.EvolutionModel, ga.PopulationSize, ga.Generations)

		// Первое обновление
		best := ga.GetBestChromosome()
		color.Green("Initial best fitness: %d", best.Fitness)
		s.UpdateChan <- best

		// Основной цикл
		for gen := 1; gen <= ga.Generations && s.IsRunning; gen++ {
			switch ga.EvolutionModel {
			case genetic.Island:
				color.Magenta("Island model: islands=%d, interval=%d",
					ga.NumIslands, ga.MigrationInterval)
				islands := ga.DistributePopulation()
				for i := range islands {
					islands[i] = ga.EvolveIsland(islands[i])
				}
				if gen%ga.MigrationInterval == 0 {
					islands = genetic.MigrateIslands(islands)
					color.Blue(" Migration at gen %d", gen)
				}
				ga.Population = genetic.MergeIslands(islands)

			case genetic.SteadyState:
				p1 := ga.SelectionStrategy.Strategy.Select(ga.Population)
				p2 := ga.SelectionStrategy.Strategy.Select(ga.Population)
				child := ga.CrossoverStrategy.Crossover(p1, p2)
				ga.MutationStrategy.Mutate(&child, ga.MutationRate, ga.Graph)
				genetic.RepairFast(&child, ga.Graph)
				genetic.Evaluate(&child, ga.Graph)
				// замена худшей
				worst := 0
				for i, c := range ga.Population {
					if c.Fitness < ga.Population[worst].Fitness {
						worst = i
					}
				}
				if child.Fitness > ga.Population[worst].Fitness {
					ga.Population[worst] = child
				}

			case genetic.Memetic:
				fmt.Println("Running Memetic GA")
				ga.RunMemetic()

			default: // Classic и Combined
				ga.EvolvePopulation()
			}

			// Лог и обновление лучшего
			current := ga.GetBestChromosome()
			color.Yellow("Gen %d best fitness: %d", gen, current.Fitness)
			if current.Fitness > best.Fitness {
				best = current
				color.Green(" New global best: %d", best.Fitness)
			}
			s.UpdateChan <- best

			// Досрочный выход
			if target > 0 && best.Fitness >= target {
				color.Green("Reached optimal at gen %d", gen)
				break
			}
		}

		color.Cyan("GA finished. Best fitness = %d", best.Fitness)
		s.BestSolution = best
		s.IsRunning = false
	}()
}

func (s *GASolver) Stop() {
	s.IsRunning = false
}
