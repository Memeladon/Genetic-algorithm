package genetic

import (
	"math/rand"
	"sort"
)

type ElitismWrapper struct {
	Strategy  SelectionStrategy
	EliteSize int
}

// ----------------------- Турнирная селекция ----------------------- //

type TournamentSelectionStrategy struct {
	TournamentSize int
}

func (t *TournamentSelectionStrategy) Select(population []Chromosome) Chromosome {
	if len(population) == 0 {
		panic("empty population")
	}
	if t.TournamentSize < 1 {
		t.TournamentSize = 3
	}

	best := population[rand.Intn(len(population))]
	for i := 1; i < t.TournamentSize; i++ {
		contender := population[rand.Intn(len(population))]
		if contender.Fitness > best.Fitness {
			best = contender
		}
	}
	return best
}

func (t *TournamentSelectionStrategy) GetName() string {
	return "Tournament"
}

// -------------------- Селекция методом рулетки -------------------- //

type RouletteWheelSelectionStrategy struct{}

func (r *RouletteWheelSelectionStrategy) Select(population []Chromosome) Chromosome {
	if len(population) == 0 {
		panic("empty population")
	}

	totalFitness := 0
	for _, c := range population {
		totalFitness += c.Fitness
	}

	if totalFitness == 0 {
		// Возвращаем случайную хромосому
		return population[rand.Intn(len(population))]
	}

	randValue := rand.Intn(totalFitness)
	cumulative := 0
	for _, c := range population {
		cumulative += c.Fitness
		if cumulative > randValue {
			return c
		}
	}
	return population[len(population)-1]
}

func (r *RouletteWheelSelectionStrategy) GetName() string {
	return "RouletteWheel"
}

// --------------------- Ранжированная селекция --------------------- //

type RankSelectionStrategy struct{}

func (r *RankSelectionStrategy) Select(population []Chromosome) Chromosome {
	if len(population) == 0 {
		panic("empty population")
	}

	sorted := make([]Chromosome, len(population))
	copy(sorted, population)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Fitness > sorted[j].Fitness
	})

	// Линейное ранжирование: вероятность выбора пропорциональна рангу
	ranks := make([]int, len(sorted))
	for i := range ranks {
		ranks[i] = len(sorted) - i
	}

	total := len(sorted) * (len(sorted) + 1) / 2
	randValue := rand.Intn(total)

	cumulative := 0
	for i, rank := range ranks {
		cumulative += rank
		if cumulative >= randValue {
			return sorted[i]
		}
	}
	return sorted[len(sorted)-1]
}

func (r *RankSelectionStrategy) GetName() string {
	return "Rank"
}

// ---------------------------- Элитизм ---------------------------- //

func (ga *Algorithm) getElites() []Chromosome {
	sorted := ga.Population
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Fitness > sorted[j].Fitness // Сортировка по убыванию
	})
	return sorted[:ga.SelectionStrategy.EliteSize] // Возвращаем первых N элементов
}

func (s *ElitismWrapper) GetName() string {
	return "Elitism+" + s.Strategy.GetName()
}
