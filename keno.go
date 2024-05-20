package main

import (
	"crypto/rand"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"strconv"

	mrand "math/rand" // need to alias as we fallback to the crypto rand

	"github.com/gin-gonic/gin"
)

func fetchSeed() int64 {
	// First try to get a random seed from the remote hardware random generator
	resp, err := http.Get("https://rand.kroy.io")
	if err == nil && resp.StatusCode == http.StatusOK {
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err == nil {
			bigSeed := new(big.Int)
			bigSeed, success := bigSeed.SetString(string(body), 16)
			if success {
				seed := bigSeed.Int64()
				return seed
			} else {
				fmt.Println("Failed to parse hexadecimal seed as big integer")
			}
		} else {
			fmt.Println("Error reading response body:", err)
		}
	} else {
		fmt.Println("HTTP request failed:", err)
	}

	fmt.Println("Starting fallback seed generation")

	// Fallback to crypto/rand for random number generation if HTTP request fails
	n, err := rand.Int(rand.Reader, big.NewInt(1<<62))
	if err != nil {
		log.Fatal("Failed to generate random seed:", err)
	}
	return n.Int64()
}
func generateUniqueRandomNumbers(count int) []int {
	numbers := make([]int, 0, count)
	seen := make(map[int]bool)
	for len(numbers) < count {
		num := mrand.Intn(80) + 1 // Generate numbers between 1 and 80
		if !seen[num] {
			seen[num] = true
			numbers = append(numbers, num)
		}
	}
	return numbers
}

func flipCoin(times int) bool {
	headsCount := 0
	for i := 0; i < times; i++ {
		if mrand.Intn(2) == 0 { // 0 for heads, 1 for tails
			headsCount++
		}
	}
	return headsCount > times-headsCount
}

func main() {
	r := gin.Default()

	r.LoadHTMLGlob("/app/templates/*")

	r.GET("/", func(c *gin.Context) {
		countStr := c.DefaultQuery("count", "4")
		count, err := strconv.Atoi(countStr)
		if err != nil || count < 1 || count > 20 {
			count = 4 // Default to 4 if any error occurs
		}

		jsonOutput := c.DefaultQuery("json", "0") // Check for json

		seed := fetchSeed()
		fmt.Println("seed: ", seed)
		mrand.Seed(seed)

		numbers := generateUniqueRandomNumbers(count)
		threespot := generateUniqueRandomNumbers(3)
		alternate := generateUniqueRandomNumbers(count)
		// Generate a random number of flips between 1 and 1,000,000
		maxFlips := big.NewInt(1000000)
		n, err := rand.Int(rand.Reader, maxFlips)
		if err != nil {
			log.Fatalf("Failed to generate random number of flips: %v", err)
		}
		coinFlip := int(n.Int64()) + 1

		if jsonOutput == "1" {
			c.JSON(http.StatusOK, gin.H{
				"seed":      seed,
				"numbers":   numbers,
				"threespot": threespot,
				"alternate": alternate,
				"coinFlip":  coinFlip,
				"flipped":   n,
			})
		} else {
			c.HTML(http.StatusOK, "index.tmpl", gin.H{
				"seed":      seed,
				"numbers":   numbers,
				"alternate": alternate,
				"threespot": threespot,
				"coinFlip":  coinFlip,
				"flipped":   n,
			})
		}
	})

	r.Run(":5000") // Listen and serve on 0.0.0.0:5000
}
