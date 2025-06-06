package genetic

import (
	"fmt"
	"time"
)

// Цвета для консольного вывода
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
	colorWhite  = "\033[37m"
)

// LogLevel определяет уровень важности сообщения
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	MILESTONE
	SUCCESS
	WARNING
	ERROR
)

// Logger представляет систему логирования для генетического алгоритма
type Logger struct {
	startTime time.Time
	lastBest  int
}

// NewLogger создаёт новый экземпляр логгера
func NewLogger() *Logger {
	return &Logger{
		startTime: time.Now(),
		lastBest:  -1,
	}
}

// log выводит сообщение с указанным уровнем и цветом
func (l *Logger) log(level LogLevel, format string, args ...interface{}) {
	var color, prefix string
	switch level {
	case DEBUG:
		color = colorWhite
		prefix = "[DEBUG]"
	case INFO:
		color = colorBlue
		prefix = "[INFO]"
	case MILESTONE:
		color = colorPurple
		prefix = "[MILESTONE]"
	case SUCCESS:
		color = colorGreen
		prefix = "[SUCCESS]"
	case WARNING:
		color = colorYellow
		prefix = "[WARNING]"
	case ERROR:
		color = colorRed
		prefix = "[ERROR]"
	}

	elapsed := time.Since(l.startTime).Round(time.Second)
	message := fmt.Sprintf(format, args...)
	fmt.Printf("%s%s %s %s%s\n", color, prefix, elapsed, message, colorReset)
}

// LogAlgorithmStart логирует начало работы алгоритма
func (l *Logger) LogAlgorithmStart(ga *Algorithm) {
	l.log(MILESTONE, "Запуск алгоритма: модель=%s, селекция=%s, кроссовер=%s, мутация=%s, размер популяции=%d, поколений=%d",
		ga.EvolutionModel.GetModelName(),
		ga.SelectionStrategy.GetName(),
		ga.CrossoverStrategy.GetName(),
		ga.MutationStrategy.GetName(),
		ga.PopulationSize, ga.Generations)
}

// countValidMatchingEdges возвращает количество рёбер в допустимом паросочетании для данной хромосомы
func countValidMatchingEdges(chrom Chromosome, graph *Graph) int {
	used := make(map[int]bool)
	count := 0
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
	return count
}

// LogGeneration логирует информацию о текущем поколении
func (l *Logger) LogGeneration(ga *Algorithm) {
	avgFitness := 0.0
	for _, chrom := range ga.Population {
		avgFitness += float64(countValidMatchingEdges(chrom, ga.Graph))
	}
	avgFitness /= float64(len(ga.Population))

	l.log(INFO, "Поколение %d: лучший фитнес=%d (рёбер=%d), средний фитнес=%.1f, глобально лучший фитнес=%d (рёбер=%d)",
		ga.CurrentGeneration,
		ga.localBest.Fitness, ga.LocalBestEdges,
		avgFitness,
		ga.bestSoFar.Fitness, ga.BestSoFarEdges)
}

// LogStrategyChange логирует изменение стратегии
func (l *Logger) LogStrategyChange(strategyType string, strategyName string) {
	l.log(INFO, "Изменение стратегии %s: %s", strategyType, strategyName)
}

// LogError логирует ошибку
func (l *Logger) LogError(err error) {
	l.log(ERROR, "%v", err)
}

// LogWarning логирует предупреждение
func (l *Logger) LogWarning(format string, args ...interface{}) {
	l.log(WARNING, format, args...)
}

// LogDebug логирует отладочную информацию
func (l *Logger) LogDebug(format string, args ...interface{}) {
	// DEBUG suppressed
}

// LogCompletion логирует завершение работы алгоритма
func (l *Logger) LogCompletion(ga *Algorithm) {
	if ga.bestSoFar.Fitness >= ga.optimalSize {
		l.log(SUCCESS, "Достигнуто оптимальное решение: %d (целевое значение: %d)",
			ga.bestSoFar.Fitness, ga.optimalSize)
	} else {
		l.log(INFO, "Алгоритм завершил работу. Лучший результат: %d (целевое значение: %d)",
			ga.bestSoFar.Fitness, ga.optimalSize)
	}
}

// LogMilestone logs a milestone message
func (l *Logger) LogMilestone(format string, args ...interface{}) {
	l.log(MILESTONE, format, args...)
}

// LogInfo logs an info message
func (l *Logger) LogInfo(format string, args ...interface{}) {
	l.log(INFO, format, args...)
}

// LogSuccess logs a success message
func (l *Logger) LogSuccess(format string, args ...interface{}) {
	l.log(SUCCESS, format, args...)
}
