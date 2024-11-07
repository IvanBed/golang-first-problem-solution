package main

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func parseResponse(statsStr string) ([]float64, error) {
	statsStrItems := strings.Split(statsStr, ",")

	var resultValues []float64
	for _, item := range statsStrItems {
		number, err := strconv.ParseFloat(strings.TrimSpace(item), 64)
		if err != nil {
			return nil, err
		}
		resultValues = append(resultValues, number)
	}

	return resultValues, nil
}

func analyzeStats(statsSlice []float64) {

	loadAverage := statsSlice[0]
	totalMemory := statsSlice[1]
	usedMemory := statsSlice[2]
	totalDisk := statsSlice[3]
	usedDisk := statsSlice[4]
	totalBandwidth := statsSlice[5]
	usedBandwidth := statsSlice[6]

	messages := []string{}

	if loadAverage > 30 {
		messages = append(messages, fmt.Sprintf("Load Average is too high: %.0f", loadAverage))
	}

	memoryUsagePercent := int((usedMemory / totalMemory) * 100)
	if memoryUsagePercent > 80 {
		messages = append(messages, fmt.Sprintf("Memory usage too high: %d%%", memoryUsagePercent))
	}

	freeDiskSpaceMb := int((totalDisk - usedDisk) / (1024 * 1024))
	if usedDisk > totalDisk*0.90 {
		messages = append(messages, fmt.Sprintf("Free disk space is too low: %d Mb left", freeDiskSpaceMb))
	}

	bandwidthUsagePercent := int((usedBandwidth / totalBandwidth) * 100)
	freeBandwidthMbit := (int(totalBandwidth - usedBandwidth) / 1000000)
	if bandwidthUsagePercent > 90 {
		messages = append(messages, fmt.Sprintf("Network bandwidth usage high: %d Mbit/s available", freeBandwidthMbit))
	}

	if len(messages) > 0 {
		fmt.Println(strings.Join(messages, "\n"))
	} 
}

func getServerStats(url string) (string, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(body)), nil
}

func main() {
	url := "http://srv.msk01.gigacorp.local/_stats"

	errorCount := 0

	for {
		responseStr, err := getServerStats(url)
		if err != nil {
			fmt.Println("Error fetching stats:", err)
			errorCount++
			if errorCount >= 3 {
				fmt.Println("Unable to fetch server statistics")
				break
			}
			time.Sleep(2 * time.Second)
			continue
		}

		stats, _ := parseResponse(responseStr)

		analyzeStats(stats)
		errorCount = 0
		time.Sleep(5 * time.Second)
	}
}