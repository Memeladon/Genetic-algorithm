package genetic

// EvolvePopulation выполняет один шаг эволюции для текущей популяции.
func (ga *Algorithm) EvolvePopulation() {
	newPopulation := make([]Chromosome, 0, ga.PopulationSize)

	elites := ga.getElites()
	newPopulation = append(newPopulation, elites...)

	for len(newPopulation) < ga.PopulationSize {
		parent1 := ga.SelectionStrategy.Strategy.Select(ga.Population)
		parent2 := ga.SelectionStrategy.Strategy.Select(ga.Population)

		child := ga.CrossoverStrategy.Crossover(parent1, parent2)

		// Применяем мутацию через стратегию
		ga.MutationStrategy.Mutate(&child, ga.MutationRate, ga.Graph)

		RepairFast(&child, ga.Graph)
		//EvaluateFast(&child, ga.Graph)
		Evaluate(&child, ga.Graph)
		newPopulation = append(newPopulation, child)
	}

	ga.Population = newPopulation

}

func (ga *Algorithm) EvolveIsland(island []Chromosome) []Chromosome {
	newPopulation := make([]Chromosome, 0, len(island))

	// Добавляем элиту
	elites := ga.getElitesFromIsland(island, ga.SelectionStrategy.EliteSize)
	newPopulation = append(newPopulation, elites...)

	for len(newPopulation) < len(island) {
		parent1 := ga.SelectionStrategy.Strategy.Select(island)
		parent2 := ga.SelectionStrategy.Strategy.Select(island)

		child := ga.CrossoverStrategy.Crossover(parent1, parent2)

		// Применяем мутацию через стратегию
		ga.MutationStrategy.Mutate(&child, ga.MutationRate, ga.Graph)

		RepairFast(&child, ga.Graph)
		//EvaluateFast(&child, ga.Graph)
		Evaluate(&child, ga.Graph)
		newPopulation = append(newPopulation, child)
	}
	return newPopulation
}
