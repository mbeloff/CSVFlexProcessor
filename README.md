# CSV Processor

This tool processes CSV files from MarginFuel according to specific business rules. It requires a `Grid.csv` file and processes any files starting with "Flexfiles" in the same directory.

## Building the Executable

To build the executable, you'll need Go installed on your development machine. Then run:

```bash
# On Unix-like systems (Linux/macOS)
./build.sh

# On Windows, you can use Git Bash or WSL to run the build script
```

This will create executables in the `build` folder for both Windows and macOS.

## Using the Executable

1. Place the executable in the same directory as your CSV files
2. Ensure you have a `Grid.csv` file in the same directory
3. Place any files to be processed (starting with "Flexfiles") in the same directory
4. Run the executable:
   - On Windows: Double-click the `csv-flex-processor.exe` file
   - On macOS: Double-click the `csv-flex-processor_mac` file

The program will process all Flexfiles*.csv files and create corresponding processed_*.txt files in the same directory.

## Requirements

- Windows 10 or later
- macOS 10.13 or later
- Go 1.16 or later (for building from source)
- No additional software or dependencies required for running the executable 