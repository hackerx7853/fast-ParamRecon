package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
)

type Config map[string][]string

func main() {
	// CLI Flags
	inputPath := flag.String("urls", "", "Path to the TXT file containing URLs")
	paramsPath := flag.String("params", "params.json", "Path to the JSON parameters file")
	flag.Parse()

	if *inputPath == "" {
		fmt.Println("\033[1;31m[!] Usage: ./scanner-visual -urls targets.txt -params params.json\033[0m")
		return
	}

	// Load Parameter Configuration
	confFile, err := os.ReadFile(*paramsPath)
	if err != nil {
		fmt.Printf("\033[1;31m[!] Error loading config: %v\033[0m\n", err)
		return
	}
	var vulnParams Config
	json.Unmarshal(confFile, &vulnParams)

	totalLines := countLines(*inputPath)
	fmt.Printf("\n\033[1;32mGOOD LUCK, hacker! Processing %d URLs...\033[0m\n", totalLines)

	startTime := time.Now()

	// Progress Bar Setup
	p := mpb.New(mpb.WithWidth(60))
	bar := p.AddBar(int64(totalLines),
		mpb.PrependDecorators(
			decor.Name("\033[1;92mFUZZ LOADING...\033[0m ", decor.WCSyncSpace),
			decor.Percentage(decor.WCSyncSpace),
		),
		mpb.AppendDecorators(
			decor.Elapsed(decor.ET_STYLE_GO, decor.WCSyncSpace),
		),
	)

	// Concurrency and Deduplication Setup
	results := make(map[string]map[string]struct{})
	for k := range vulnParams {
		results[k] = make(map[string]struct{})
	}

	var mu sync.Mutex
	var wg sync.WaitGroup
	linesChan := make(chan string, 50000)

	// Launch Workers
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for line := range linesChan {
				line = strings.TrimSpace(line)
				bar.Increment()
				
				if line == "" || !strings.Contains(line, "?") {
					continue
				}

				u, err := url.Parse(line)
				if err != nil {
					continue
				}

				queryParams := u.Query()
				
				for vuln, keywords := range vulnParams {
					for _, kw := range keywords {
						// Exact Key Match Logic
						if _, found := queryParams[kw]; found {
							mu.Lock()
							results[vuln][line] = struct{}{}
							mu.Unlock()
							break
						}
					}
				}
			}
		}()
	}

	// Efficient Stream Reading
	f, _ := os.Open(*inputPath)
	scanner := bufio.NewScanner(f)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		linesChan <- scanner.Text()
	}
	close(linesChan)
	wg.Wait()
	p.Wait()

	// Time Formatting (MM:SS)
	duration := time.Since(startTime)
	mins := int(duration.Minutes())
	secs := int(duration.Seconds()) % 60
	formattedTime := fmt.Sprintf("%02d:%02d", mins, secs)

	// Output Report
	fmt.Println("\n\033[1;33m[#] SCAN REPORT\033[0m")
	baseName := strings.TrimSuffix(*inputPath, ".txt")
	
	for vuln, urlsMap := range results {
		if len(urlsMap) > 0 {
			outputName := fmt.Sprintf("%s-%s.txt", baseName, vuln)
			outFile, _ := os.Create(outputName)
			for url := range urlsMap {
				outFile.WriteString(url + "\n")
			}
			outFile.Close()
			fmt.Printf("\033[1;34m[+]\033[0m \033[1;37m%s\033[0m: \033[1;32m%d\033[0m unique URLs -> \033[1;36m%s\033[0m\n", 
				strings.ToUpper(vuln), len(urlsMap), outputName)
		}
	}

	fmt.Printf("\n\033[1;35m--------------------------------------------\033[0m")
	fmt.Printf("\n\033[1;37m[!] TOTAL FUZZ TIME: %s\033[0m", formattedTime)
	if totalLines > 0 {
		throughput := float64(totalLines) / duration.Seconds()
		fmt.Printf("\n\033[1;37m[!] Processing Speed: %.0f URLs/sec\033[0m", throughput)
	}
	fmt.Printf("\n\033[1;35m--------------------------------------------\033[0m\n\n")
}

func countLines(path string) int {
	f, _ := os.Open(path)
	defer f.Close()
	s := bufio.NewScanner(f)
	c := 0
	for s.Scan() {
		c++
	}
	return c
}
