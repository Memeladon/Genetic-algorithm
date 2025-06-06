package genetic

// EvolutionModelStrategy определяет интерфейс для различных моделей эволюции
// в генетическом алгоритме
type EvolutionModelStrategy interface {
	// Evolve выполняет один шаг эволюции популяции
	Evolve(ga *Algorithm) error
	// GetRequiredStrategies возвращает список необходимых стратегий
	GetRequiredStrategies() []string
	// ValidateStrategies проверяет корректность настроек стратегий
	ValidateStrategies(ga *Algorithm) error
	// GetModelName возвращает название модели эволюции
	GetModelName() string
}

// EvolutionModel представляет тип модели эволюции
type EvolutionModel int

const (
	Classic EvolutionModel = iota
	Island
	SteadyState
	Memetic
	Combined
)

// Edge представляет ребро в графе
type Edge struct {
	U int // Начальная вершина
	V int // Конечная вершина
}

// Graph представляет граф для задачи о максимальном паросочетании
type Graph struct {
	NumVertices int    // Количество вершин
	Edges       []Edge // Список рёбер
}

// Config содержит основные параметры генетического алгоритма
type Config struct {
	UseFastRepair    bool    // Использовать быструю версию починки
	UseCachedFitness bool    // Использовать кэширование значений приспособленности
	PopulationSize   int     // Размер популяции
	MutationRate     float64 // Вероятность мутации
	Generations      int     // Максимальное число поколений
}

// EvolutionModelConfig содержит настройки модели эволюции
type EvolutionModelConfig struct {
	Model EvolutionModel // Тип модели эволюции
	// For Combined model, specify which strategies to use
	UseSelection   bool
	UseCrossover   bool
	UseMutation    bool
	UseLocalSearch bool
}

// Algorithm представляет основной класс генетического алгоритма
type Algorithm struct {
	Graph             *Graph
	Population        []Chromosome
	EvolutionModel    EvolutionModelStrategy
	CrossoverStrategy CrossoverStrategy
	SelectionStrategy ElitismWrapper
	MutationStrategy  MutationStrategy
	PopulationSize    int
	Generations       int
	MutationRate      float64
	CrossoverRate     float64
	NumIslands        int // Для островной модели
	MigrationInterval int // Число поколений между миграциями
	CurrentGeneration int // Текущее поколение

	bestSoFar             Chromosome // Лучшая хромосома за всё время
	localBest             Chromosome // Лучшая хромосома в текущей популяции
	BestSoFarEdges        int        // Число рёбер в лучшем паросочетании за всё время
	LocalBestEdges        int        // Число рёбер в лучшем паросочетании текущей популяции
	optimalSize           int        // Оптимальный размер паросочетания (вычисляется алгоритмом Эдмондса)
	useOptimalTermination bool       // Использовать ли оптимальное решение как условие остановки
	Logger                *Logger    // Логгер для вывода информации
}

// NewAlgorithm создаёт новый экземпляр генетического алгоритма
func NewAlgorithm(graph *Graph, config Config) *Algorithm {
	// Вычисляем оптимальное решение алгоритмом Эдмондса
	optimalSize := MaxMatchingOld(graph)

	return &Algorithm{
		Graph:                 graph,
		PopulationSize:        config.PopulationSize,
		MutationRate:          config.MutationRate,
		Generations:           config.Generations,
		optimalSize:           optimalSize,
		useOptimalTermination: true, // По умолчанию используем оптимальное решение
		Logger:                NewLogger(),
	}
}

// ShouldTerminate проверяет условия остановки алгоритма
func (ga *Algorithm) ShouldTerminate() bool {
	// Проверяем достижение оптимального решения
	if ga.useOptimalTermination && ga.bestSoFar.Fitness >= ga.optimalSize {
		return true
	}

	// Проверяем достижение максимального числа поколений
	if ga.CurrentGeneration >= ga.Generations {
		return true
	}

	return false
}

func (e EvolutionModel) String() string {
	switch e {
	case Classic:
		return "Classic"
	case Island:
		return "Island"
	case SteadyState:
		return "SteadyState"
	case Memetic:
		return "Memetic"
	case Combined:
		return "Combined"
	default:
		return "Unknown"
	}
}

// NewGeneticAlgorithm is defined in base.go

// SelectionStrategy определяет интерфейс для стратегий селекции
type SelectionStrategy interface {
	Select(population []Chromosome) Chromosome
	GetName() string
}

// CrossoverStrategy определяет интерфейс для стратегий скрещивания
type CrossoverStrategy interface {
	Crossover(parent1, parent2 Chromosome) Chromosome
	WithRate(rate float64) CrossoverStrategy
	GetName() string
}

// MutationStrategy определяет интерфейс для стратегий мутации
type MutationStrategy interface {
	Mutate(chromosome *Chromosome, rate float64, graph *Graph)
	GetName() string
}
