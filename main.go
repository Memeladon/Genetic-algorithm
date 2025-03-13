package main

import (
	genetic2 "Genetic-algorithm/backend/genetic"
	"fmt"
	"time"
)

func main() {
	edges := []genetic2.Edge{
		{U: 0, V: 1}, {U: 0, V: 2}, {U: 1, V: 2},
		{U: 1, V: 3}, {U: 2, V: 4}, {U: 3, V: 4},
		{U: 3, V: 5}, {U: 4, V: 5},
	}
	graph := genetic2.Graph{
		NumVertices: 6,
		Edges:       edges,
	}

	populationSize := 200
	generations := 200
	mutationRate := 0.03
	crossoverRate := 0.9
	numIslands := 4
	migrationInterval := 15

	ga := genetic2.NewGeneticAlgorithm(
		&graph,
		genetic2.Island,
		populationSize,
		generations,
		mutationRate,
		crossoverRate,
		numIslands,
		migrationInterval,
	)

	// Явная инициализация популяции перед замером времени
	ga.InitializePopulation()

	start := time.Now()
	best := ga.Run() // Теперь измеряется только время эволюции
	elapsed := time.Since(start)

	fmt.Printf("Лучшее паросочетание: %d ребер\n", best.Fitness)
	fmt.Println("Пример паросочетания:")
	printMatching(best, graph)

	allBest := ga.GetAllBestChromosomes()
	fmt.Printf("\nНайдено %d уникальных решений:\n", len(allBest))
	for i, matching := range allBest {
		fmt.Printf("Решение %d:\n", i+1)
		printMatching(matching, graph)
	}
	fmt.Printf("\nВремя выполнения: %s\n", elapsed)
}

func printMatching(chrom genetic2.Chromosome, graph genetic2.Graph) {
	for i, gene := range chrom.Genes {
		if gene {
			fmt.Printf("%d-%d ", graph.Edges[i].U, graph.Edges[i].V)
		}
	}
	fmt.Println()
}
