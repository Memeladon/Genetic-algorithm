# Genetic Algorithm Visualization Tool

A modern, interactive application for visualizing and experimenting with genetic algorithms. This tool provides a graphical user interface to run genetic algorithm simulations and visualize their results in real-time.

## Features

- Interactive GUI for controlling genetic algorithm parameters
- Real-time visualization of algorithm performance
- Customizable genetic algorithm parameters
- Graph plotting capabilities
- Cross-platform support

## Technologies Used

- **Go** (Golang) - Core programming language
- **Fyne** - GUI framework for creating cross-platform applications
- **Gonum Plot** - Scientific plotting library
- **Go Modules** - Dependency management

## Project Structure

```
.
├── backend/           # Core genetic algorithm implementation
│   ├── genetic/      # Genetic algorithm logic
│   ├── plot.go       # Plotting functionality
│   └── solver.go     # Algorithm solver implementation
├── frontend/         # GUI implementation
│   ├── app.go        # Main application setup
│   ├── controls_panel.go  # Control panel UI
│   └── graph_panel.go     # Graph visualization
├── main.go           # Application entry point
└── go.mod           # Go module definition
```

## How It Works

1. The application uses a genetic algorithm to solve optimization problems
2. Users can control various parameters through the GUI:
   - Population size
   - Mutation rate
   - Selection method
   - Other algorithm-specific parameters
3. The algorithm runs in the backend while providing real-time feedback
4. Results are visualized using interactive graphs
5. The application supports saving and loading different configurations

## Getting Started

### Prerequisites

- Go 1.24 or higher
- Required Go modules (automatically installed via `go mod`)

### Installation

1. Clone the repository:
   ```bash
   git clone [repository-url]
   ```

2. Navigate to the project directory:
   ```bash
   cd genetic-algorithm
   ```

3. Install dependencies:
   ```bash
   go mod download
   ```

4. Run the application:
   ```bash
   go run main.go
   ```

## Usage

1. Launch the application
2. Configure the genetic algorithm parameters in the control panel
3. Start the simulation
4. Observe the results in the graph panel
5. Adjust parameters as needed to optimize the results

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details. 