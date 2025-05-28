package genetic

// EvolutionModel задаёт режим эволюции алгоритма.
type EvolutionModel int

const (
	Classic EvolutionModel = iota
	Island
	SteadyState
	Memetic
	Combined
)

// Edge представляет ребро графа.
type Edge struct {
	U, V int
}

// Graph задаёт граф, для которого ищется наибольшее паросочетание.
type Graph struct {
	NumVertices int
	Edges       []Edge
}

type Config struct {
	UseFastRepair    bool
	UseCachedFitness bool
}

// Algorithm хранит параметры алгоритма и текущее состояние популяции.
type Algorithm struct {
	Graph             *Graph
	Population        []Chromosome
	EvolutionModel    EvolutionModel
	CrossoverStrategy CrossoverStrategy
	SelectionStrategy ElitismWrapper
	MutationStrategy  MutationStrategy
	PopulationSize    int
	Generations       int
	MutationRate      float64
	CrossoverRate     float64
	NumIslands        int // Для островной модели
	MigrationInterval int // Число поколений между миграциями

	bestSoFar Chromosome // Лучшая хромосома за всё время
}
