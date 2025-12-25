package generators

import (
	"discordBot/util"
	"math/rand"
	"strconv"
	"strings"
)

// GenerateRandomNumber returns a random number as a string with the specified length
func GenerateRandomNumber(input string) string {
	util.LoggerInit("generators", "randomnumber")
	if input == "" {
		return ""
	}
	length, err := strconv.Atoi(input)
	if err != nil || length <= 0 {
		return ""
	}
	// int64 can only safely represent up to 18 digits
	if length > 18 {
		return "" // or return an error message if you prefer
	}
	min := int64(1)
	for i := 1; i < length; i++ {
		min *= 10
	}
	max := min*10 - 1
	n := rand.Int63n(max-min+1) + min
	return strconv.FormatInt(n, 10)
}

// GenerateUsername takes user input and returns a realistic username
func GenerateUsername(input string) string {
	util.LoggerInit("generators", "username")
	base := strings.ToLower(strings.ReplaceAll(input, " ", ""))
	suffixes := []string{"123", "_x", "99", "_dev", "_pro", "_bot", "_01", "_real", "_tv", "_official"}
	prefix := ""
	if rand.Intn(2) == 0 {
		prefix = []string{"the", "real", "its", "mr", "ms", "dr", "pro", "x"}[rand.Intn(8)]
	}
	suffix := suffixes[rand.Intn(len(suffixes))]
	username := prefix + base + suffix
	return username
}
