package main

import (
	"crypto/rand"
	"io/ioutil"
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
		body, err := ioutil.ReadAll(resp.Body)
		if err == nil {
			seed, err := strconv.ParseInt(string(body), 10, 64)
			if err == nil {
				return seed
			}
		}
	}

	// Fallback to crypto/rand for random number generation if HTTP request fails
	n, err := rand.Int(rand.Reader, big.NewInt(1<<62))
	if err != nil {
		log.Fatal("Failed to generate random seed:", err)
	}
	return n.Int64()
}

func generateUniqueRandomNumbers(seed int64, count int) []int {
	mrand.Seed(seed) // Seed the math/rand number generator
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

func flipCoin(seed int64, times int) bool {
	mrand.Seed(seed) // Seed the math/rand number generator
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

		jsonOutput := c.DefaultQuery("json", "0") // Check for json query parameter


		seed := fetchSeed()
		numbers := generateUniqueRandomNumbers(seed, count)
		threespot := generateUniqueRandomNumbers(seed, 3)
		alternate := generateUniqueRandomNumbers(seed, count)
		coinFlip := flipCoin(seed, 574673)
		if jsonOutput == "1" {
			c.JSON(http.StatusOK, gin.H{
				"seed":        seed,
				"numbers":     numbers,
				"threespot":   threespot,
				"alternate":   alternate,
				"coinFlip":    coinFlip,
			})
		} else {
			c.HTML(http.StatusOK, "index.tmpl", gin.H{
				"seed":       seed,
				"numbers":    numbers,
				"alternate":  alternate,
				"threespot":  threespot,
				"coinFlip":   coinFlip,
			})
		}
	})

	r.Run(":5000") // Listen and serve on 0.0.0.0:5000
}
