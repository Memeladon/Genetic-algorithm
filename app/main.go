package main

import (
	"fmt"
	"time"

	"Genetic-algorithm/genetic"
)

func main() {
	// Пример графа:
	// Вершины: 0,1,2,3,4,5
	// Рёбра: (0,1), (0,2), (1,2), (1,3), (2,4), (3,4), (3,5), (4,5)
	edges := []genetic.Edge{
		{U: 0, V: 1},
		{U: 0, V: 2},
		{U: 1, V: 2},
		{U: 1, V: 3},
		{U: 2, V: 4},
		{U: 3, V: 4},
		{U: 3, V: 5},
		{U: 4, V: 5},
	}
	graph := genetic.Graph{
		NumVertices: 6,
		Edges:       edges,
	}

	// Параметры алгоритма.
	populationSize := 100
	generations := 100
	mutationRate := 0.05
	crossoverRate := 0.8
	numIslands := 4
	migrationInterval := 10

	// Выбор модели эволюции: Classic, Island или Combined.
	// Здесь выбран комбинированный режим.
	model := genetic.Island

	ga := genetic.NewGeneticAlgorithm(&graph, model, populationSize, generations, mutationRate, crossoverRate, numIslands, migrationInterval)

	start := time.Now()
	// Запуск алгоритма. При этом, для дальнейшего вывода всех найденных решений,
	// финальная популяция сохраняется в объекте ga.
	best := ga.Run()
	elapsed := time.Since(start)

	fmt.Printf("Лучшее найденное паросочетание: %d ребер\n", best.Fitness)
	fmt.Println("Пример паросочетания:")
	printMatching(best, graph)

	// Вывод всех уникальных наибольших паросочетаний, найденных в финальной популяции.
	allBest := ga.GetAllBestChromosomes()
	fmt.Printf("\nНайдено %d уникальных паросочетаний:\n", len(allBest))
	for i, matching := range allBest {
		fmt.Printf("Паросочетание %d:\n", i+1)
		printMatching(matching, graph)
		fmt.Println()
	}

	fmt.Printf("Время выполнения: %s\n", elapsed)
}

// printMatching выводит рёбра паросочетания, соответствующего данной хромосоме.
func printMatching(chrom genetic.Chromosome, graph genetic.Graph) {
	for i, selected := range chrom.Genes {
		if selected {
			edge := graph.Edges[i]
			fmt.Printf("(%d, %d) ", edge.U, edge.V)
		}
	}
	fmt.Println()
}
