package main

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"math/big"
	mrand "math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func fetchSeed() (int64, string) {
	// Attempt to fetch the seed from https://rand.kroy.io/
	seed, err := fetchSeedFromURL("https://rand.kroy.io/")
	if err == nil {
		return seed, "rand.kroy.io"
	}

	// If that fails, attempt to fetch the seed from random.org
	seed, err = fetchSeedFromURL("https://www.random.org/integers/?num=1&min=1&max=1000000&col=1&base=10&format=plain&rnd=new")
	if err == nil {
		return seed, "random.org"
	}

	// If both external sources fail, fall back to on-device generation
	return generateOnDeviceSeed(), "ondevice"
}

func fetchSeedFromURL(url string) (int64, error) {
	client := http.Client{
		Timeout: 4 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		fmt.Println("Failed to fetch seed from URL:", url, err)
		return 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Failed to read response body:", err)
		return 0, err
	}

	seedStr := strings.TrimSpace(string(body))
	seed, err := strconv.ParseInt(seedStr, 10, 64)
	if err != nil {
		fmt.Println("Failed to parse seed:", err)
		return 0, err
	}

	return seed, nil
}

func generateOnDeviceSeed() int64 {
	// Create a byte slice to hold the random bytes
	seedBytes := make([]byte, 8) // 8 bytes = 64 bits

	// Read random bytes from crypto/rand
	_, err := rand.Read(seedBytes)
	if err != nil {
		fmt.Println("Failed to generate random bytes:", err)
		return 0
	}

	// Convert the random bytes to an int64
	seed := int64(binary.LittleEndian.Uint64(seedBytes))
	return seed
}

func generateUniqueRandomNumbers(rng *mrand.Rand, count int) []int {
	numbers := make([]int, 0, count)
	seen := make(map[int]bool)
	for len(numbers) < count {
		num := rng.Intn(80) + 1 // Generate numbers between 1 and 80
		if !seen[num] {
			seen[num] = true
			numbers = append(numbers, num)
		}
	}
	return numbers
}

func flipCoin(rng *mrand.Rand, times int) bool {
	headsCount := 0
	for i := 0; i < times; i++ {
		if rng.Intn(2) == 0 { // 0 for heads, 1 for tails
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

		seed, source := fetchSeed()
		fmt.Println("seed: ", seed)

		// Create a new RNG instance using the fetched seed
		rng := mrand.New(mrand.NewSource(seed))

		numbers := generateUniqueRandomNumbers(rng, count)
		threespot := generateUniqueRandomNumbers(rng, 3)
		sixspot := generateUniqueRandomNumbers(rng, 6)
		alternate := generateUniqueRandomNumbers(rng, count)

		// Generate a random number of flips between 1 and 1,000,000 using crypto/rand
		maxFlips := big.NewInt(1000000)
		n, err := rand.Int(rand.Reader, maxFlips)
		if err != nil {
			log.Fatalf("Failed to generate random number of flips: %v", err)
		}

		two := big.NewInt(2)
		remainder := new(big.Int).Mod(n, two)
		flipped := remainder.Cmp(big.NewInt(0)) == 0

		if jsonOutput == "1" {
			c.JSON(http.StatusOK, gin.H{
				"seed":      seed,
				"numbers":   numbers,
				"threespot": threespot,
				"sixspot":   sixspot,
				"alternate": alternate,
				"coinFlip":  flipped,
				"flipped":   n,
				"source":    source,
			})
		} else {
			c.HTML(http.StatusOK, "index.tmpl", gin.H{
				"seed":      seed,
				"numbers":   numbers,
				"alternate": alternate,
				"threespot": threespot,
				"sixspot":   sixspot,
				"coinFlip":  flipped,
				"flipped":   n,
				"source":    source,
			})
		}
	})

	r.Run(":5000") // Listen and serve on 0.0.0.0:5000
}
