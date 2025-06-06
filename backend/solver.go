package backend

import (
	"Genetic-algorithm/backend/genetic"
	"log"
	"time"

	"github.com/fatih/color"
)

// Params содержит параметры для запуска генетического алгоритма
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
	TournamentSize    int            // Новый параметр для турнирной селекции
	Config            genetic.Config // Конфигурация генетического алгоритма
}

// ExperimentResult содержит результаты одного эксперимента
type ExperimentResult struct {
	GraphName           string
	Algorithm           string
	GraphVertices       int
	GraphEdges          int
	TimeTaken           time.Duration
	BestFitness         int
	AverageFitness      float64
	FitnessHistory      []int
	BestMatchingEdges   []int  // Индексы рёбер в наибольшем допустимом паросочетании
	BestChromosomeGenes []bool // Гены лучшей хромосомы
}

// GASolver представляет решатель задачи о максимальном паросочетании
type GASolver struct {
	Graph        *genetic.Graph
	Params       Params
	StopChan     chan struct{}
	Done         chan struct{}
	GA           *genetic.Algorithm
	IsRunning    bool
	BestSolution genetic.Chromosome
	UpdateChan   chan genetic.Chromosome
	Results      []ExperimentResult
}

// NewGASolver создаёт новый экземпляр решателя
func NewGASolver(graph *genetic.Graph, params Params) *GASolver {
	return &GASolver{
		Graph:    graph,
		Params:   params,
		StopChan: make(chan struct{}),
		Done:     make(chan struct{}),
	}
}

func (s *GASolver) Start(graph genetic.Graph, params Params, graphName string) {
	s.IsRunning = true
	s.Done = make(chan struct{}) // Create new done channel
	startTime := time.Now()

	// Инициализация результата
	result := ExperimentResult{
		GraphName:      graphName,
		Algorithm:      params.EvolutionModel.String(),
		GraphVertices:  graph.NumVertices,
		GraphEdges:     len(graph.Edges),
		FitnessHistory: make([]int, 0, params.Generations),
	}

	go func() {
		defer func() {
			s.IsRunning = false
			close(s.UpdateChan) // Close update channel first
			close(s.Done)       // Then signal completion
		}()

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
			return
		}

		// Log algorithm start
		ga.Logger.LogAlgorithmStart(ga)

		// Вычисление условия остановки работы
		var target int
		func() {
			defer func() {
				if r := recover(); r != nil {
					color.Yellow("MaxMatchingGreed failed (%v), disabling early stop", r)
					target = 0
				}
			}()
			target = genetic.MaxMatchingGreed(ga.Graph)
		}()

		// Log target
		if target > 0 {
			ga.Logger.LogMilestone("Target (max matching) = %d", target)
		} else {
			ga.Logger.LogWarning("Target unknown or error, running all %d generations", ga.Generations)
		}

		// Инициализация популяции
		ga.InitializePopulation()
		ga.Logger.LogInfo("Start GA: model=%v, popSize=%d, gens=%d",
			ga.EvolutionModel.GetModelName(), ga.PopulationSize, ga.Generations)

		// Первое обновление
		best := ga.GetBestChromosome()
		// initBestValid := countValidMatchingEdges(best, ga.Graph)
		// ga.Logger.LogSuccess("Initial best matching: %d", initBestValid)
		s.UpdateChan <- best

		// Основной цикл
		for gen := 1; gen <= ga.Generations && s.IsRunning; gen++ {
			ga.CurrentGeneration = gen
			// Evolve population using the selected model strategy
			if err := ga.EvolutionModel.Evolve(ga); err != nil {
				log.Printf("Error in evolution: %v", err)
				break
			}

			// Лог и обновление лучшего
			current := ga.GetBestChromosome()
			ga.SetLocalBest(current)
			if ga.LocalBestEdges > ga.BestSoFarEdges {
				genetic.RepairFast(&current, ga.Graph)
				genetic.Evaluate(&current, ga.Graph)
				ga.SetBestSoFar(current)
				ga.Logger.LogMilestone("Найдено новое лучшее паросочетание: %d рёбер (поколение %d)", ga.BestSoFarEdges, ga.CurrentGeneration)
			}
			best = current
			s.UpdateChan <- best
			result.FitnessHistory = append(result.FitnessHistory, ga.BestSoFarEdges)

			// Досрочный выход
			if target > 0 && ga.BestSoFarEdges >= target {
				ga.Logger.LogSuccess("Reached optimal at gen %d", gen)
				break
			}

		}

		finalBest := ga.GetBestSoFar()
		finalBestValid := countValidMatchingEdges(finalBest, ga.Graph)
		ga.Logger.LogSuccess("GA finished. Best matching = %d", finalBestValid)

		// Фиксируем результаты
		result.TimeTaken = time.Since(startTime)
		result.BestFitness = finalBestValid
		// Сохраняем индексы рёбер наибольшего паросочетания из глобального bestSoFar
		globalBest := ga.GetBestSoFar()
		result.BestMatchingEdges = getValidMatchingEdges(globalBest, ga.Graph)
		result.BestChromosomeGenes = make([]bool, len(globalBest.Genes))
		copy(result.BestChromosomeGenes, globalBest.Genes)
		s.Results = append(s.Results, result)

		s.BestSolution = finalBest
	}()
}

func (s *GASolver) Stop() {
	s.IsRunning = false
	<-s.Done // Wait for completion
}

func (s *GASolver) Run() error {
	ga, err := genetic.NewGeneticAlgorithm(
		s.Graph,
		s.Params.EvolutionModel,
		s.Params.CrossoverStrategy,
		s.Params.SelectionStrategy,
		s.Params.MutationStrategy,
		s.Params.PopulationSize,
		s.Params.Generations,
		s.Params.MutationRate,
		s.Params.CrossoverRate,
		s.Params.NumIslands,
		s.Params.MigrationInterval,
	)
	if err != nil {
		return err
	}

	ga.Logger.LogAlgorithmStart(ga)

	ga.InitializePopulation()

	// Main evolution loop
	for !ga.ShouldTerminate() {
		select {
		case <-s.StopChan:
			ga.Logger.LogWarning("Алгоритм остановлен пользователем")
			return nil
		default:
			if err := ga.EvolutionModel.Evolve(ga); err != nil {
				ga.Logger.LogError(err)
				return err
			}
		}
	}

	// Log completion
	ga.Logger.LogCompletion(ga)
	return nil
}

// getValidMatchingEdges возвращает индексы рёбер в допустимом паросочетании для данной хромосомы
func getValidMatchingEdges(chrom genetic.Chromosome, graph *genetic.Graph) []int {
	used := make(map[int]bool)
	var indices []int
	for i, gene := range chrom.Genes {
		if gene {
			edge := graph.Edges[i]
			if !used[edge.U] && !used[edge.V] {
				used[edge.U] = true
				used[edge.V] = true
				indices = append(indices, i)
			}
		}
	}
	return indices
}

// countValidMatchingEdges возвращает количество рёбер в допустимом паросочетании для данной хромосомы
func countValidMatchingEdges(chrom genetic.Chromosome, graph *genetic.Graph) int {
	used := make(map[int]bool)
	count := 0
	for i, gene := range chrom.Genes {
		if gene {
			edge := graph.Edges[i]
			if !used[edge.U] && !used[edge.V] {
				used[edge.U] = true
				used[edge.V] = true
				count++
			}
		}
	}
	return count
}
