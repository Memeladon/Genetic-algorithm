package genetic

import (
	"errors"
	"fmt"
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

	// Привязываем rate к кроссоверу
	cs := crossoverStrategy.WithRate(crossoverRate)

	// Если это турнирная селекция, убедимся, что EliteSize и TournamentSize заданы
	ew := ElitismWrapper{
		Strategy:  selectionStrategy,
		EliteSize: populationSize / 10,
	}

	return &Algorithm{
		Graph:             graph,
		EvolutionModel:    evolutionModel,
		CrossoverStrategy: cs,
		SelectionStrategy: ew,
		MutationStrategy:  mutationStrategy,
		PopulationSize:    populationSize,
		Generations:       generations,
		MutationRate:      mutationRate,
		CrossoverRate:     crossoverRate,
		NumIslands:        numIslands,
		MigrationInterval: migrationInterval,
	}, nil
}

// Run запускает генетический алгоритм и возвращает лучшее найденное решение.
func (ga *Algorithm) Run() Chromosome {
	ga.InitializePopulation()
	for gen := 0; gen < ga.Generations; gen++ {
		ga.EvolvePopulation()
	}
	return ga.GetBestChromosome()
}

func (ga *Algorithm) runSteadyState() {
	ga.InitializePopulation()
	for _, c := range ga.Population {
		if c.Fitness > ga.bestSoFar.Fitness {
			ga.bestSoFar = c
		}
	}
	for gen := 0; gen < ga.Generations; gen++ {
		p1 := ga.SelectionStrategy.Strategy.Select(ga.Population)
		p2 := ga.SelectionStrategy.Strategy.Select(ga.Population)
		child := ga.CrossoverStrategy.Crossover(p1, p2)
		ga.MutationStrategy.Mutate(&child, ga.MutationRate, ga.Graph)
		RepairFast(&child, ga.Graph)
		Evaluate(&child, ga.Graph)
		// замена худшей
		worstIdx := 0
		for i, c := range ga.Population {
			if c.Fitness < ga.Population[worstIdx].Fitness {
				worstIdx = i
			}
		}
		if child.Fitness > ga.Population[worstIdx].Fitness {
			ga.Population[worstIdx] = child
			if child.Fitness > ga.bestSoFar.Fitness {
				ga.bestSoFar = child
			}
		}
	}
}

func (ga *Algorithm) RunMemetic() {
	// Инициализация популяции
	ga.InitializePopulation()
	// Фиксируем bestSoFar
	for _, c := range ga.Population {
		if c.Fitness > ga.bestSoFar.Fitness {
			ga.bestSoFar = c
		}
	}

	// Основной цикл поколений
	for gen := 1; gen <= ga.Generations; gen++ {
		newPop := make([]Chromosome, 0, ga.PopulationSize)

		// Не удаляем элиту
		elites := ga.getElites()
		newPop = append(newPop, elites...)

		// Генерируем остальных
		for len(newPop) < ga.PopulationSize {
			p1 := ga.SelectionStrategy.Strategy.Select(ga.Population)
			p2 := ga.SelectionStrategy.Strategy.Select(ga.Population)

			// Кроссовер и мутация
			child := ga.CrossoverStrategy.Crossover(p1, p2)
			ga.MutationStrategy.Mutate(&child, ga.MutationRate, ga.Graph)

			// Локальный поиск: применяем одну итерацию увеличивающей цепи
			// если стратегия Augmenting-Path выбрана — она уже умеет это делать;
			// иначе — можно вызвать специальный метод:
			ApplyAugmentingPath(child.Genes, ga.Graph)

			// Ремонт и оценка
			RepairFast(&child, ga.Graph)
			Evaluate(&child, ga.Graph)

			newPop = append(newPop, child)
		}

		ga.Population = newPop

		// Обновляем глобальный best
		current := ga.GetBestChromosome()
		fmt.Printf("[Memetic] Gen %d best = %d\n", gen, current.Fitness)
		if current.Fitness > ga.bestSoFar.Fitness {
			ga.bestSoFar = current
			fmt.Printf("[Memetic] New bestSoFar = %d\n", ga.bestSoFar.Fitness)
		}
	}
}
