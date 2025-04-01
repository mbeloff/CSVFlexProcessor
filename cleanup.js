"use strict";
const fs = require('fs');
const path = require('path');

function parseCSV(content) {
    // Split into lines and handle both \r\n and \n
    const lines = content.replace(/\r\n/g, '\n').split('\n');
    
    // Filter out empty lines
    return lines.filter(line => line.trim()).map(line => {
        // Handle quoted values if needed
        return line.split(',').map(value => value.trim());
    });
}

function processCSV(inputFile, flexFile) {
    try {
        console.log('\nReading input file:', inputFile);
        const inputContent = fs.readFileSync(inputFile, 'utf-8');
        console.log('Input file size:', inputContent.length, 'bytes');

        console.log('Parsing input CSV...');
        const inputLines = parseCSV(inputContent);
        const headers = inputLines[0];
        const inputData = inputLines.slice(1).map(line => {
            const row = {};
            headers.forEach((header, index) => {
                row[header] = line[index] || '';
            });
            return row;
        });
        console.log('Parsed input rows:', inputData.length);
        console.log('Input headers:', headers);

        console.log('\nReading flex file:', flexFile);
        const flexContent = fs.readFileSync(flexFile, 'utf-8');
        console.log('Flex file size:', flexContent.length, 'bytes');
        console.log('Flex file content:', flexContent);

        console.log('Parsing flex CSV...');
        const flexLines = parseCSV(flexContent);
        console.log('Parsed flex rows:', flexLines.length);

        if (!flexLines || flexLines.length < 2) {
            throw new Error('Flex data is empty or invalid');
        }

        // Convert flex data to the format we need
        const flexHeaders = flexLines[0];
        const flexEntries = [];
        
        for (let i = 1; i < flexLines.length; i++) {
            const row = flexLines[i];
            const rowHeader = row[0]; // First column is the row header
            
            for (let j = 1; j < row.length; j++) {
                const price = parseFloat(row[j]);
                if (!isNaN(price)) {
                    flexEntries.push({
                        rowHeader: rowHeader,
                        colHeader: flexHeaders[j],
                        price: price
                    });
                }
            }
        }
        console.log('Processed flex entries:', flexEntries.length);

        // Sort flex entries by price
        flexEntries.sort((a, b) => a.price - b.price);

        // Filter out unwanted vehicles and fromDay values
        const fromDayValues = ["0", "1", "14", "21", "29", "7"];
        const vehicleValues = [
            "Aventus 2-seater (AT)",
            "Mystery Machine 2",
            "Mystery Machine 2 Hightop",
            "Mystery Machine 3",
            "Budget Mini-Camper",
            "Grip 4x4",
        ];

        console.log('\nFiltering data...');
        let filteredData = inputData.filter(row => {
            if (!row) return false;
            return !fromDayValues.includes(row.FromDay) && !vehicleValues.includes(row.VehicleCode);
        });
        console.log('Filtered rows:', filteredData.length);

        // Process vehicle codes
        console.log('Processing vehicle codes...');
        filteredData = filteredData.map(row => ({
            ...row,
            VehicleCode: row.VehicleCode === "D5AWD Adventure Camper" ? "D5" :
                row.VehicleCode === "Desert Sands" ? "DSANDS" :
                    row.VehicleCode === "Johnny Feelgood" ? "JFG" :
                        row.VehicleCode
        }));

        // Add VehicleDesc column
        filteredData = filteredData.map(row => ({
            ...row,
            VehicleDesc: ""
        }));

        // Add FlexRate and Availability columns
        console.log('\nAdding flex rates and availability...');
        filteredData = filteredData.map(row => {
            const price = parseFloat(row.Price);
            let flexRate = "";
            if (!isNaN(price)) {
                const targetPrice = price * 0.75;
                let closestEntry = flexEntries[0];
                let minDiff = Math.abs(flexEntries[0].price - targetPrice);
                for (const entry of flexEntries) {
                    const diff = Math.abs(entry.price - targetPrice);
                    if (diff < minDiff) {
                        minDiff = diff;
                        closestEntry = entry;
                    }
                }
                flexRate = closestEntry.rowHeader + closestEntry.colHeader;
            }
            return {
                ...row,
                FlexRate: flexRate,
                Availability: "RQ"
            };
        });

        // Sort by dates
        console.log('\nSorting by dates...');
        filteredData.sort((a, b) => {
            const dateCompare = new Date(a.PickupDateTo).getTime() - new Date(b.PickupDateTo).getTime();
            if (dateCompare === 0) {
                return new Date(a.PickupDateFrom).getTime() - new Date(b.PickupDateFrom).getTime();
            }
            return dateCompare;
        });

        // Get earliest date
        const earliestDate = (filteredData[0]?.PickupDateFrom) || "";
        console.log('Earliest date:', earliestDate);

        // Prepare output data
        console.log('\nPreparing output data...');
        const outputData = [
            ["FIRST-BOOK-DATE", formatDate(earliestDate)],
            ...filteredData.map(row => {
                return [
                    row.PickupLocationCode,
                    row.VehicleCode,
                    "", // Empty VehicleDesc without extra quotes
                    formatDate(row.PickupDateFrom),
                    formatDate(row.PickupDateTo),
                    row.FlexRate,
                    row.Availability
                ];
            }),
            ["END OF FILE"]
        ];

        // Write output file
        console.log('\nWriting output file...');
        const outputContent = outputData.map(row => 
            row.map(cell => {
                if (cell === "END OF FILE") return cell;
                return `"${cell}"`;
            }).join(',')
        ).join('\n');

        const outputFile = path.join(path.dirname(inputFile), `processed_${path.basename(inputFile, '.csv')}.txt`);
        fs.writeFileSync(outputFile, outputContent);
        console.log(`Successfully saved output file: ${outputFile}`);

    } catch (error) {
        console.error('\nDetailed error information:');
        console.error('Error message:', error.message);
        console.error('Error stack:', error.stack);
        throw error; // Re-throw to be caught by main()
    }
}

// Format date as MM/DD/YYYY
function formatDate(dateStr) {
    if (!dateStr) {
        console.log('Empty date string received');
        return "";
    }
    
    try {
        // Try parsing the date string
        let date;
        // Check if the date is in DD/MM/YYYY format
        if (dateStr.includes('/')) {
            const [day, month, year] = dateStr.split('/');
            date = new Date(year, month - 1, day);
        } else {
            date = new Date(dateStr);
        }
        
        if (isNaN(date.getTime())) {
            console.log(`Invalid date received: ${dateStr}`);
            return "";
        }
        
        const month = String(date.getMonth() + 1).padStart(2, '0');
        const day = String(date.getDate()).padStart(2, '0');
        const year = date.getFullYear();
        const formatted = `${month}/${day}/${year}`;
        return formatted;
    } catch (e) {
        console.error(`Error parsing date: ${dateStr}`, e);
        return "";
    }
}

function main() {
    try {
        console.log('Starting CSV processing...');
        const currentDir = process.cwd();
        console.log(`Working directory: ${currentDir}`);
        
        const files = fs.readdirSync(currentDir);
        console.log('Files found in directory:', files);

        // Find flex file
        const flexFile = files.find(file => file === "Grid.csv");
        if (!flexFile) {
            throw new Error("Grid.csv not found in the current directory");
        }
        console.log('Found Grid.csv file');

        // Process all Flexfiles*.csv
        const flexFiles = files.filter(file => file.startsWith("Flexfiles") && file.endsWith(".csv"));
        if (flexFiles.length === 0) {
            throw new Error("No files starting with 'Flexfiles' found in the current directory");
        }
        console.log(`Found ${flexFiles.length} Flexfiles to process:`, flexFiles);

        for (const file of flexFiles) {
            console.log(`\nProcessing ${file}...`);
            processCSV(path.join(currentDir, file), path.join(currentDir, flexFile));
        }

        console.log('\nProcessing complete!');
    } catch (error) {
        console.error('\nError occurred:');
        console.error(error.message);
    }
    
    // Keep the window open and handle exit properly
    if (process.platform === 'win32') {
        console.log('Press any key to exit...');
        process.stdin.setRawMode(true);
        process.stdin.resume();
        process.stdin.on('data', () => {
            process.exit(0);
        });
    } else {
        process.exit(0);
    }
}

main();
