package main

import (

	"bufio"
    "fmt"
    "log"
	"strings"	
	"strconv"
    "os"
	"time"
	"sync"
	"runtime/pprof"
)

type CityTemp struct {
	City string
	Temp float64
}

func readFile(lineChan chan<- CityTemp, wg *sync.WaitGroup) {
	defer wg.Done()
	f, err := os.Open("cities.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	reader := bufio.NewScanner(f)
	for reader.Scan() {
		line := reader.Text()
		parts := strings.Split(line, ";")
		city := parts[0]
        tempStr := parts[1]
		temp, err := strconv.ParseFloat(tempStr, 64)
        if err != nil {
            fmt.Println("Invalid temperature:", tempStr)
            continue
        }
		lineChan <- CityTemp{City: city, Temp: temp}
		
	}
	time.Sleep(5 * time.Second)
}

func worker(lineChan <-chan CityTemp, cityMap map[string][]float64, mapMutex *sync.Mutex, wg *sync.WaitGroup) {
	defer wg.Done()

	for ct := range lineChan {
		mapMutex.Lock()
		cityMap[ct.City] = append(cityMap[ct.City], ct.Temp)
		mapMutex.Unlock()
	}
}

func calculateMinMaxAndMean(myMap map[string][]float64) {
	newMap := make(map[string][]float64)
	for city, temps := range myMap {
		min := temps[0]
		max := temps[0]
		sum := 0.0
		if len(temps) == 0 {
			continue
		}

		for _, temp := range temps {
			if temp < min {
				min = temp
			}
			if temp > max {
				max = temp
			}
			sum += temp
		}

		mean := sum / float64(len(temps))
		fmt.Printf("%s -> Min: %.1f, Mean: %.1f, Max: %.1f\n", city, min, mean, max)
		newMap[city] = append(newMap[city], min)
		newMap[city] = append(newMap[city], max)
		newMap[city] = append(newMap[city], mean)
		
	}

}

func main() {
	// Create CPU profile file
	f, err := os.Create("cpu.prof")
	if err != nil {
		log.Fatal("could not create CPU profile: ", err)
	}
	defer f.Close()

	// Start CPU profiling
	if err := pprof.StartCPUProfile(f); err != nil {
		log.Fatal("could not start CPU profile: ", err)
	}
	defer pprof.StopCPUProfile()
	start := time.Now()

	readFileChan := make(chan CityTemp, 20)
	cityMap := make(map[string][]float64)	
	var readWg sync.WaitGroup
	var workerWg sync.WaitGroup
	var mapMutex sync.Mutex

	readWg.Add(1)
	go readFile(readFileChan, &readWg)

	numWorkers := 5
	for i := 0; i < numWorkers;i++ {
		workerWg.Add(1)
		go worker(readFileChan, cityMap, &mapMutex, &workerWg)
	}
	
	readWg.Wait()
	close(readFileChan)
	workerWg.Wait()


	
	 calculateMinMaxAndMean(cityMap)
	fmt.Println(cityMap)
	elapsed := time.Since(start)
	fmt.Println("Execution time:", elapsed)
}