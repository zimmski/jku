package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Program must be called with <benchmark CSV file> as arguments.")

		os.Exit(1)
	}

	f, err := os.Open(os.Args[1])
	if err != nil {
		panic(err)
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.Comma = ';'

	records, err := r.ReadAll()
	if err != nil {
		panic(err)
	}

	// Drop the header.
	records = records[1:]

	fileGroups := [][][]string{}

	currentFile := ""
	group := [][]string{}
	for _, r := range records {
		if currentFile == "" || r[0] == currentFile {
			group = append(group, r)
		} else {
			fileGroups = append(fileGroups, group)
			fmt.Printf("Added file group with %d rows\n", len(group))

			group = [][]string{r}
		}

		currentFile = r[0]
	}
	fileGroups = append(fileGroups, group)
	fmt.Printf("Added file group with %d rows\n", len(group))

	fmt.Println("File groups:", len(fileGroups))

	fileRunGroups := [][][][]string{}

	for _, fg := range fileGroups {
		runGroups := [][][]string{}

		currentProgramRun := ""
		group = [][]string{}

		for _, r := range fg {
			if currentProgramRun == "" || r[2]+r[3] == currentProgramRun {
				group = append(group, r)
			} else {
				runGroups = append(runGroups, group)
				fmt.Printf("Added run group with %d rows\n", len(group))

				group = [][]string{r}
			}

			currentProgramRun = r[2] + r[3]
		}
		runGroups = append(runGroups, group)
		fmt.Printf("Added run group with %d rows\n", len(group))

		fmt.Println("Run groups:", len(runGroups))

		fileRunGroups = append(fileRunGroups, runGroups)
	}
	fmt.Println("File run groups:", len(fileRunGroups))

	fmt.Println("HERE COMES THE CSV DATA:")
	fmt.Println("File;Program;Number of CPUs;Average Time in Seconds;Average CPU Usage in Percentage;Average Minor Pagefaults;(absolute) speedup;(absolute) efficiency")

	for _, frg := range fileRunGroups {
		var sequentialTime float64

		for iFileRunGroup, rg := range frg {
			// Remove rows with worst and best time

			best := -1
			var bestTime float64
			worst := -1
			var worstTime float64
			for i, r := range rg {
				t, err := strconv.ParseFloat(r[4], 64)
				if err != nil {
					panic(err)
				}

				if best == -1 {
					best = i
					bestTime = t
					worst = i
					worstTime = t
				}

				if bestTime > t {
					best = i
					bestTime = t
				}
				if worstTime < t {
					worst = i
					worstTime = t
				}
			}

			_ = worst

			avgTime := float64(0.0)
			avgCPUUsage := float64(0)
			avgMinorPagefaults := float64(0)

			for i, r := range rg {
				if i == best || i == worst {
					continue
				}

				t, err := strconv.ParseFloat(r[4], 64)
				if err != nil {
					panic(err)
				}
				avgTime += t

				c, err := strconv.ParseInt(r[5], 10, 64)
				if err != nil {
					panic(err)
				}
				avgCPUUsage += float64(c)

				p, err := strconv.ParseInt(r[6], 10, 64)
				if err != nil {
					panic(err)
				}
				avgMinorPagefaults += float64(p)
			}

			count := float64(len(rg) - 2)

			avgTime /= count
			avgCPUUsage /= count
			avgMinorPagefaults /= count

			if iFileRunGroup == 0 {
				// The first run group must always be the sequential program.
				sequentialTime = avgTime
			}

			numberOfCPUs, err := strconv.ParseInt(rg[0][3], 10, 64)
			if err != nil {
				panic(err)
			}

			absoluteSpeedup := sequentialTime / avgTime
			absoluteEfficiency := absoluteSpeedup / float64(numberOfCPUs)

			fmt.Printf("%s;%s;%s;%0.7f;%0.7f;%0.7f;%0.7f;%0.7f\n", rg[0][0], rg[0][2], rg[0][3], avgTime, avgCPUUsage, avgMinorPagefaults, absoluteSpeedup, absoluteEfficiency)
		}
	}
}
