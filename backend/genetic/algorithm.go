package genetic

import (
	"math/rand"
	"strings"
	"time"
)

// NewGeneticAlgorithm создаёт экземпляр алгоритма с заданными параметрами.
func NewGeneticAlgorithm(graph *Graph,
	model EvolutionModel,
	popSize, generations int,
	mutationRate, crossoverRate float64,
	numIslands, migrationInterval int,
) *GeneticAlgorithm {
	return &GeneticAlgorithm{
		Graph:             graph,
		EvolutionModel:    model,
		PopulationSize:    popSize,
		Generations:       generations,
		MutationRate:      mutationRate,
		CrossoverRate:     crossoverRate,
		NumIslands:        numIslands,
		MigrationInterval: migrationInterval,
	}
}

// InitializePopulation генерирует начальную популяцию.
func (ga *GeneticAlgorithm) InitializePopulation() {
	ga.Population = make([]Chromosome, ga.PopulationSize)
	numEdges := len(ga.Graph.Edges)
	for i := 0; i < ga.PopulationSize; i++ {
		genes := make([]bool, numEdges)
		// Случайное включение ребра с вероятностью 0.5.
		for j := 0; j < numEdges; j++ {
			genes[j] = rand.Float64() < 0.5
		}
		chrom := Chromosome{Genes: genes}
		Repair(&chrom, ga.Graph)
		Evaluate(&chrom, ga.Graph)
		ga.Population[i] = chrom
	}
}

// Evaluate вычисляет функцию приспособленности – число ребер в допустимом паросочетании.
func Evaluate(chrom *Chromosome, graph *Graph) {
	count := 0
	for _, gene := range chrom.Genes {
		if gene {
			count++
		}
	}
	chrom.Fitness = count
}

// Repair приводит хромосому к допустимому виду: удаляет ребра, нарушающие условие (общая вершина).
func Repair(chrom *Chromosome, graph *Graph) {
	used := make(map[int]bool)
	indices := make([]int, len(graph.Edges))
	for i := range indices {
		indices[i] = i
	}
	rand.Shuffle(len(indices), func(i, j int) {
		indices[i], indices[j] = indices[j], indices[i]
	})

	for _, idx := range indices {
		if chrom.Genes[idx] {
			edge := graph.Edges[idx]
			if used[edge.U] || used[edge.V] {
				chrom.Genes[idx] = false
			} else {
				used[edge.U] = true
				used[edge.V] = true
			}
		}
	}
}

// TournamentSelection выбирает родителя из популяции методом турнира.
func TournamentSelection(population []Chromosome, tournamentSize int) Chromosome {
	best := population[rand.Intn(len(population))]
	for i := 1; i < tournamentSize; i++ {
		contender := population[rand.Intn(len(population))]
		if contender.Fitness > best.Fitness {
			best = contender
		}
	}
	return best
}

// ClassicCrossover выполняет одноточечный кроссовер между двумя родителями.
func ClassicCrossover(parent1, parent2 Chromosome) Chromosome {
	length := len(parent1.Genes)
	point := rand.Intn(length)
	childGenes := make([]bool, length)
	for i := 0; i < length; i++ {
		if i < point {
			childGenes[i] = parent1.Genes[i]
		} else {
			childGenes[i] = parent2.Genes[i]
		}
	}
	return Chromosome{Genes: childGenes}
}

// IslandCrossover выполняет двухточечный кроссовер – оператор, характерный для островной модели.
func IslandCrossover(parent1, parent2 Chromosome) Chromosome {
	length := len(parent1.Genes)
	point1 := rand.Intn(length)
	point2 := rand.Intn(length)
	if point1 > point2 {
		point1, point2 = point2, point1
	}
	childGenes := make([]bool, length)
	for i := 0; i < length; i++ {
		if i >= point1 && i < point2 {
			childGenes[i] = parent2.Genes[i]
		} else {
			childGenes[i] = parent1.Genes[i]
		}
	}
	return Chromosome{Genes: childGenes}
}

// ClassicMutation реализует простую мутацию – случайное переключение битов.
func ClassicMutation(chrom *Chromosome, mutationRate float64) {
	for i := 0; i < len(chrom.Genes); i++ {
		if rand.Float64() < mutationRate {
			chrom.Genes[i] = !chrom.Genes[i]
		}
	}
}

// IslandMutation – более сложный оператор мутации для островной модели, выполняющий локальный поиск.
func IslandMutation(chrom *Chromosome, mutationRate float64, graph *Graph) {
	originalFitness := chrom.Fitness
	tempGenes := make([]bool, len(chrom.Genes))
	copy(tempGenes, chrom.Genes)

	for i := 0; i < len(chrom.Genes); i++ {
		if rand.Float64() < mutationRate {
			tempGenes[i] = !tempGenes[i]
		}
	}

	tempChrom := Chromosome{Genes: tempGenes}
	Repair(&tempChrom, graph)
	Evaluate(&tempChrom, graph)

	if tempChrom.Fitness >= originalFitness {
		chrom.Genes = tempChrom.Genes
		chrom.Fitness = tempChrom.Fitness
	}
}

// EvolvePopulation выполняет один шаг эволюции для текущей популяции.
func (ga *GeneticAlgorithm) EvolvePopulation() {
	newPopulation := make([]Chromosome, 0, ga.PopulationSize)
	tournamentSize := 3

	for len(newPopulation) < ga.PopulationSize {
		// Отбор родителей.
		parent1 := TournamentSelection(ga.Population, tournamentSize)
		parent2 := TournamentSelection(ga.Population, tournamentSize)
		var child Chromosome

		// Выбор оператора кроссовера в зависимости от EvolutionModel.
		switch ga.EvolutionModel {
		case Classic:
			if rand.Float64() < ga.CrossoverRate {
				child = ClassicCrossover(parent1, parent2)
			} else {
				child = parent1
			}
		case Island:
			if rand.Float64() < ga.CrossoverRate {
				child = IslandCrossover(parent1, parent2)
			} else {
				child = parent1
			}
		case Combined:
			if rand.Float64() < 0.5 {
				if rand.Float64() < ga.CrossoverRate {
					child = ClassicCrossover(parent1, parent2)
				} else {
					child = parent1
				}
			} else {
				if rand.Float64() < ga.CrossoverRate {
					child = IslandCrossover(parent1, parent2)
				} else {
					child = parent1
				}
			}
		}
		// Применение оператора мутации.
		switch ga.EvolutionModel {
		case Classic:
			ClassicMutation(&child, ga.MutationRate)
		case Island:
			IslandMutation(&child, ga.MutationRate, ga.Graph)
		case Combined:
			if rand.Float64() < 0.5 {
				ClassicMutation(&child, ga.MutationRate)
			} else {
				IslandMutation(&child, ga.MutationRate, ga.Graph)
			}
		}
		// Ремонт и оценка.
		Repair(&child, ga.Graph)
		Evaluate(&child, ga.Graph)
		newPopulation = append(newPopulation, child)
	}
	ga.Population = newPopulation
}

// GetBestChromosome возвращает лучшую хромосому из популяции.
func (ga *GeneticAlgorithm) GetBestChromosome() Chromosome {
	best := ga.Population[0]
	for _, chrom := range ga.Population {
		if chrom.Fitness > best.Fitness {
			best = chrom
		}
	}
	return best
}

// GetAllBestChromosomes возвращает все уникальные хромосомы с максимальным значением fitness из финальной популяции.
func (ga *GeneticAlgorithm) GetAllBestChromosomes() []Chromosome {
	bestFitness := 0
	for _, chrom := range ga.Population {
		if chrom.Fitness > bestFitness {
			bestFitness = chrom.Fitness
		}
	}

	uniqueMap := make(map[string]Chromosome)
	for _, chrom := range ga.Population {
		if chrom.Fitness == bestFitness {
			key := chromosomeKey(chrom)
			uniqueMap[key] = chrom
		}
	}

	result := make([]Chromosome, 0, len(uniqueMap))
	for _, chrom := range uniqueMap {
		result = append(result, chrom)
	}
	return result
}

// chromosomeKey формирует строковое представление хромосомы для устранения дубликатов.
func chromosomeKey(chrom Chromosome) string {
	var sb strings.Builder
	for _, gene := range chrom.Genes {
		if gene {
			sb.WriteByte('1')
		} else {
			sb.WriteByte('0')
		}
	}
	return sb.String()
}

// Run запускает генетический алгоритм и возвращает лучшее найденное решение.
func (ga *GeneticAlgorithm) Run() Chromosome {
	rand.Seed(time.Now().UnixNano())
	if ga.EvolutionModel == Island {
		islands := ga.initializeIslands()
		for gen := 0; gen < ga.Generations; gen++ {
			for i := 0; i < ga.NumIslands; i++ {
				islands[i] = evolveIsland(islands[i], ga)
			}
			if (gen+1)%ga.MigrationInterval == 0 {
				islands = migrate(islands)
			}
		}
		merged := mergeIslands(islands)
		ga.Population = merged
		return ga.GetBestChromosome()
	} else if ga.EvolutionModel == Combined {
		// Комбинированная модель: чередуем шаги классической и островной эволюции.
		ga.InitializePopulation()
		for gen := 0; gen < ga.Generations; gen++ {
			if gen%2 == 0 {
				ga.EvolvePopulation()
			} else {
				ga.EvolvePopulation()
				// Дополнительная диверсификация: усиливаем мутацию для части популяции.
				for i := 0; i < len(ga.Population)/10; i++ {
					idx := rand.Intn(len(ga.Population))
					for j := range ga.Population[idx].Genes {
						if rand.Float64() < ga.MutationRate*2 {
							ga.Population[idx].Genes[j] = !ga.Population[idx].Genes[j]
						}
					}
					Repair(&ga.Population[idx], ga.Graph)
					Evaluate(&ga.Population[idx], ga.Graph)
				}
			}
		}
		return ga.GetBestChromosome()
	} else {
		// Классическая модель
		for gen := 0; gen < ga.Generations; gen++ {
			ga.EvolvePopulation()
		}
		return ga.GetBestChromosome()
	}
}

// ----- Функции для островной модели -----

// initializeIslands делит начальную популяцию на заданное число островов.
func (ga *GeneticAlgorithm) initializeIslands() [][]Chromosome {
	islands := make([][]Chromosome, ga.NumIslands)
	islandSize := ga.PopulationSize / ga.NumIslands

	// Создаём популяцию для всех островов
	allChromosomes := make([]Chromosome, ga.PopulationSize)
	numEdges := len(ga.Graph.Edges)
	for i := 0; i < ga.PopulationSize; i++ {
		genes := make([]bool, numEdges)
		for j := 0; j < numEdges; j++ {
			genes[j] = rand.Float64() < 0.5
		}
		chrom := Chromosome{Genes: genes}
		Repair(&chrom, ga.Graph)
		Evaluate(&chrom, ga.Graph)
		allChromosomes[i] = chrom
	}

	// Перемешиваем и распределяем по островам
	rand.Shuffle(len(allChromosomes), func(i, j int) {
		allChromosomes[i], allChromosomes[j] = allChromosomes[j], allChromosomes[i]
	})

	for i := 0; i < ga.NumIslands; i++ {
		start := i * islandSize
		end := start + islandSize
		if i == ga.NumIslands-1 {
			end = len(allChromosomes)
		}
		islands[i] = allChromosomes[start:end]
	}
	return islands
}

// evolveIsland выполняет эволюцию для одного острова.
func evolveIsland(population []Chromosome, ga *GeneticAlgorithm) []Chromosome {
	newPopulation := make([]Chromosome, 0, len(population))
	tournamentSize := 3
	for len(newPopulation) < len(population) {
		parent1 := TournamentSelection(population, tournamentSize)
		parent2 := TournamentSelection(population, tournamentSize)
		var child Chromosome
		if rand.Float64() < ga.CrossoverRate {
			child = IslandCrossover(parent1, parent2)
		} else {
			child = parent1
		}
		IslandMutation(&child, ga.MutationRate, ga.Graph)
		Repair(&child, ga.Graph)
		Evaluate(&child, ga.Graph)
		newPopulation = append(newPopulation, child)
	}
	return newPopulation
}

// Migrate выполняет обмен лучшими особями между островами.
func migrate(islands [][]Chromosome) [][]Chromosome {
	numIslands := len(islands)
	bests := make([]Chromosome, numIslands)

	for i := 0; i < numIslands; i++ {
		bests[i] = islands[i][0]
		for _, chrom := range islands[i] {
			if chrom.Fitness > bests[i].Fitness {
				bests[i] = chrom
			}
		}
	}

	for i := 0; i < numIslands; i++ {
		sourceIndex := (i - 1 + numIslands) % numIslands
		minFitness := islands[i][0].Fitness
		minIdx := 0
		for j, chrom := range islands[i] {
			if chrom.Fitness < minFitness {
				minFitness = chrom.Fitness
				minIdx = j
			}
		}
		islands[i][minIdx] = bests[sourceIndex]
	}

	return islands
}

// mergeIslands объединяет популяции всех островов.
func mergeIslands(islands [][]Chromosome) []Chromosome {
	merged := []Chromosome{}
	for _, island := range islands {
		merged = append(merged, island...)
	}
	return merged
}
