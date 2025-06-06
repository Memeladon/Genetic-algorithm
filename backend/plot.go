package backend

import (
	"Genetic-algorithm/backend/genetic"
	"errors"
	"image/color"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
)

// PlotResults создает графики на основе накопленных данных
func (s *GASolver) PlotResults() error {
	if len(s.Results) == 0 {
		return errors.New("нет данных для построения графиков")
	}

	// Создаем директорию для графиков
	dir := "plots_" + time.Now().Format("20060102-150405")
	if err := os.Mkdir(dir, 0755); err != nil {
		return err
	}

	// График 1: Время от размерности задачи
	if err := s.plotTimeVsSize(dir); err != nil {
		return err
	}

	// График 2: Фитнес и количество рёбер от итераций для каждого графа
	for _, graphName := range s.uniqueGraphNames() {
		if err := s.plotFitnessAndEdgesVsIterations(dir, graphName); err != nil {
			return err
		}
	}

	return nil
}

func (s *GASolver) plotTimeVsSize(dir string) error {
	p := plot.New()
	p.Title.Text = "Зависимость времени выполнения от размерности задачи"
	p.Title.TextStyle.Font.Size = 14
	p.X.Label.Text = "Размер задачи (Вершины + Рёбра)"
	p.Y.Label.Text = "Время выполнения (мс)"

	// Настраиваем сетку
	p.Add(plotter.NewGrid())

	// Группируем данные по алгоритмам
	data := make(map[string]plotter.XYs)
	for _, res := range s.Results {
		key := res.Algorithm
		size := res.GraphVertices + res.GraphEdges
		timeMs := res.TimeTaken.Milliseconds()

		// Debug print
		println("Algorithm:", key, "Graph:", res.GraphName, "Size:", size, "Time:", timeMs, "ms")

		// Jitter X if all points have the same size (for visibility)
		jitter := 0.0
		if len(data[key]) > 0 && int(data[key][len(data[key])-1].X) == size {
			jitter = float64(len(data[key])) * 0.1 // small offset
		}

		if _, ok := data[key]; !ok {
			data[key] = make(plotter.XYs, 0)
		}
		data[key] = append(data[key], plotter.XY{X: float64(size) + jitter, Y: float64(timeMs)})
	}

	// Добавляем линии для каждого алгоритма
	for algo, points := range data {
		line, err := plotter.NewLine(points)
		if err != nil {
			return err
		}
		line.Color = getAlgorithmColor(algo)
		line.Width = vg.Points(2)
		p.Add(line)
		p.Legend.Add(algo, line)
	}

	// Настраиваем легенду
	p.Legend.TextStyle.Font.Size = 10
	p.Legend.Padding = 5
	p.Legend.Top = true
	p.Legend.Left = true

	// Сохраняем график
	return savePlot(p, filepath.Join(dir, "time_vs_size.png"))
}

func (s *GASolver) plotFitnessAndEdgesVsIterations(dir, graphName string) error {
	// Создаем два графика: один для фитнеса, другой для количества рёбер
	p := plot.New()
	p.Title.Text = "Сходимость алгоритмов: " + graphName
	p.Title.TextStyle.Font.Size = 14
	p.X.Label.Text = "Поколение"
	p.Y.Label.Text = "Значение"

	// Настраиваем сетку
	p.Add(plotter.NewGrid())

	// Фильтруем результаты для данного графа
	results := s.resultsForGraph(graphName)
	if len(results) == 0 {
		return nil
	}

	// Добавляем линии для каждого алгоритма
	for _, res := range results {
		// Линия фитнеса
		fitnessPoints := make(plotter.XYs, len(res.FitnessHistory))
		for i, fitness := range res.FitnessHistory {
			fitnessPoints[i] = plotter.XY{X: float64(i), Y: float64(fitness)}
		}

		line, err := plotter.NewLine(fitnessPoints)
		if err != nil {
			return err
		}
		line.Color = getAlgorithmColor(res.Algorithm)
		line.Width = vg.Points(2)
		p.Add(line)
		p.Legend.Add(res.Algorithm+" (фитнес)", line)

		// Добавляем маркеры для лучших значений
		scatter, err := plotter.NewScatter(fitnessPoints)
		if err != nil {
			return err
		}
		scatter.Color = line.Color
		scatter.Shape = draw.CircleGlyph{}
		scatter.Radius = vg.Points(2)
		p.Add(scatter)
	}

	// Настраиваем легенду
	p.Legend.TextStyle.Font.Size = 10
	p.Legend.Padding = 5
	p.Legend.Top = true
	p.Legend.Left = true

	// Сохраняем график
	filename := "fitness_" + sanitizeFilename(graphName) + ".png"
	return savePlot(p, filepath.Join(dir, filename))
}

// Вспомогательные функции
func (s *GASolver) uniqueGraphNames() []string {
	seen := make(map[string]bool)
	names := []string{}
	for _, res := range s.Results {
		if !seen[res.GraphName] {
			seen[res.GraphName] = true
			names = append(names, res.GraphName)
		}
	}
	return names
}

func (s *GASolver) resultsForGraph(graphName string) []ExperimentResult {
	result := []ExperimentResult{}
	for _, res := range s.Results {
		if res.GraphName == graphName {
			result = append(result, res)
		}
	}
	return result
}

func getAlgorithmColor(algo string) color.Color {
	// Определяем цвета для разных алгоритмов
	colors := map[string]color.Color{
		"Classic":     color.RGBA{R: 31, G: 119, B: 180, A: 255},  // Синий
		"Island":      color.RGBA{R: 255, G: 127, B: 14, A: 255},  // Оранжевый
		"SteadyState": color.RGBA{R: 44, G: 160, B: 44, A: 255},   // Зеленый
		"Memetic":     color.RGBA{R: 214, G: 39, B: 40, A: 255},   // Красный
		"Combined":    color.RGBA{R: 148, G: 103, B: 189, A: 255}, // Фиолетовый
	}

	if color, ok := colors[algo]; ok {
		return color
	}
	// Возвращаем серый цвет для неизвестных алгоритмов
	return color.RGBA{R: 128, G: 128, B: 128, A: 255}
}

func savePlot(p *plot.Plot, filename string) error {
	// Увеличиваем размер графика для лучшей читаемости
	return p.Save(12*vg.Inch, 8*vg.Inch, filename)
}

func sanitizeFilename(name string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			return r
		}
		return '_'
	}, name)
}

// RunExperimentsVaryingSize runs experiments for a range of graph sizes and collects results for time_vs_size plot
func (s *GASolver) RunExperimentsVaryingSize(
	minSize, maxSize, step int,
	algorithms []Params,
) error {
	for size := minSize; size <= maxSize; size += step {
		for _, params := range algorithms {
			graph := generateRandomGraph(size)
			graphName := params.EvolutionModel.String() + "_size" + string(rune(size))
			start := time.Now()
			// Run the solver for this graph and params
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
				return err
			}
			ga.InitializePopulation()
			for gen := 1; gen <= ga.Generations; gen++ {
				ga.CurrentGeneration = gen
				if err := ga.EvolutionModel.Evolve(ga); err != nil {
					return err
				}
			}
			finalBest := ga.GetBestSoFar()
			finalBestValid := countValidMatchingEdges(finalBest, ga.Graph)
			result := ExperimentResult{
				GraphName:           graphName,
				Algorithm:           params.EvolutionModel.String(),
				GraphVertices:       graph.NumVertices,
				GraphEdges:          len(graph.Edges),
				TimeTaken:           time.Since(start),
				BestFitness:         finalBestValid,
				FitnessHistory:      []int{},
				BestMatchingEdges:   getValidMatchingEdges(finalBest, ga.Graph),
				BestChromosomeGenes: make([]bool, len(finalBest.Genes)),
			}
			copy(result.BestChromosomeGenes, finalBest.Genes)
			s.Results = append(s.Results, result)
		}
	}
	return nil
}

// generateRandomGraph creates a random graph with the given size (vertices + edges)
func generateRandomGraph(size int) genetic.Graph {
	// For simplicity, use size/2 vertices and size/2 edges
	vertices := size / 2
	edges := size - vertices

	g := genetic.Graph{
		NumVertices: vertices,
		Edges:       make([]genetic.Edge, 0, edges),
	}
	used := make(map[[2]int]bool)
	for len(g.Edges) < edges {
		u := randInt(0, vertices-1)
		v := randInt(0, vertices-1)
		if u != v && !used[[2]int{u, v}] && !used[[2]int{v, u}] {
			g.Edges = append(g.Edges, genetic.Edge{U: u, V: v})
			used[[2]int{u, v}] = true
		}
	}
	return g
}

func randInt(min, max int) int {
	return min + int(time.Now().UnixNano())%(max-min+1)
}
