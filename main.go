// Every Go program starts with a package declaration.
// `main` is special: it tells Go that this package should be compiled into an executable program.
package main

// The import statement lists all the packages our program needs.
import (
	"bufio"       // For reading text line-by-line efficiently
	"encoding/json" // For parsing JSON data
	"fmt"         // For formatted printing to the console
	"os"          // For interacting with the Operating System, like reading files
)

// The main function is the entry point of our program.
func main() {
	// 1. Open the log file.
	// os.Open returns two values: a pointer to the file and an error.
	file, err := os.Open("logs.jsonl")
	// 2. Always check for errors! This is a fundamental concept in Go.
	// If `err` is not `nil`, it means something went wrong opening the file.
	if err != nil {
		fmt.Println("Error opening file:", err)
		return // Exit the program if we can't open the file.
	}
	// 3. `defer` is a special Go keyword. It schedules the `file.Close()` call
	// to be run just before the `main` function exits. This ensures the file
	// is always closed, even if errors occur later.
	defer file.Close()

	// 4. Create a new "scanner" to read the file line by line.
	// This is more efficient than reading the whole file into memory at once.
	scanner := bufio.NewScanner(file)

	// 5. Loop through the file, one line at a time.
	// `scanner.Scan()` reads the next line and returns `true` if it was successful.
	for scanner.Scan() {
		// Get the text of the current line.
		line := scanner.Text()

		// 6. We need a place to store the parsed JSON. Since we don't know the exact
		// structure of every log line, we use a map. `map[string]interface{}` means
		// the keys are strings, and the values can be of any type (string, number, another map, etc.).
		var logEntry map[string]interface{}

		// 7. Parse the JSON. `json.Unmarshal` takes the JSON data as a slice of bytes
		// and a pointer to the variable where it should store the result.
		if err := json.Unmarshal([]byte(line), &logEntry); err != nil {
			// If parsing fails, print the raw line and an error message, then continue to the next line.
			fmt.Println("Error parsing JSON:", err)
			fmt.Println("Raw line:", line)
			continue
		}

		// 8. "Pretty-print" the parsed JSON. `json.MarshalIndent` converts the Go map
		// back into JSON, but with nice indentation for readability.
		prettyJSON, err := json.MarshalIndent(logEntry, "", "  ") // Use 2 spaces for indentation.
		if err != nil {
			fmt.Println("Error formatting JSON:", err)
			continue
		}

		// 9. Print the final, formatted result to the console.
		fmt.Println(string(prettyJSON))
		fmt.Println("---") // Add a separator for clarity
	}

	// Finally, check if the scanner itself encountered any errors during the process.
	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading file:", err)
	}
}