package backend

import (
	"Genetic-algorithm/backend/genetic"
	"time"
)

type Params struct {
	EvolutionModel    genetic.EvolutionModel
	PopulationSize    int
	Generations       int
	MutationRate      float64
	CrossoverRate     float64
	NumIslands        int
	MigrationInterval int
}

type GASolver struct {
	GA           *genetic.GeneticAlgorithm
	IsRunning    bool
	BestSolution genetic.Chromosome
	UpdateChan   chan genetic.Chromosome // Канал для отправки обновлений
}

func (s *GASolver) Start(graph genetic.Graph, params Params) {
	s.IsRunning = true
	go func() {
		ga := genetic.NewGeneticAlgorithm(
			&graph,
			params.EvolutionModel,
			params.PopulationSize,
			params.Generations,
			params.MutationRate,
			params.CrossoverRate,
			params.NumIslands,
			params.MigrationInterval,
		)
		ga.InitializePopulation()
		for gen := 0; gen < params.Generations && s.IsRunning; gen++ {
			ga.EvolvePopulation()
			best := ga.GetBestChromosome()
			s.UpdateChan <- best               // Отправляем лучшее решение в GUI
			time.Sleep(100 * time.Millisecond) // Для плавной анимации
		}
		s.IsRunning = false
	}()
}

func (s *GASolver) Stop() {
	s.IsRunning = false
}
