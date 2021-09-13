package main

import (
	"fmt"
	"github.com/emedvedev/enigma"
	"io/ioutil"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
)

// Default global parameters
var englishLetters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
var allPossibleRotors = []string{"I", "II", "V", "VI", "Beta", "Gamma"}

var enigmaConfig = struct {
	Reflector string
	Rings     []int
	Positions []string
	Rotors    []string
}{
	Reflector: "C-thin",
	Rings:     []int{1, 1, 1, 16},
	Positions: []string{"A", "A", "B", "Q"},
	Rotors:    []string{"Beta", "II", "IV", "III"},
}

// Function to compute index of coincidence
// input  -- text string
// output -- float value that gives
func computeIOC(text string) float64 {
	ioc := 0.0
	n := float64(len(text))
	// Loop runs for all alphabets
	for i := 0; i < 26; i++ {
		// count the occurrence of each alphabet
		temp := float64(strings.Count(text, string(englishLetters[i])))
		ioc += (temp * (temp - 1))
	}
	ioc = float64(ioc / (n * (n - 1)))
	return ioc
}

func ComputeTrigramScore(text, plugboardSetting string) float64 {

	totalScore := float64(0)
	tempPlugboard := createEnigmaPlugboard(plugboardSetting)
	enigmaOutput := setEnigmaAndDecode(text, tempPlugboard)

	PopulateTrigramScores()
	for i := 0; i < len(enigmaOutput)-3; i++ {
		// if there is a trigram, its score is added to total score
		totalScore += float64(trigramScores[enigmaOutput[i:i+3]])
	}

	return totalScore
}

// function to swap characters in Plugboard string
func swapCharacters(char1 string, char2 string, PlugboardSetting string) string {
	PlugboardSetting = strings.ReplaceAll(PlugboardSetting, char1, "*")
	PlugboardSetting = strings.ReplaceAll(PlugboardSetting, char2, char1)
	PlugboardSetting = strings.ReplaceAll(PlugboardSetting, "*", char2)
	return PlugboardSetting
}

// declare and populate english trigram scores
var trigramScores = make(map[string]float64)

// function to populate scores
func PopulateTrigramScores() {
	// read english trigram file
	file, err := os.Open("english_trigrams.txt")
	if err != nil {
		log.Fatal(err)
	}
	dataBytes, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatal(err)
	}
	fileContents := string(dataBytes)
	// split each line. each one will contain one trigram and its frequency
	trigramFreqPairs := strings.Split(fileContents, "\n")
	// to compute sample set size
	sum := float64(0)

	// iterates line by line
	for i := 0; i < len(trigramFreqPairs)-1; i++ {
		// split each line to trigram and freq pairs
		temp := strings.Split(trigramFreqPairs[i], " ")
		// convert frequency into number
		freq, err := strconv.Atoi(temp[1])
		if err != nil {
			log.Fatal(err)
		}
		// insert into map
		trigramScores[temp[0]] = float64(freq)
		// summation to find sample set size
		sum += float64(freq)
	}
	// Normalise freq values
	for i := 0; i < len(trigramFreqPairs)-1; i++ {
		var k = strings.Split(trigramFreqPairs[i], " ")
		trigramScores[k[0]] = math.Log(trigramScores[k[0]] / sum)
	}
}

// Function to load settings into enigma and encrypt the string
// Input  -- text string to encrypt and plugboard settings
// output -- encrypted string
func setEnigmaAndDecode(inputText string, plugboardSettings []string) string {
	config := make([]enigma.RotorConfig, len(enigmaConfig.Rotors))

	for index, rotor := range enigmaConfig.Rotors {
		ring := enigmaConfig.Rings[index]
		value := enigmaConfig.Positions[index][0]
		config[index] = enigma.RotorConfig{ID: rotor, Start: value, Ring: ring}
	}
	e := enigma.NewEnigma(config, enigmaConfig.Reflector, plugboardSettings)
	encoded := e.EncodeString(inputText)

	return string(encoded)
}

// function to get swapped values in pairs
func createEnigmaPlugboard(currentPlugboard string) []string {
	var swappedAlphabets []string
	DefaultAlphabets := englishLetters
	for i := 0; i < len(currentPlugboard); i++ {
		// Check if swapped pair is in list or if the current plug board has no swaps
		if currentPlugboard[i] != DefaultAlphabets[i] && string(currentPlugboard[i]) != "-" && string(DefaultAlphabets[i]) != "-" {
			// if not add to list
			swappedAlphabets = append(swappedAlphabets, string(DefaultAlphabets[i])+string(currentPlugboard[i]))
			l := string(currentPlugboard[i])
			m := string(DefaultAlphabets[i])
			// replace swapped pair to arbitrary value, so it is skipped in next iteration
			currentPlugboard = strings.ReplaceAll(currentPlugboard, l, "-")
			DefaultAlphabets = strings.ReplaceAll(DefaultAlphabets, l, "-")
			currentPlugboard = strings.ReplaceAll(currentPlugboard, m, "-")
			DefaultAlphabets = strings.ReplaceAll(DefaultAlphabets, m, "-")
		}
	}
	return swappedAlphabets
}

func HillClimb(text string) string {

	var tempPlugboardSettings string
	var currentBestPlugboardSettings string
	var currentPlugboard string
	var bestPlugboard string
	var tempPlugboard []string
	var tempDecrypted string
	bestPlugboard = englishLetters
	currentPlugboard = englishLetters
	maxIOC := float64(0)
	total := float64(0)
	count := float64(0)

	for i := 0; i < 26; i++ {
		currentPlugboard = bestPlugboard
		IOC := float64(0)
		for j := i + 1; j < 26; j++ {
			// If letter is already swapped
			// check which if the four permutations
			var enigmaPlugboard []string
			var decrypted string
			if string(currentPlugboard[j]) != string(englishLetters[j]) {

				enigmaPlugboard = createEnigmaPlugboard(englishLetters)
				decrypted = setEnigmaAndDecode(text, enigmaPlugboard)
				IOC = computeIOC(decrypted)
				if total == 0 {
					total = IOC
					count++
				} else if IOC <= total/count {
					continue
				} else if IOC > total/count {
					total += IOC
					count++
				}

				// swap them back to initial positions
				tempPlugboardSettings = swapCharacters(string(englishLetters[j]), string(currentPlugboard[j]), currentPlugboard)
				tempPlugboardSettings = swapCharacters(string(englishLetters[i]), string(currentPlugboard[i]), tempPlugboardSettings)

				// check for 4 alternate possibilities in hill climbing
				tempPlugboardSettings1 := swapCharacters(string(englishLetters[i]), string(currentPlugboard[i]), tempPlugboardSettings)
				tempPlugboard = createEnigmaPlugboard(tempPlugboardSettings1)
				tempDecrypted = setEnigmaAndDecode(text, tempPlugboard)
				IOC1 := computeIOC(tempDecrypted)

				tempPlugboardSettings2 := swapCharacters(string(englishLetters[j]), string(currentPlugboard[j]), tempPlugboardSettings)
				tempPlugboard = createEnigmaPlugboard(tempPlugboardSettings2)
				tempDecrypted = setEnigmaAndDecode(text, tempPlugboard)
				IOC2 := computeIOC(tempDecrypted)

				tempPlugboardSettings3 := swapCharacters(string(englishLetters[i]), string(currentPlugboard[i]), tempPlugboardSettings)
				tempPlugboard = createEnigmaPlugboard(tempPlugboardSettings3)
				tempDecrypted = setEnigmaAndDecode(text, tempPlugboard)
				IOC3 := computeIOC(tempDecrypted)

				tempPlugboardSettings4 := swapCharacters(string(englishLetters[j]), string(currentPlugboard[j]), tempPlugboardSettings)
				tempPlugboard = createEnigmaPlugboard(tempPlugboardSettings4)
				tempDecrypted = setEnigmaAndDecode(text, tempPlugboard)
				IOC4 := computeIOC(tempDecrypted)

				// Choose the Highest IOC as temp settings for Plugboard
				if IOC1 > IOC2 && IOC1 > IOC3 && IOC1 > IOC4 {
					IOC = IOC1
					tempPlugboardSettings = tempPlugboardSettings1
				} else if IOC2 > IOC1 && IOC2 > IOC3 && IOC2 > IOC4 {
					IOC = IOC2
					tempPlugboardSettings = tempPlugboardSettings2
				} else if IOC3 > IOC2 && IOC3 > IOC1 && IOC3 > IOC4 {
					IOC = IOC3
					tempPlugboardSettings = tempPlugboardSettings3
				} else {
					IOC = IOC4
					tempPlugboardSettings = tempPlugboardSettings4
				}

			} else {
				tempPlugboardSettings = swapCharacters(string(englishLetters[i]), string(currentPlugboard[j]), currentPlugboard)
				enigmaPlugboard = createEnigmaPlugboard(tempPlugboardSettings)
				decrypted = setEnigmaAndDecode(text, enigmaPlugboard)
				IOC = computeIOC(decrypted)
			}

			// if IOC of this instance is highest, preserve it as best settings
			if IOC > maxIOC {
				maxIOC = IOC
				currentBestPlugboardSettings = tempPlugboardSettings
			}

		}
		bestPlugboard = currentBestPlugboardSettings
	}
	return bestPlugboard
}

func ComputePlugboardScores(text string) (string, float64) {
	iocPlugboardSetting := HillClimb(text)
	trigramPlugboardScore := float64(ComputeTrigramScore(text, iocPlugboardSetting))
	return iocPlugboardSetting, trigramPlugboardScore
}

func main() {
	bestScore := math.Inf(-1)
	var rotor1Setting string
	var rotor2Setting string
	var position1Setting string
	var position2Setting string
	var bestPlugboardSetting string

	file, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	dataBytes, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatal(err)
	}
	cipherText := string(dataBytes)

	for i := 0; i < len(allPossibleRotors); i++ {

		for j := 0; j < len(allPossibleRotors); j++ {
			if i == j {
				continue
			}
			enigmaConfig.Rotors[0] = allPossibleRotors[i]
			enigmaConfig.Rotors[1] = allPossibleRotors[j]
			fmt.Println("Iteration: "+enigmaConfig.Rotors[0], enigmaConfig.Rotors[1])
			for a := 0; a < 26; a++ {
				for b := 0; b < 26; b++ {
					enigmaConfig.Positions[0] = string(a + 65)
					enigmaConfig.Positions[1] = string(b + 65)
					bestPlugboardInstance, trigramScore := ComputePlugboardScores(cipherText)
					if trigramScore > bestScore {
						bestScore = trigramScore
						bestPlugboardSetting = bestPlugboardInstance
						rotor1Setting = enigmaConfig.Rotors[0]
						rotor2Setting = enigmaConfig.Rotors[1]
						position1Setting = enigmaConfig.Positions[0]
						position2Setting = enigmaConfig.Positions[1]

					}
				}
			}

		}
	}

	enigmaConfig.Rotors[0] = rotor1Setting
	enigmaConfig.Rotors[1] = rotor2Setting
	enigmaConfig.Positions[0] = position1Setting
	enigmaConfig.Positions[1] = position2Setting

	bestEnigmaPlugboard := createEnigmaPlugboard(bestPlugboardSetting)
	text := setEnigmaAndDecode(cipherText, bestEnigmaPlugboard)
	fmt.Println("Plain Text:")
	fmt.Println(text)

	for index := range enigmaConfig.Rotors {
		if index != 0 {
			fmt.Print(" ")
		}
		fmt.Print(enigmaConfig.Rotors[index])
	}
	fmt.Println()
	for index := range enigmaConfig.Positions {
		if index != 0 {
			fmt.Print(" ")
		}
		fmt.Print(enigmaConfig.Positions[index])
	}
	fmt.Println()

	a := createEnigmaPlugboard(bestPlugboardSetting)
	for index := range a {
		if index != 0 {
			fmt.Print(" ")
		}
		fmt.Print(a[index])
	}
}
