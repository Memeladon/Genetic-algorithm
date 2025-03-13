package genetic

// EvolutionModel задаёт режим эволюции алгоритма.
type EvolutionModel int

const (
	Classic EvolutionModel = iota
	Island
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

// Chromosome – хромосома, кодирующая решение в виде булевого среза.
// Значение true означает, что соответствующее ребро включено в паросочетание.
type Chromosome struct {
	Genes   []bool
	Fitness int
}

// GeneticAlgorithm хранит параметры алгоритма и текущее состояние популяции.
type GeneticAlgorithm struct {
	Graph             *Graph
	Population        []Chromosome
	EvolutionModel    EvolutionModel
	PopulationSize    int
	Generations       int
	MutationRate      float64
	CrossoverRate     float64
	NumIslands        int // для островной модели
	MigrationInterval int // число поколений между миграциями
}
