package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"sort"
	"strings"
)

type Config struct {
	Labels              []string `json:"labels"`
	FileExtensions      []string `json:"fileExtensions"`
	SingleLineDelim     string   `json:"singleLineDelim"`
	MultiLineDelimStart string   `json:"multiLineDelimStart"`
	MultiLineDelimEnd   string   `json:"multiLineDelimEnd"`
}

type Result struct {
	label      string
	body       string
	filename   string
	lineNumber int
}

var (
	configFile = flag.String("config", "config.json", "config file (default: config.json)")
	directory  = flag.String("dir", ".", "directory to examine (default: .)")
)

func main() {
	flag.Parse()
	config := readConfig(*configFile)
	results := getResults(*directory, config)
	prettyPrint(results)
}

type ResultSlice []Result
func (a ResultSlice) Len() int           { return len(a) }
func (a ResultSlice) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ResultSlice) Less(i, j int) bool {
	
	filenameLengthCompare := len(a[i].filename) - len(a[j].filename)
	if filenameLengthCompare < 0 {
		return true
	} else if filenameLengthCompare > 0 {
		return false
	}

	filenameCompare := strings.Compare(a[i].filename, a[j].filename)
	if filenameCompare < 0 {
		return true
	} else if filenameCompare > 0 {
		return false
	} else {
		labelCompare := strings.Compare(a[i].label, a[j].label)
		if labelCompare < 0 {
			return true
		} else if labelCompare > 0 {
			return false
		} else {
			return a[i].lineNumber < a[j].lineNumber
		}
	}
	
}

func prettyPrint(results []Result) {
	sort.Sort(ResultSlice(results))

	currentFile := ""
	currentLabel := ""
	for _, result := range results {
		if currentFile != result.filename {
			currentFile = result.filename
			currentLabel = ""
			fmt.Println(currentFile)
		}
		if currentLabel != result.label {
			currentLabel = result.label
			fmt.Printf("\t%s\n", currentLabel)	
		}
		fmt.Printf("\t\t%d: %s\n", result.lineNumber, result.body)
	}
}

func readConfig(configFile string) Config {
	configFileReader, err := os.Open(configFile)
	if err != nil {
		panic(err)
	}
	decoder := json.NewDecoder(configFileReader)
	var config Config
	if err := decoder.Decode(&config); err != nil {
		panic(err)
	}
	return config
}

func getResults(directory string, config Config) []Result {
	var include func(string) bool = nil
	if len(config.FileExtensions) == 0 {
		include = func(filename string) bool {
			return true
		}
	} else {
		extensions := map[string]bool{}
		for _, extension := range config.FileExtensions {
			extensions[strings.Replace(extension, ".", "", -1)] = true
		}
		include = func(filename string) bool {
			filenamePieces := strings.Split(filename, ".")
			extension := filenamePieces[len(filenamePieces)-1]
			_, found := extensions[extension]
			return found
		}
	}
	return crawlAndCollect(directory, include, config)
}

func crawlAndCollect(directory string, include func(string) bool, config Config) []Result {
	results := []Result{}
	fInfo, error := ioutil.ReadDir(directory)

	if error != nil {
		panic(error)
	} else {
		directories := []string{}

		for _, file := range fInfo {
			filename := file.Name()
			relativePath := strings.Join([]string{directory, filename}, "/")

			if file.IsDir() {
				directories = append(directories, relativePath)
			} else if include(filename) {
				results = append(results, parseFileForResults(relativePath, config)...)
			}
		}

		for _, directory := range directories {
			results = append(results, crawlAndCollect(directory, include, config)...)
		}
	}
	return results
}

func parseFileForResults(filePath string, config Config) []Result {
	results := []Result{}
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading file (%s): %v", filePath, err)
	}
	defer file.Close()

	lineNumber := 1
	scanner := bufio.NewScanner(file)
	collecting := false
	collectString := ""
	for scanner.Scan() {
		line := scanner.Text()
		if !collecting && strings.Index(line, config.SingleLineDelim) >= 0 {
			results = append(results, parseStringForResults(line, filePath, lineNumber, config)...)
		} else if strings.Index(line, config.MultiLineDelimStart) >= 0 {
			collecting = true
			collectString = collectString + "\n" + line
		} else if collecting && strings.Index(line, config.MultiLineDelimEnd) >= 0 {
			collectString = collectString + "\n" + line
			results = append(results, parseStringForResults(collectString, filePath, lineNumber, config)...)
			collectString = ""
			collecting = false
		} else if collecting {
			collectString = collectString + "\n" + line
		}
		lineNumber++
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}

	return results
}

func parseStringForResults(possibleResults, filePath string, endingLineNumber int, config Config) []Result {
	results := []Result{}
	for _, label := range config.Labels {
		regExp := regexp.MustCompile(label + "(.*)")
		regExpResults := regExp.FindAllString(possibleResults, -1)
		for _, regexpResult := range regExpResults {
			regexpResult = strings.Replace(regexpResult, config.MultiLineDelimEnd, "", -1)

			offset := 0
			lines := strings.Split(possibleResults, regexpResult)
			if len(lines) > 0 {
				offset = strings.Count(lines[1], "\n")
			}
			results = append(results, Result {
					label : label,
					body : strings.TrimSpace(strings.Replace(regexpResult, label, "", 1)),
					filename : filePath,
					lineNumber : endingLineNumber - offset,
				})
		}
	}
	return results
}
