package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"
)

type FlexEntry struct {
	RowHeader string
	ColHeader string
	Price     float64
}

type InputRow struct {
	PickupLocationCode string
	VehicleCode        string
	PickupDateFrom     string
	PickupDateTo       string
	Price              string
	FromDay            string
	FlexRate           string
	Availability       string
}

func parseCSV(filename string) ([][]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("error opening file %s: %v", filename, err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("error reading CSV: %v", err)
	}

	// Filter out empty lines
	var filtered [][]string
	for _, record := range records {
		if len(record) > 0 && strings.TrimSpace(record[0]) != "" {
			filtered = append(filtered, record)
		}
	}
	return filtered, nil
}

func formatDate(dateStr string) string {
	if dateStr == "" {
		return ""
	}

	// Try different date formats
	formats := []string{
		"2006-01-02",
		"02/01/2006",
		"01/02/2006",
	}

	var t time.Time
	var err error
	for _, format := range formats {
		t, err = time.Parse(format, dateStr)
		if err == nil {
			return t.Format("01/02/2006")
		}
	}

	log.Printf("Warning: Could not parse date: %s", dateStr)
	return ""
}

func processCSV(inputFile, flexFile string) error {
	fmt.Printf("\nReading input file: %s\n", inputFile)
	inputRecords, err := parseCSV(inputFile)
	if err != nil {
		return fmt.Errorf("error parsing input CSV: %v", err)
	}

	if len(inputRecords) < 2 {
		return fmt.Errorf("input file is empty or invalid")
	}

	// Parse headers and data
	headers := inputRecords[0]
	var inputData []InputRow
	for _, record := range inputRecords[1:] {
		row := InputRow{}
		for i, value := range record {
			if i < len(headers) {
				switch headers[i] {
				case "PickupLocationCode":
					row.PickupLocationCode = value
				case "VehicleCode":
					row.VehicleCode = value
				case "PickupDateFrom":
					row.PickupDateFrom = value
				case "PickupDateTo":
					row.PickupDateTo = value
				case "Price":
					row.Price = value
				case "FromDay":
					row.FromDay = value
				}
			}
		}
		inputData = append(inputData, row)
	}

	fmt.Printf("Parsed input rows: %d\n", len(inputData))

	// Read and parse flex file
	fmt.Printf("\nReading flex file: %s\n", flexFile)
	flexRecords, err := parseCSV(flexFile)
	if err != nil {
		return fmt.Errorf("error parsing flex CSV: %v", err)
	}

	if len(flexRecords) < 2 {
		return fmt.Errorf("flex data is empty or invalid")
	}

	// Process flex data
	flexHeaders := flexRecords[0]
	var flexEntries []FlexEntry
	for i := 1; i < len(flexRecords); i++ {
		row := flexRecords[i]
		rowHeader := row[0]
		for j := 1; j < len(row); j++ {
			if price, err := strconv.ParseFloat(row[j], 64); err == nil {
				flexEntries = append(flexEntries, FlexEntry{
					RowHeader: rowHeader,
					ColHeader: flexHeaders[j],
					Price:     price,
				})
			}
		}
	}

	// Sort flex entries by price
	sort.Slice(flexEntries, func(i, j int) bool {
		return flexEntries[i].Price < flexEntries[j].Price
	})

	// Filter data
	fromDayValues := map[string]bool{"0": true, "1": true, "14": true, "21": true, "29": true, "7": true}
	vehicleValues := map[string]bool{
		"Aventus 2-seater (AT)":     true,
		"Mystery Machine 2":         true,
		"Mystery Machine 2 Hightop": true,
		"Mystery Machine 3":         true,
		"Budget Mini-Camper":        true,
		"Grip 4x4":                  true,
	}

	var filteredData []InputRow
	for _, row := range inputData {
		if !fromDayValues[row.FromDay] && !vehicleValues[row.VehicleCode] {
			// Process vehicle codes
			switch row.VehicleCode {
			case "D5AWD Adventure Camper":
				row.VehicleCode = "D5"
			case "Desert Sands":
				row.VehicleCode = "DSANDS"
			case "Johnny Feelgood":
				row.VehicleCode = "JFG"
			}
			filteredData = append(filteredData, row)
		}
	}

	// Add flex rates and availability
	for i := range filteredData {
		if price, err := strconv.ParseFloat(filteredData[i].Price, 64); err == nil {
			targetPrice := price * 0.75
			closestEntry := flexEntries[0]
			minDiff := abs(flexEntries[0].Price - targetPrice)

			for _, entry := range flexEntries {
				diff := abs(entry.Price - targetPrice)
				if diff < minDiff {
					minDiff = diff
					closestEntry = entry
				}
			}
			filteredData[i].FlexRate = closestEntry.RowHeader + closestEntry.ColHeader
		}
		filteredData[i].Availability = "RQ"
	}

	// Sort by dates
	sort.Slice(filteredData, func(i, j int) bool {
		dateI := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
		dateJ := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)

		if t, err := time.Parse("02/01/2006", filteredData[i].PickupDateTo); err == nil {
			dateI = t
		}
		if t, err := time.Parse("02/01/2006", filteredData[j].PickupDateTo); err == nil {
			dateJ = t
		}

		if dateI.Equal(dateJ) {
			dateI, _ = time.Parse("02/01/2006", filteredData[i].PickupDateFrom)
			dateJ, _ = time.Parse("02/01/2006", filteredData[j].PickupDateFrom)
		}
		return dateI.Before(dateJ)
	})

	// Prepare output
	outputFile := filepath.Join(filepath.Dir(inputFile), "processed_"+filepath.Base(inputFile[:len(inputFile)-4])+".txt")
	outputFileHandle, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("error creating output file: %v", err)
	}
	defer outputFileHandle.Close()

	// Write first book date
	if len(filteredData) > 0 {
		fmt.Fprintf(outputFileHandle, "\"%s\",\"%s\"\r\n", "FIRST-BOOK-DATE", formatDate(filteredData[0].PickupDateFrom))
	}

	// Write data rows
	for _, row := range filteredData {
		fmt.Fprintf(outputFileHandle, "\"%s\",\"%s\",\"\",\"%s\",\"%s\",\"%s\",\"%s\"\r\n",
			row.PickupLocationCode,
			row.VehicleCode,
			formatDate(row.PickupDateFrom),
			formatDate(row.PickupDateTo),
			row.FlexRate,
			row.Availability)
	}

	// Write end of file marker without quotes
	fmt.Fprintf(outputFileHandle, "END OF FILE\r\n")

	fmt.Printf("Successfully saved output file: %s\n", outputFile)
	return nil
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func main() {
	startTime := time.Now()
	fmt.Println("Starting CSV processing...")

	currentDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Error getting working directory: %v", err)
	}
	fmt.Printf("Working directory: %s\n", currentDir)

	// Find flex file
	flexFile := filepath.Join(currentDir, "Grid.csv")
	if _, err := os.Stat(flexFile); os.IsNotExist(err) {
		log.Fatal("Grid.csv not found in the current directory")
	}

	// Find all Flexfiles*.csv
	files, err := os.ReadDir(currentDir)
	if err != nil {
		log.Fatalf("Error reading directory: %v", err)
	}

	var flexFiles []string
	for _, file := range files {
		if strings.HasPrefix(file.Name(), "Flexfiles") && strings.HasSuffix(file.Name(), ".csv") {
			flexFiles = append(flexFiles, filepath.Join(currentDir, file.Name()))
		}
	}

	if len(flexFiles) == 0 {
		log.Fatal("No files starting with 'Flexfiles' found in the current directory")
	}

	fmt.Printf("Found %d Flexfiles to process\n", len(flexFiles))

	// Process each file
	for _, file := range flexFiles {
		fmt.Printf("\nProcessing %s...\n", file)
		if err := processCSV(file, flexFile); err != nil {
			log.Printf("Error processing %s: %v\n", file, err)
		}
	}

	processingTime := time.Since(startTime)
	fmt.Printf("\nProcessing complete! Total time: %v\n", processingTime)

	// Keep window open on Windows
	if runtime.GOOS == "windows" {
		fmt.Println("Press Enter to exit...")
		fmt.Scanln()
	}
}
