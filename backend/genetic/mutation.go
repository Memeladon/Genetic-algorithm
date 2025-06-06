package genetic

import (
	"container/list"
	"math/rand"
)

// -------------------------------- Classic Mutation -------------------------------- //

type ClassicMutationStrategy struct{}

func (s *ClassicMutationStrategy) Mutate(chrom *Chromosome, rate float64, _ *Graph) {
	for i := range chrom.Genes {
		if rand.Float64() < rate {
			chrom.Genes[i] = !chrom.Genes[i]
		}
	}
}

func (s *ClassicMutationStrategy) GetName() string {
	return "Classic"
}

// -------------------------------- Island Mutation -------------------------------- //

type IslandMutationStrategy struct{}

func (s *IslandMutationStrategy) Mutate(chrom *Chromosome, rate float64, graph *Graph) {
	originalFitness := chrom.Fitness
	tempGenes := make([]bool, len(chrom.Genes))
	copy(tempGenes, chrom.Genes)

	for i := range tempGenes {
		if rand.Float64() < rate {
			tempGenes[i] = !tempGenes[i]
		}
	}

	tempChrom := Chromosome{Genes: tempGenes}
	RepairFast(&tempChrom, graph)
	//EvaluateFast(&tempChrom, graph)
	Evaluate(&tempChrom, graph)

	if tempChrom.Fitness >= originalFitness {
		chrom.Genes = tempGenes
		chrom.Fitness = tempChrom.Fitness
	}
}

func (s *IslandMutationStrategy) GetName() string {
	return "Island"
}

// --------------------------- Steady-State Mutation --------------------------- //

type SteadyStateMutationStrategy struct{}

func (s *SteadyStateMutationStrategy) Mutate(chrom *Chromosome, rate float64, graph *Graph) {
	for i := range chrom.Genes {
		if rand.Float64() < rate {
			chrom.Genes[i] = !chrom.Genes[i]
		}
	}
}

func (s *SteadyStateMutationStrategy) GetName() string {
	return "SteadyState"
}

// ----------------- Conflict-Adaptive Mutation ----------------- //

// ConflictAdaptiveMutationStrategy увеличивает вероятность мутации
// ребра пропорционально числу «конфликтов» (пересечений) в текущем паросочетании.
type ConflictAdaptiveMutationStrategy struct{}

func (s *ConflictAdaptiveMutationStrategy) Mutate(chrom *Chromosome, rate float64, graph *Graph) {
	n := len(chrom.Genes)
	if n == 0 {
		return
	}

	// Считаем для каждого выбранного ребра число конфликтов (других ребер, пересекающихся по вершине)
	conflicts := make([]int, n)
	maxC := 1
	for i, on := range chrom.Genes {
		if !on {
			continue
		}
		u, v := graph.Edges[i].U, graph.Edges[i].V
		for j, on2 := range chrom.Genes {
			if j == i || !on2 {
				continue
			}
			e2 := graph.Edges[j]
			if e2.U == u || e2.U == v || e2.V == u || e2.V == v {
				conflicts[i]++
			}
		}
		if conflicts[i] > maxC {
			maxC = conflicts[i]
		}
	}

	// Мутируем каждый ген с вероятностью rate * (1 + conflicts/maxC)
	for i := 0; i < n; i++ {
		p := rate * (1 + float64(conflicts[i])/float64(maxC))
		if rand.Float64() < p {
			chrom.Genes[i] = !chrom.Genes[i]
		}
	}
	// После мутации восстанавливаем допустимость
	RepairFast(chrom, graph)
	Evaluate(chrom, graph)
}

func (s *ConflictAdaptiveMutationStrategy) GetName() string {
	return "ConflictAdaptive"
}

// ----------------- Augmenting-Path Mutation ----------------- //

// AugmentingPathMutationStrategy ищет одну увеличивающую цепь и флипает все ребра на ней.
type AugmentingPathMutationStrategy struct{}

func (s *AugmentingPathMutationStrategy) Mutate(chrom *Chromosome, rate float64, graph *Graph) {
	// применяем с заданной базовой вероятностью
	if rand.Float64() > rate {
		return
	}

	// строим текущее паросочетание
	matched := make(map[int]bool)
	for i, on := range chrom.Genes {
		if on {
			e := graph.Edges[i]
			matched[e.U] = true
			matched[e.V] = true
		}
	}

	// Ищем самую короткую увеличивающую цепь: начинаем с любой свободной вершины U
	for start := 0; start < graph.NumVertices; start++ {
		if matched[start] {
			continue
		}
		if path := findAugmentingPath(start, graph, chrom); len(path) > 0 {
			// path — список индексов рёбер, по которым чередуемся
			for _, ei := range path {
				chrom.Genes[ei] = !chrom.Genes[ei]
			}
			break
		}
	}

	RepairFast(chrom, graph)
	Evaluate(chrom, graph)
}

// findAugmentingPath возвращает индексы рёбер в увеличивающей цепи (или nil), BFS по дуальному графу.
func findAugmentingPath(start int, graph *Graph, chrom *Chromosome) []int {
	// Вершины и состояние «на шаге»: очередь из (vertex, isMatchedPhase, pathEdges)
	type state struct {
		v       int
		matched bool  // true — ищем ребро НЕ в matching, false — ищем ребро в matching
		path    []int // по каким индексам рёбер пришли
	}
	visited := make(map[int]map[bool]bool)
	q := list.New()
	q.PushBack(state{v: start, matched: false, path: nil})
	visited[start] = map[bool]bool{false: true}

	for q.Len() > 0 {
		curr := q.Remove(q.Front()).(state)
		for ei, e := range graph.Edges {
			inMatching := chrom.Genes[ei]
			// фаза: если matched==false, ищем ребро не в matching, иначе — в matching
			if inMatching != curr.matched {
				continue
			}
			// текущая вершина должна совпадать с одним из концов ребра
			var nxt int
			if e.U == curr.v {
				nxt = e.V
			} else if e.V == curr.v {
				nxt = e.U
			} else {
				continue
			}
			// если дошли до свободной вершины и только что прошли matching-ребро — нашли augmenting path
			if !inMatching && !curr.matched && !matchedVertex(nxt, chrom, graph) {
				// путь + этот ei
				return append(curr.path, ei)
			}
			// иначе идём дальше, чередуя состояние matched-фазы
			if visited[nxt] == nil {
				visited[nxt] = make(map[bool]bool)
			}
			nextPhase := !curr.matched
			if visited[nxt][nextPhase] {
				continue
			}
			visited[nxt][nextPhase] = true
			newPath := append([]int{}, curr.path...)
			newPath = append(newPath, ei)
			q.PushBack(state{v: nxt, matched: nextPhase, path: newPath})
		}
	}
	return nil
}

// matchedVertex проверяет, занята ли v в текущем паросочетании
func matchedVertex(v int, chrom *Chromosome, graph *Graph) bool {
	for i, on := range chrom.Genes {
		if on {
			e := graph.Edges[i]
			if e.U == v || e.V == v {
				return true
			}
		}
	}
	return false
}

func (s *AugmentingPathMutationStrategy) GetName() string {
	return "AugmentingPath"
}

// ------------------------------- Combined Mutation ------------------------------- //

type CombinedMutationStrategy struct {
	Strategies []MutationStrategy
}

func (s *CombinedMutationStrategy) Mutate(chrom *Chromosome, rate float64, graph *Graph) {
	for _, strategy := range s.Strategies {
		strategy.Mutate(chrom, rate, graph)
	}
}

func (s *CombinedMutationStrategy) GetName() string {
	return "Combined"
}
