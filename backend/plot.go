package backend

import (
	"Genetic-algorithm/backend/genetic"
	"errors"
	"fmt"
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

// TableResult represents a single row in the results table
type TableResult struct {
	Model         string
	GraphName     string
	ExecutionTime time.Duration
	BestMatching  int
	BestGen       int // Generation where best result was found
	TotalGens     int // Total number of generations
}

// ResultsTable stores all results for table visualization
var ResultsTable []TableResult

// PlotResults creates plots based on accumulated data
func (s *GASolver) PlotResults() error {
	if len(s.Results) == 0 {
		return errors.New("no data for plotting")
	}

	// Create directory for plots
	dir := "plots_" + time.Now().Format("20060102-150405")
	if err := os.Mkdir(dir, 0755); err != nil {
		return err
	}

	// Update results table
	s.updateResultsTable()

	// Create HTML results table
	if err := s.createResultsTable(dir); err != nil {
		return err
	}

	// Plot 1: Time vs Problem Size
	if err := s.plotTimeVsSize(dir); err != nil {
		return err
	}

	// Plot 2: Objective Function vs Iterations
	for _, graphName := range s.uniqueGraphNames() {
		if err := s.plotObjectiveVsIterations(dir, graphName); err != nil {
			return err
		}
	}

	// Plot 3: Scatter Clouds
	if err := plotScatterClouds(dir); err != nil {
		return err
	}

	// Plot 4: Distributions
	if err := plotDistributions(dir); err != nil {
		return err
	}

	return nil
}

// updateResultsTable adds new results to the table
func (s *GASolver) updateResultsTable() {
	for _, res := range s.Results {
		// Format model name
		modelName := res.Algorithm
		if modelName == "Combined" {
			modelName = fmt.Sprintf("Combined (%s)", res.MutationStrategy)
		}

		// Find generation with best result
		bestGen := 0
		bestFitness := 0
		for gen, fitness := range res.FitnessHistory {
			if fitness > bestFitness {
				bestFitness = fitness
				bestGen = gen + 1
			}
		}

		// Create new entry
		tableResult := TableResult{
			Model:         modelName,
			GraphName:     res.GraphName,
			ExecutionTime: res.TimeTaken,
			BestMatching:  len(res.BestMatchingEdges),
			BestGen:       bestGen,
			TotalGens:     len(res.FitnessHistory),
		}

		// Add entry to table
		ResultsTable = append(ResultsTable, tableResult)
	}
}

// createResultsTable создает HTML файл с таблицей результатов
func (s *GASolver) createResultsTable(dir string) error {
	// Создаем HTML файл
	htmlFile := filepath.Join(dir, "results_table.html")
	file, err := os.Create(htmlFile)
	if err != nil {
		return err
	}
	defer file.Close()

	// Записываем HTML заголовок
	file.WriteString(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Результаты выполнения генетического алгоритма</title>
    <style>
        table {
            border-collapse: collapse;
            width: 100%;
            margin: 20px 0;
        }
        th, td {
            border: 1px solid #ddd;
            padding: 8px;
            text-align: left;
        }
        th {
            background-color: #f2f2f2;
            position: sticky;
            top: 0;
        }
        tr:nth-child(even) {
            background-color: #f9f9f9;
        }
        tr:hover {
            background-color: #f5f5f5;
        }
    </style>
</head>
<body>
    <h1>Результаты выполнения генетического алгоритма</h1>
    <table>
        <thead>
            <tr>
                <th>Модель</th>
                <th>Граф</th>
                <th>Время выполнения</th>
                <th>Лучшее паросочетание</th>
                <th>Поколение лучшего</th>
                <th>Всего поколений</th>
            </tr>
        </thead>
        <tbody>
`)

	// Записываем данные таблицы
	for _, result := range ResultsTable {
		file.WriteString(fmt.Sprintf(`
            <tr>
                <td>%s</td>
                <td>%s</td>
                <td>%v</td>
                <td>%d</td>
                <td>%d</td>
                <td>%d</td>
            </tr>`,
			result.Model,
			result.GraphName,
			result.ExecutionTime.Round(time.Millisecond),
			result.BestMatching,
			result.BestGen,
			result.TotalGens,
		))
	}

	// Записываем закрывающие теги
	file.WriteString(`
        </tbody>
    </table>
</body>
</html>`)

	return nil
}

func (s *GASolver) plotTimeVsSize(dir string) error {
	p := plot.New()
	p.Title.Text = "Зависимость времени выполнения от размерности задачи"
	p.Title.TextStyle.Font.Size = 14
	p.X.Label.Text = "Размер задачи (Вершины + Рёбра)"
	p.Y.Label.Text = "Время выполнения (мс)"
	p.Add(plotter.NewGrid())

	data := make(map[string]plotter.XYs)
	for _, res := range s.Results {
		key := res.Algorithm
		if key == "Combined" {
			key = fmt.Sprintf("Combined (%s)", res.MutationStrategy)
		}
		size := res.GraphVertices + res.GraphEdges
		timeMs := res.TimeTaken.Milliseconds()
		if _, ok := data[key]; !ok {
			data[key] = make(plotter.XYs, 0)
		}
		data[key] = append(data[key], plotter.XY{X: float64(size), Y: float64(timeMs)})
	}

	// Find min/max for X and Y for margin
	minX, maxX, minY, maxY := 1e9, -1e9, 1e9, -1e9
	for _, points := range data {
		for _, pt := range points {
			if pt.X < minX {
				minX = pt.X
			}
			if pt.X > maxX {
				maxX = pt.X
			}
			if pt.Y < minY {
				minY = pt.Y
			}
			if pt.Y > maxY {
				maxY = pt.Y
			}
		}
	}
	if minX == maxX {
		minX -= 1
		maxX += 1
	}
	if minY == maxY {
		minY -= 1
		maxY += 1
	}
	marginX := 0.05 * (maxX - minX)
	marginY := 0.05 * (maxY - minY)
	p.X.Min = minX - marginX
	p.X.Max = maxX + marginX
	p.Y.Min = minY - marginY
	p.Y.Max = maxY + marginY

	for algo, points := range data {
		scatter, err := plotter.NewScatter(points)
		if err != nil {
			return err
		}
		scatter.GlyphStyle.Color = getAlgorithmColor(algo)
		scatter.GlyphStyle.Radius = vg.Points(6)
		scatter.GlyphStyle.Shape = draw.CircleGlyph{}
		p.Add(scatter)
		p.Legend.Add(algo, scatter)

		// Add regression line (best fit line)
		if len(points) > 1 {
			sumX, sumY, sumXY, sumXX := 0.0, 0.0, 0.0, 0.0
			for _, pt := range points {
				sumX += pt.X
				sumY += pt.Y
				sumXY += pt.X * pt.Y
				sumXX += pt.X * pt.X
			}
			n := float64(len(points))
			slope := (n*sumXY - sumX*sumY) / (n*sumXX - sumX*sumX)
			intercept := (sumY - slope*sumX) / n
			minXl, maxXl := points[0].X, points[0].X
			for _, pt := range points {
				if pt.X < minXl {
					minXl = pt.X
				}
				if pt.X > maxXl {
					maxXl = pt.X
				}
			}
			linePoints := plotter.XYs{
				{X: minXl, Y: slope*minXl + intercept},
				{X: maxXl, Y: slope*maxXl + intercept},
			}
			line, err := plotter.NewLine(linePoints)
			if err == nil {
				line.Color = getAlgorithmColor(algo)
				line.Width = vg.Points(2)
				p.Add(line)
			}
		}
	}

	p.Legend.Top = true
	p.Legend.Left = false
	p.Legend.XOffs = 0
	p.Legend.YOffs = 0
	p.Legend.TextStyle.Font.Size = 12

	return savePlot(p, filepath.Join(dir, "time_vs_size.png"))
}

func (s *GASolver) plotObjectiveVsIterations(dir, graphName string) error {
	p := plot.New()
	p.Title.Text = fmt.Sprintf("Сходимость алгоритмов: %s", graphName)
	p.Title.TextStyle.Font.Size = 14
	p.X.Label.Text = "Поколение"
	p.Y.Label.Text = "Значение целевой функции"
	p.Add(plotter.NewGrid())

	results := s.resultsForGraph(graphName)
	if len(results) == 0 {
		return nil
	}

	// Group results by model
	modelRuns := make(map[string][][]float64)
	for _, res := range results {
		model := res.Algorithm
		if model == "Combined" {
			model = fmt.Sprintf("Combined (%s)", res.MutationStrategy)
		}
		fitness := make([]float64, len(res.FitnessHistory))
		for i, f := range res.FitnessHistory {
			fitness[i] = float64(f)
		}
		modelRuns[model] = append(modelRuns[model], fitness)
	}

	for model, runs := range modelRuns {
		// Plot all runs as faint lines
		for _, fitness := range runs {
			points := make(plotter.XYs, len(fitness))
			for i, f := range fitness {
				points[i] = plotter.XY{X: float64(i), Y: f}
			}
			line, err := plotter.NewLine(points)
			if err != nil {
				return err
			}
			c := getAlgorithmColor(model)
			line.Color = c
			line.Width = vg.Points(1)
			line.Dashes = []vg.Length{vg.Points(2), vg.Points(2)}
			p.Add(line)
		}
		// Plot mean curve as solid line and add to legend
		maxLen := 0
		for _, fitness := range runs {
			if len(fitness) > maxLen {
				maxLen = len(fitness)
			}
		}
		mean := make([]float64, maxLen)
		count := make([]int, maxLen)
		for _, fitness := range runs {
			for i, f := range fitness {
				mean[i] += f
				count[i]++
			}
		}
		for i := range mean {
			if count[i] > 0 {
				mean[i] /= float64(count[i])
			}
		}
		meanPoints := make(plotter.XYs, maxLen)
		for i, f := range mean {
			meanPoints[i] = plotter.XY{X: float64(i), Y: f}
		}
		meanLine, err := plotter.NewLine(meanPoints)
		if err != nil {
			return err
		}
		meanLine.Color = getAlgorithmColor(model)
		meanLine.Width = vg.Points(2)
		p.Add(meanLine)
		p.Legend.Add(model, meanLine)
	}

	p.Legend.Top = true
	p.Legend.Left = false
	p.Legend.XOffs = 0
	p.Legend.YOffs = 0
	p.Legend.TextStyle.Font.Size = 12

	filename := "objective_" + sanitizeFilename(graphName) + ".png"
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
		"Classic":     color.RGBA{R: 31, G: 119, B: 180, A: 255}, // Синий
		"Island":      color.RGBA{R: 255, G: 127, B: 14, A: 255}, // Оранжевый
		"SteadyState": color.RGBA{R: 44, G: 160, B: 44, A: 255},  // Зеленый
		"Memetic":     color.RGBA{R: 214, G: 39, B: 40, A: 255},  // Красный
	}

	// Для Combined модели используем разные цвета в зависимости от стратегии мутации
	if strings.HasPrefix(algo, "Combined") {
		if strings.Contains(algo, "Classic") {
			return color.RGBA{R: 148, G: 103, B: 189, A: 255} // Фиолетовый
		} else if strings.Contains(algo, "Island") {
			return color.RGBA{R: 140, G: 86, B: 75, A: 255} // Коричневый
		} else if strings.Contains(algo, "Steady-State") {
			return color.RGBA{R: 227, G: 119, B: 194, A: 255} // Розовый
		} else if strings.Contains(algo, "Conflict-Adaptive") {
			return color.RGBA{R: 127, G: 127, B: 127, A: 255} // Серый
		} else if strings.Contains(algo, "Augmenting-Path") {
			return color.RGBA{R: 188, G: 189, B: 34, A: 255} // Оливковый
		} else if strings.Contains(algo, "Combined") {
			return color.RGBA{R: 23, G: 190, B: 207, A: 255} // Бирюзовый
		}
	}

	if color, ok := colors[algo]; ok {
		return color
	}
	// Возвращаем серый цвет для неизвестных алгоритмов
	return color.RGBA{R: 128, G: 128, B: 128, A: 255}
}

func savePlot(p *plot.Plot, filename string) error {
	// Увеличиваем размер графика для лучшей читаемости
	return p.Save(14*vg.Inch, 8*vg.Inch, filename)
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

// Plot scatter clouds for all results: Time vs Generations
func plotScatterClouds(dir string) error {
	p := plot.New()
	p.Title.Text = "Execution Time vs Number of Generations"
	p.X.Label.Text = "Total Generations"
	p.Y.Label.Text = "Execution Time (ms)"
	p.Add(plotter.NewGrid())

	models := map[string][]plotter.XY{}
	for _, r := range ResultsTable {
		models[r.Model] = append(models[r.Model], plotter.XY{
			X: float64(r.TotalGens),
			Y: float64(r.ExecutionTime.Milliseconds()),
		})
	}

	// Set plot margins
	setPlotMargins(p, models)

	for model, points := range models {
		scatter, err := plotter.NewScatter(plotter.XYs(points))
		if err != nil {
			return err
		}
		scatter.GlyphStyle.Color = getAlgorithmColor(model)
		scatter.GlyphStyle.Radius = vg.Points(6)
		scatter.GlyphStyle.Shape = draw.CircleGlyph{}
		p.Add(scatter)
		p.Legend.Add(model, scatter)
	}

	setLegendStyle(p)
	return savePlot(p, filepath.Join(dir, "scatter_time_vs_gens.png"))
}

// Distribution plots: histogram of execution times per model
func plotDistributions(dir string) error {
	p := plot.New()
	p.Title.Text = "Execution Time Distribution by Model"
	p.X.Label.Text = "Execution Time (ms)"
	p.Y.Label.Text = "Frequency"
	p.Add(plotter.NewGrid())

	// Group by model
	modelTimes := make(map[string][]float64)
	for _, r := range ResultsTable {
		modelTimes[r.Model] = append(modelTimes[r.Model], float64(r.ExecutionTime.Milliseconds()))
	}

	// Set plot margins
	setPlotMargins(p, modelTimes)

	for model, times := range modelTimes {
		if len(times) == 0 {
			continue
		}
		vals := make(plotter.Values, len(times))
		copy(vals, times)
		h, err := plotter.NewHist(vals, 20)
		if err != nil {
			return err
		}
		h.Color = getAlgorithmColor(model)
		h.FillColor = getAlgorithmColor(model)
		h.LineStyle.Width = vg.Points(0.5)
		h.LineStyle.Color = getAlgorithmColor(model)
		p.Add(h)
		p.Legend.Add(model, h)
	}

	setLegendStyle(p)
	return savePlot(p, filepath.Join(dir, "distribution_time.png"))
}

// Helper function to set plot margins
func setPlotMargins(p *plot.Plot, data interface{}) {
	switch d := data.(type) {
	case map[string][]plotter.XY:
		minX, maxX, minY, maxY := 1e9, -1e9, 1e9, -1e9
		for _, points := range d {
			for _, pt := range points {
				if pt.X < minX {
					minX = pt.X
				}
				if pt.X > maxX {
					maxX = pt.X
				}
				if pt.Y < minY {
					minY = pt.Y
				}
				if pt.Y > maxY {
					maxY = pt.Y
				}
			}
		}
		if minX == maxX {
			minX -= 1
			maxX += 1
		}
		if minY == maxY {
			minY -= 1
			maxY += 1
		}
		marginX := 0.05 * (maxX - minX)
		marginY := 0.05 * (maxY - minY)
		p.X.Min = minX - marginX
		p.X.Max = maxX + marginX
		p.Y.Min = minY - marginY
		p.Y.Max = maxY + marginY

	case map[string][]float64:
		minY, maxY := 1e9, -1e9
		for _, values := range d {
			for _, v := range values {
				if v < minY {
					minY = v
				}
				if v > maxY {
					maxY = v
				}
			}
		}
		if minY == maxY {
			minY -= 1
			maxY += 1
		}
		marginY := 0.05 * (maxY - minY)
		p.X.Min = 0
		p.X.Max = maxY + 0.05*maxY
		p.Y.Min = minY - marginY
		p.Y.Max = maxY + marginY
	}
}

// Helper function to set legend style
func setLegendStyle(p *plot.Plot) {
	p.Legend.Top = true
	p.Legend.Left = false
	p.Legend.XOffs = 0
	p.Legend.YOffs = 0
	p.Legend.TextStyle.Font.Size = 12
}
