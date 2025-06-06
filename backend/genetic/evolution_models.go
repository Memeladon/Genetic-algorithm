package genetic

import (
	"errors"
	"fmt"
	"math/rand"
)

// ClassicEvolutionModel реализует классический генетический алгоритм
// Использует элитизм и генерационную модель эволюции
type ClassicEvolutionModel struct{}

// Evolve выполняет один шаг эволюции в классической модели
// 1. Сохраняет элиту популяции
// 2. Генерирует новое поколение через селекцию, скрещивание и мутацию
// 3. Поддерживает постоянный размер популяции
func (m *ClassicEvolutionModel) Evolve(ga *Algorithm) error {
	if len(ga.Population) == 0 {
		ga.Logger.LogError(errors.New("пустая популяция"))
		return errors.New("empty population")
	}

	newPop := make([]Chromosome, 0, ga.PopulationSize)

	// Elitism
	elites := ga.getElites()
	newPop = append(newPop, elites...)
	ga.Logger.LogDebug("Сохранено %d элитных особей", len(elites))

	// Generate rest of population
	for len(newPop) < ga.PopulationSize {
		// Selection
		p1 := ga.SelectionStrategy.Strategy.Select(ga.Population)
		p2 := ga.SelectionStrategy.Strategy.Select(ga.Population)

		// Crossover
		child := ga.CrossoverStrategy.Crossover(p1, p2)

		// Mutation
		ga.MutationStrategy.Mutate(&child, ga.MutationRate, ga.Graph)

		// Repair if needed
		RepairFast(&child, ga.Graph)

		// Explicit fitness evaluation
		Evaluate(&child, ga.Graph)

		newPop = append(newPop, child)
	}

	// Ensure population size is maintained
	if len(newPop) != ga.PopulationSize {
		err := fmt.Errorf("несоответствие размера популяции: ожидалось %d, получено %d",
			ga.PopulationSize, len(newPop))
		ga.Logger.LogError(err)
		return err
	}

	ga.Population = newPop
	ga.SetLocalBest(ga.GetBestChromosome())
	ga.CurrentGeneration++

	// Log generation information
	ga.Logger.LogGeneration(ga)

	// Check if we should terminate and log completion
	if ga.ShouldTerminate() {
		ga.Logger.LogCompletion(ga)
	}

	return nil
}

func (m *ClassicEvolutionModel) GetRequiredStrategies() []string {
	return []string{"Selection", "Crossover", "Mutation"}
}

func (m *ClassicEvolutionModel) ValidateStrategies(ga *Algorithm) error {
	if ga.SelectionStrategy.Strategy == nil {
		return errors.New("selection strategy is required for classic model")
	}
	if ga.CrossoverStrategy == nil {
		return errors.New("crossover strategy is required for classic model")
	}
	if ga.MutationStrategy == nil {
		return errors.New("mutation strategy is required for classic model")
	}
	if ga.PopulationSize < 2 {
		return errors.New("population size must be at least 2")
	}
	return nil
}

func (m *ClassicEvolutionModel) GetModelName() string {
	return "Classic"
}

func (m *ClassicEvolutionModel) String() string {
	return "Classic"
}

// IslandEvolutionModel реализует островную модель генетического алгоритма
// Разделяет популяцию на изолированные подпопуляции (острова)
// Периодически обменивается особями между островами
type IslandEvolutionModel struct{}

func (m *IslandEvolutionModel) Evolve(ga *Algorithm) error {
	islands := ga.DistributePopulation()
	for i := range islands {
		islands[i] = ga.EvolveIsland(islands[i])
	}
	if ga.CurrentGeneration%ga.MigrationInterval == 0 {
		islands = MigrateIslands(islands)
	}
	ga.Population = MergeIslands(islands)
	ga.CurrentGeneration++
	ga.Logger.LogGeneration(ga)

	// Check if we should terminate and log completion
	if ga.ShouldTerminate() {
		ga.Logger.LogCompletion(ga)
	}

	return nil
}

func (m *IslandEvolutionModel) GetRequiredStrategies() []string {
	return []string{"Selection", "Crossover", "Mutation"}
}

func (m *IslandEvolutionModel) ValidateStrategies(ga *Algorithm) error {
	if ga.SelectionStrategy.Strategy == nil {
		return errors.New("selection strategy is required for island model")
	}
	if ga.CrossoverStrategy == nil {
		return errors.New("crossover strategy is required for island model")
	}
	if ga.MutationStrategy == nil {
		return errors.New("mutation strategy is required for island model")
	}
	return nil
}

func (m *IslandEvolutionModel) GetModelName() string {
	return "Island"
}

func (m *IslandEvolutionModel) String() string {
	return "Island"
}

// SteadyStateEvolutionModel реализует модель с постоянным состоянием
// В каждом поколении заменяет только одну особь
// Подходит для задач, где важно сохранять хорошие решения
type SteadyStateEvolutionModel struct{}

func (m *SteadyStateEvolutionModel) Evolve(ga *Algorithm) error {
	p1 := ga.SelectionStrategy.Strategy.Select(ga.Population)
	p2 := ga.SelectionStrategy.Strategy.Select(ga.Population)
	child := ga.CrossoverStrategy.Crossover(p1, p2)
	ga.MutationStrategy.Mutate(&child, ga.MutationRate, ga.Graph)
	RepairFast(&child, ga.Graph)
	Evaluate(&child, ga.Graph)

	// Replace worst individual
	worst := 0
	for i, c := range ga.Population {
		if c.Fitness < ga.Population[worst].Fitness {
			worst = i
		}
	}
	if child.Fitness > ga.Population[worst].Fitness {
		ga.Population[worst] = child
		// Update bestSoFar if the new child is better
		ga.SetBestSoFar(child)
	}
	ga.CurrentGeneration++
	ga.Logger.LogGeneration(ga)

	// Check if we should terminate and log completion
	if ga.ShouldTerminate() {
		ga.Logger.LogCompletion(ga)
	}

	return nil
}

func (m *SteadyStateEvolutionModel) GetRequiredStrategies() []string {
	return []string{"Selection", "Crossover", "Mutation"}
}

func (m *SteadyStateEvolutionModel) ValidateStrategies(ga *Algorithm) error {
	if ga.SelectionStrategy.Strategy == nil {
		return errors.New("selection strategy is required for steady state model")
	}
	if ga.CrossoverStrategy == nil {
		return errors.New("crossover strategy is required for steady state model")
	}
	if ga.MutationStrategy == nil {
		return errors.New("mutation strategy is required for steady state model")
	}
	return nil
}

func (m *SteadyStateEvolutionModel) GetModelName() string {
	return "SteadyState"
}

func (m *SteadyStateEvolutionModel) String() string {
	return "SteadyState"
}

// MemeticEvolutionModel реализует меметический алгоритм
// Комбинирует генетический алгоритм с локальным поиском
// Применяет локальное улучшение к каждой новой особи
type MemeticEvolutionModel struct{}

func (m *MemeticEvolutionModel) Evolve(ga *Algorithm) error {
	newPop := make([]Chromosome, 0, ga.PopulationSize)

	// Elitism
	elites := ga.GetAllBestChromosomes()
	newPop = append(newPop, elites...)

	// Generate rest through memetic loop
	for len(newPop) < ga.PopulationSize {
		p1 := ga.SelectionStrategy.Strategy.Select(ga.Population)
		p2 := ga.SelectionStrategy.Strategy.Select(ga.Population)
		child := ga.CrossoverStrategy.Crossover(p1, p2)
		ga.MutationStrategy.Mutate(&child, ga.MutationRate, ga.Graph)
		// Local search: one iteration of augmenting path
		ApplyAugmentingPath(child.Genes, ga.Graph)
		RepairFast(&child, ga.Graph)
		Evaluate(&child, ga.Graph)
		newPop = append(newPop, child)
	}

	ga.Population = newPop
	ga.SetLocalBest(ga.GetBestChromosome())
	ga.CurrentGeneration++
	ga.Logger.LogGeneration(ga)

	// Check if we should terminate and log completion
	if ga.ShouldTerminate() {
		ga.Logger.LogCompletion(ga)
	}

	return nil
}

func (m *MemeticEvolutionModel) GetRequiredStrategies() []string {
	return []string{"Selection", "Crossover", "Mutation"}
}

func (m *MemeticEvolutionModel) ValidateStrategies(ga *Algorithm) error {
	if ga.SelectionStrategy.Strategy == nil {
		return errors.New("selection strategy is required for memetic model")
	}
	if ga.CrossoverStrategy == nil {
		return errors.New("crossover strategy is required for memetic model")
	}
	if ga.MutationStrategy == nil {
		return errors.New("mutation strategy is required for memetic model")
	}
	return nil
}

func (m *MemeticEvolutionModel) GetModelName() string {
	return "Memetic"
}

func (m *MemeticEvolutionModel) String() string {
	return "Memetic"
}

// CombinedEvolutionModel реализует гибкую модель эволюции
// Позволяет включать/отключать различные операторы
// Подходит для экспериментов с разными комбинациями стратегий
type CombinedEvolutionModel struct {
	Config EvolutionModelConfig // Конфигурация модели
}

func (m *CombinedEvolutionModel) Evolve(ga *Algorithm) error {
	newPop := make([]Chromosome, 0, ga.PopulationSize)

	// Elitism
	elites := ga.getElites()
	newPop = append(newPop, elites...)

	// Generate rest of population
	for len(newPop) < ga.PopulationSize {
		var child Chromosome

		// Selection
		if m.Config.UseSelection {
			p1 := ga.SelectionStrategy.Strategy.Select(ga.Population)
			p2 := ga.SelectionStrategy.Strategy.Select(ga.Population)
			child = ga.CrossoverStrategy.Crossover(p1, p2)
		} else {
			// Random selection if no selection strategy
			child = ga.Population[rand.Intn(len(ga.Population))]
		}

		// Crossover
		if m.Config.UseCrossover {
			p2 := ga.SelectionStrategy.Strategy.Select(ga.Population)
			child = ga.CrossoverStrategy.Crossover(child, p2)
		}

		// Mutation
		if m.Config.UseMutation {
			ga.MutationStrategy.Mutate(&child, ga.MutationRate, ga.Graph)
		}

		// Local search
		if m.Config.UseLocalSearch {
			ApplyAugmentingPath(child.Genes, ga.Graph)
		}

		RepairFast(&child, ga.Graph)
		newPop = append(newPop, child)
	}

	ga.Population = newPop
	ga.SetLocalBest(ga.GetBestChromosome())
	ga.CurrentGeneration++
	ga.Logger.LogGeneration(ga)

	// Check if we should terminate and log completion
	if ga.ShouldTerminate() {
		ga.Logger.LogCompletion(ga)
	}

	return nil
}

func (m *CombinedEvolutionModel) GetRequiredStrategies() []string {
	var strategies []string
	if m.Config.UseSelection {
		strategies = append(strategies, "Selection")
	}
	if m.Config.UseCrossover {
		strategies = append(strategies, "Crossover")
	}
	if m.Config.UseMutation {
		strategies = append(strategies, "Mutation")
	}
	return strategies
}

func (m *CombinedEvolutionModel) ValidateStrategies(ga *Algorithm) error {
	if m.Config.UseSelection && ga.SelectionStrategy.Strategy == nil {
		return errors.New("selection strategy is required but not provided")
	}
	if m.Config.UseCrossover && ga.CrossoverStrategy == nil {
		return errors.New("crossover strategy is required but not provided")
	}
	if m.Config.UseMutation && ga.MutationStrategy == nil {
		return errors.New("mutation strategy is required but not provided")
	}
	return nil
}

func (m *CombinedEvolutionModel) GetModelName() string {
	return "Combined"
}

func (m *CombinedEvolutionModel) String() string {
	return "Combined"
}

// Factory function to create evolution model strategies
func NewEvolutionModelStrategy(config EvolutionModelConfig) (EvolutionModelStrategy, error) {
	switch config.Model {
	case Classic:
		return &ClassicEvolutionModel{}, nil
	case Island:
		return &IslandEvolutionModel{}, nil
	case SteadyState:
		return &SteadyStateEvolutionModel{}, nil
	case Memetic:
		return &MemeticEvolutionModel{}, nil
	case Combined:
		return &CombinedEvolutionModel{Config: config}, nil
	default:
		return nil, fmt.Errorf("unsupported evolution model: %v", config.Model)
	}
}
