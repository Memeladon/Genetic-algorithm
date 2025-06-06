package genetic

import (
	"math/rand"
)

// ---------------------- Классический одноточечный кроссовер ---------------------- //

type SinglePoint struct {
	rate float64
}

func (s *SinglePoint) Crossover(p1, p2 Chromosome) Chromosome {
	if rand.Float64() >= s.rate {
		return p1
	}

	length := len(p1.Genes)
	point := rand.Intn(length)
	childGenes := make([]bool, length)
	copy(childGenes[:point], p1.Genes[:point])
	copy(childGenes[point:], p2.Genes[point:])
	return Chromosome{Genes: childGenes}
}

func (s *SinglePoint) WithRate(rate float64) CrossoverStrategy {
	s.rate = rate
	return s
}

func (c *SinglePoint) GetName() string {
	return "SinglePoint"
}

// ------------------------ Островной двухточечный кроссовер ------------------------ //

type TwoPoint struct {
	rate float64
}

func (s *TwoPoint) Crossover(p1, p2 Chromosome) Chromosome {
	if rand.Float64() >= s.rate {
		return p1
	}

	length := len(p1.Genes)
	point1 := rand.Intn(length)
	point2 := rand.Intn(length)
	if point1 > point2 {
		point1, point2 = point2, point1
	}

	childGenes := make([]bool, length)
	copy(childGenes[:point1], p1.Genes[:point1])
	copy(childGenes[point1:point2], p2.Genes[point1:point2])
	copy(childGenes[point2:], p1.Genes[point2:])
	return Chromosome{Genes: childGenes}
}

func (s *TwoPoint) WithRate(rate float64) CrossoverStrategy {
	s.rate = rate
	return s
}

func (c *TwoPoint) GetName() string {
	return "TwoPoint"
}

// ---------------------- Комбинированная стратегия кроссовера ---------------------- //

type CombinedCrossover struct {
	strategies []CrossoverStrategy
	rate       float64
}

func (s *CombinedCrossover) Crossover(p1, p2 Chromosome) Chromosome {
	if rand.Float64() >= s.rate {
		return p1
	}

	strategy := s.strategies[rand.Intn(len(s.strategies))]
	return strategy.Crossover(p1, p2)
}

func (s *CombinedCrossover) WithRate(rate float64) CrossoverStrategy {
	s.rate = rate
	return s
}

func (c *CombinedCrossover) GetName() string {
	return "Combined"
}
