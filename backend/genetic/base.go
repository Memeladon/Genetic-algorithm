package genetic

import (
	"errors"
)

// NewGeneticAlgorithm создаёт экземпляр алгоритма с заданными параметрами.
func NewGeneticAlgorithm(
	graph *Graph,
	evolutionModel EvolutionModel,
	crossoverStrategy CrossoverStrategy,
	selectionStrategy SelectionStrategy,
	mutationStrategy MutationStrategy,
	populationSize, generations int,
	mutationRate, crossoverRate float64,
	numIslands, migrationInterval int,
) (*Algorithm, error) {
	if populationSize < 1 {
		return nil, errors.New("populationSize must be ≥ 1")
	}
	if mutationRate < 0 || mutationRate > 1 {
		return nil, errors.New("mutationRate must be in [0, 1]")
	}

	config := EvolutionModelConfig{
		Model: evolutionModel,

		UseSelection:   true,
		UseCrossover:   true,
		UseMutation:    true,
		UseLocalSearch: evolutionModel == Memetic,
	}

	modelStrategy, err := NewEvolutionModelStrategy(config)
	if err != nil {
		return nil, err
	}

	// Привязываем rate к кроссоверу
	cs := crossoverStrategy.WithRate(crossoverRate)

	// Если это турнирная селекция, убедимся, что EliteSize и TournamentSize заданы
	eliteSize := populationSize / 10
	if eliteSize < 1 {
		eliteSize = 1
	}
	ew := ElitismWrapper{
		Strategy:  selectionStrategy,
		EliteSize: eliteSize,
	}

	ga := &Algorithm{
		Graph:             graph,
		EvolutionModel:    modelStrategy,
		CrossoverStrategy: cs,
		SelectionStrategy: ew,
		MutationStrategy:  mutationStrategy,
		PopulationSize:    populationSize,
		Generations:       generations,
		MutationRate:      mutationRate,
		CrossoverRate:     crossoverRate,
		NumIslands:        numIslands,
		MigrationInterval: migrationInterval,
		Logger:            NewLogger(),
		optimalSize:       MaxMatchingGreed(graph),
	}

	if err := modelStrategy.ValidateStrategies(ga); err != nil {
		return nil, err
	}

	return ga, nil
}
