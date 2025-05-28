package genetic

import (
	"math/rand"
	"sort"
	"strings"
)

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

// Evaluate вычисляет реальный размер паросочетания
func Evaluate(chrom *Chromosome, graph *Graph) {
	used := make(map[int]bool)
	count := 0

	// Проверяем только включенные ребра
	for i, gene := range chrom.Genes {
		if gene {
			edge := graph.Edges[i]
			if !used[edge.U] && !used[edge.V] {
				used[edge.U] = true
				used[edge.V] = true
				count++
			}
		}
	}

	chrom.Fitness = count
}

// EvaluateFast вычисляет функцию приспособленности – число ребер в допустимом паросочетании.
func EvaluateFast(chrom *Chromosome, graph *Graph) {
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

// RepairFast детерминированная версия Repair
func RepairFast(chrom *Chromosome, graph *Graph) {
	used := make(map[int]bool)

	// Перебираем рёбра в фиксированном порядке
	for idx, edge := range graph.Edges {
		if chrom.Genes[idx] {
			if !used[edge.U] && !used[edge.V] {
				used[edge.U] = true
				used[edge.V] = true
			} else {
				chrom.Genes[idx] = false
			}
		}
	}

	// Пересчитываем фитнес
	Evaluate(chrom, graph)
	//EvaluateFast(chrom, graph)
}

func (ga *Algorithm) getElitesFromIsland(island []Chromosome, n int) []Chromosome {
	sorted := make([]Chromosome, len(island))
	copy(sorted, island)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Fitness > sorted[j].Fitness
	})
	if n > len(sorted) {
		n = len(sorted)
	}
	return sorted[:n]
}

// ApplyAugmentingPath пытается найти одну увеличивающую цепь и «флипает» по ней.
// Если цепь не найдена — не меняет chrom.Genes.
func ApplyAugmentingPath(genes []bool, graph *Graph) {
	// Строим хромосому-времянку
	chrom := Chromosome{Genes: make([]bool, len(genes))}
	copy(chrom.Genes, genes)

	strat := AugmentingPathMutationStrategy{}
	strat.Mutate(&chrom, 1.0, graph)

	// Копируем результаты обратно
	copy(genes, chrom.Genes)
}

// MaxMatchingOld возвращает размер наибольшего паросочетания в неориентированном графе.
func MaxMatchingOld(graph *Graph) int {
	n := graph.NumVertices
	// match[v] = u, если v спарен с u, или -1
	match := make([]int, n)
	// parent[v] = w, откуда мы пришли в v при BFS
	parent := make([]int, n)
	// base[v] = базовая вершина текущего «цветка» (blossom)
	base := make([]int, n)
	inQueue := make([]bool, n)
	inPath := make([]bool, n)

	for i := 0; i < n; i++ {
		match[i] = -1
		base[i] = i
	}

	// строим матрицу смежности для быстрого доступа
	adj := make([][]bool, n)
	for i := range adj {
		adj[i] = make([]bool, n)
	}
	for _, e := range graph.Edges {
		adj[e.U][e.V] = true
		adj[e.V][e.U] = true
	}

	// findBase спускается по base[], сняв метки inPath, чтобы найти представителя
	findBase := func(v int) int {
		for base[v] != v {
			v = base[v]
		}
		return v
	}

	// augmentPath безопасно флипает пары вдоль найденного пути
	augmentPath := func(u, v int) {
		for {
			prevU := match[u]
			match[u] = v
			match[v] = u
			if prevU < 0 {
				break
			}
			// v ← parent[prevU], если он определён
			nextV := parent[prevU]
			u, v = prevU, nextV
		}
	}

	// bfs ищет увеличивающий путь от start, возвращает true, если найден
	bfs := func(start int) bool {
		// подготовка
		for i := 0; i < n; i++ {
			parent[i] = -1
			inQueue[i] = false
			inPath[i] = false
			base[i] = findBase(i)
		}
		queue := make([]int, 0, n)
		queue = append(queue, start)
		inQueue[start] = true

		for qi := 0; qi < len(queue); qi++ {
			u := queue[qi]
			for v := 0; v < n; v++ {
				// смежное и не в том же базовом blossom и не пытка самого себя
				if !adj[u][v] || base[u] == base[v] || match[u] == v {
					continue
				}
				// если обнаружен новый цветок
				if parent[v] != -1 {
					// поиск LCA (наименьшая общая база)
					x, y := u, v
					pathMark := make(map[int]bool)
					for {
						x = findBase(x)
						pathMark[x] = true
						if match[x] < 0 {
							break
						}
						x = parent[match[x]]
					}
					for {
						y = findBase(y)
						if pathMark[y] {
							break
						}
						if match[y] < 0 {
							break
						}
						y = parent[match[y]]
					}
					blossomBase := findBase(y)
					// помечаем все вершины цвета
					for i := 0; i < n; i++ {
						if pathMark[findBase(i)] {
							base[i] = blossomBase
						}
						if pathMark[findBase(i)] {
							inPath[i] = true
						}
					}
					// повторно добавляем в очередь все вершины blossom
					for i := 0; i < n; i++ {
						if inPath[findBase(i)] && !inQueue[i] {
							inQueue[i] = true
							queue = append(queue, i)
						}
					}
				} else {
					// обычный случай: расширяем дерево
					parent[v] = u
					if match[v] < 0 {
						// найден увеличивающий путь
						augmentPath(v, u)
						return true
					}
					// добавляем пару вершину в очередь
					next := match[v]
					if !inQueue[next] {
						inQueue[next] = true
						queue = append(queue, next)
					}
				}
			}
		}
		return false
	}

	// основной цикл: для каждой свободной вершины пытаемся найти путь
	result := 0
	for v := 0; v < n; v++ {
		if match[v] < 0 {
			if bfs(v) {
				result++
			}
		}
	}
	return result
}

func MaxMatching(graph *Graph) int {
	n := graph.NumVertices
	used := make([]bool, n)
	count := 0
	for _, e := range graph.Edges {
		if !used[e.U] && !used[e.V] {
			used[e.U] = true
			used[e.V] = true
			count++
		}
	}
	return count
}
