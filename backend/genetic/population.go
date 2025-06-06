package genetic

import (
	"math/rand"
)

// Chromosome – хромосома, кодирующая решение в виде булевого среза.
// Значение true означает, что соответствующее ребро включено в паросочетание.
type Chromosome struct {
	Genes   []bool
	Fitness int
}

// InitializePopulation генерирует начальную популяцию
func (ga *Algorithm) InitializePopulation() {
	ga.Population = make([]Chromosome, ga.PopulationSize)
	for i := range ga.Population {
		ga.Population[i] = ga.GenerateChromosome()
	}
}

func (ga *Algorithm) InitializeIslands() [][]Chromosome {
	ga.InitializePopulation()
	return ga.DistributePopulation()
}

// ------------------------ Main ------------------------- //

// GenerateChromosome создаёт новую хромосому с корректным фитнесом
func (ga *Algorithm) GenerateChromosome() Chromosome {
	numEdges := len(ga.Graph.Edges)
	genes := make([]bool, numEdges)
	for j := 0; j < numEdges; j++ {
		genes[j] = rand.Float64() < 0.5
	}
	chrom := Chromosome{Genes: genes}
	RepairFast(&chrom, ga.Graph)
	Evaluate(&chrom, ga.Graph)
	return chrom
}

// ------------------------ Best ------------------------- //

// GetBestChromosome возвращает лучшую хромосому из популяции.
func (ga *Algorithm) GetBestChromosome() Chromosome {
	best := ga.Population[0]
	for _, chrom := range ga.Population {
		if chrom.Fitness > best.Fitness {
			best = chrom
		}
	}
	return best
}

// GetIslandBest возвращает лучшую хромосому на острове
func (ga *Algorithm) GetIslandBest(island []Chromosome) Chromosome {
	best := island[0]
	for _, chrom := range island {
		if chrom.Fitness > best.Fitness {
			best = chrom
		}
	}
	return best
}

// SetBestSoFar обновляет лучшее решение за всё время
func (ga *Algorithm) SetBestSoFar(chrom Chromosome) {
	edges := countValidMatchingEdges(chrom, ga.Graph)
	if edges > ga.BestSoFarEdges {
		ga.bestSoFar = chrom
		ga.BestSoFarEdges = edges
	}
}

// SetLocalBest обновляет лучший результат в текущей популяции
func (ga *Algorithm) SetLocalBest(chrom Chromosome) {
	ga.localBest = chrom
	ga.LocalBestEdges = countValidMatchingEdges(chrom, ga.Graph)
}

// GetBestSoFar возвращает глобально лучшее решение
func (ga *Algorithm) GetBestSoFar() Chromosome {
	return ga.bestSoFar
}

// GetAllBestChromosomes возвращает все уникальные хромосомы с максимальным значением fitness из финальной популяции.
func (ga *Algorithm) GetAllBestChromosomes() []Chromosome {
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

// ----------------------- Island ----------------------- //

// MigrateIslands выполняет обмен лучшими особями между островами.
func MigrateIslands(islands [][]Chromosome) [][]Chromosome {
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

// MergeIslands объединяет популяции всех островов.
func MergeIslands(islands [][]Chromosome) []Chromosome {
	merged := []Chromosome{}
	for _, island := range islands {
		merged = append(merged, island...)
	}
	return merged
}

// DistributePopulation распределяет популяцию по островам
func (ga *Algorithm) DistributePopulation() [][]Chromosome {
	islands := make([][]Chromosome, ga.NumIslands)
	islandSize := ga.PopulationSize / ga.NumIslands

	rand.Shuffle(ga.PopulationSize, func(i, j int) {
		ga.Population[i], ga.Population[j] = ga.Population[j], ga.Population[i]
	})

	for i := 0; i < ga.NumIslands; i++ {
		start := i * islandSize
		end := start + islandSize
		if i == ga.NumIslands-1 {
			end = ga.PopulationSize
		}
		islands[i] = ga.Population[start:end]
	}
	return islands
}
