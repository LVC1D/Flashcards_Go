package main

import (
	"bufio"
	. "fmt"
	"log"
	"math/rand"
	"os"
	"slices"
	"strconv"
	"strings"
)

type Cached struct {
	data       map[string]string
	failedData map[string]int
}

var (
	c                                        = Cached{data: make(map[string]string), failedData: make(map[string]int)}
	logData                                  []string
	fileName, importFileName, exportFileName string
)

func main() {
	for _, arg := range os.Args[1:] {
		argSlc := strings.Split(arg, "=")
		flag, value := argSlc[0], argSlc[1]
		switch flag {
		case "--import_from":
			importFileName = value
			c.importCards(importFileName)
		case "--export_to":
			exportFileName = value
		default:
			Println("Unknown flag:", flag)
		}
	}

	runGameLogic(exportFileName)
}

func runGameLogic(exportFileName string) {
	for {
		var action string
		prompt := "Input the action (add, remove, import, export, ask, exit, log, hardest card, reset stats):"
		Println(prompt)
		logData = append(logData, prompt)

		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		action = scanner.Text()
		logData = append(logData, action)

		switch action {
		case "add":
			c.addCards()
		case "remove":
			c.removeCard()
		case "ask":
			c.playGame()
		case "export":
			Println("File name:")
			Scanln(&exportFileName)
			logData = append(logData, "File name:", fileName)
			c.exportCards(exportFileName)
		case "import":
			Println("File name:")
			Scanln(&fileName)
			logData = append(logData, "File name:", fileName)
			c.importCards(fileName)
		case "log":
			logStuff(logData)
		case "hardest card":
			showHardest(&c)
		case "reset stats":
			resetStats(&c)

		case "exit":
			bye := "Bye bye!"
			logData = append(logData, bye)
			Println(bye)

			if exportFileName != "" {
				c.exportCards(exportFileName)
			}
			os.Exit(0)
		}
	}
}

func (deck *Cached) addCards() {
	reader := bufio.NewReader(os.Stdin)
	Println("The card:")
	term, _ := reader.ReadString('\n')
	term = Sprintf("\"%s\"", strings.TrimSpace(term))
	logData = append(logData, "The card:", term)

	for oldTerm, _ := range deck.data {
		for strings.EqualFold(term, oldTerm) {
			alreadyExists := Sprintf("The term %s already exists. Try again:\n", oldTerm)
			logData = append(logData, alreadyExists)
			Println(alreadyExists)

			term, _ = reader.ReadString('\n')
			term = Sprintf("\"%s\"", strings.TrimSpace(term))
		}
	}

	Println("The definition of the card:")
	logData = append(logData, "The definition of the card:")
	definition, _ := reader.ReadString('\n')
	definition = strings.TrimSpace(definition)

	for _, oldDef := range deck.data {
		for oldDef == definition {
			alsoExists := Sprintf("The definition \"%s\" already exists. Try again:\n", oldDef)
			logData = append(logData, alsoExists)
			Println(alsoExists)
			definition, _ = reader.ReadString('\n')
			definition = strings.TrimSpace(definition)
		}
	}

	deck.data[term] = definition
	cardAddedMessage := Sprintf("The pair (%s:\"%s\") has been added.\n", term, definition)
	logData = append(logData, cardAddedMessage)
	Println(cardAddedMessage)
}

func (deck *Cached) removeCard() {
	reader := bufio.NewReader(os.Stdin)
	Println("Which card?")
	term, _ := reader.ReadString('\n')
	term = Sprintf("\"%s\"", strings.TrimSpace(term))
	logData = append(logData, "Which card?", term)

	if _, exists := deck.data[term]; exists {
		delete(deck.data, term)
		removed := "The card has been removed."
		logData = append(logData, removed)
		Println(removed)
	} else {
		cantRemove := Sprintf("Can't remove \"%s\": there is no such card.\n", term)
		logData = append(logData, cantRemove)
		Println(cantRemove)
	}
}

func (deck *Cached) playGame() {
	var count int
	var anotherFound bool
	reader := bufio.NewReader(os.Stdin)

	Println("How many times to ask?")
	Scanln(&count)
	logData = append(logData, "How many times to ask?", strconv.Itoa(count))

	keys := make([]string, 0, count)
	for key, _ := range deck.data {
		keys = append(keys, key)
	}

	for i, n := 0, count; i < n; i++ {
		if len(deck.data) < 1 {
			noCardsMess := "There are no cards available in memory."
			logData = append(logData, noCardsMess)
			Println(noCardsMess)
			return
		}
		randomIndex := rand.Intn(len(deck.data))
		randomKey := keys[randomIndex]

		guessPrompt := Sprintf("Print the definition of %s:", randomKey)
		Println(guessPrompt)
		answer, _ := reader.ReadString('\n')
		answer = strings.TrimSpace(answer)
		logData = append(logData, guessPrompt, answer)

		if answer == deck.data[randomKey] {
			logData = append(logData, "Correct!")
			Println("Correct!")
		} else {
			for termAgain, defAgain := range deck.data {
				if answer == defAgain {
					wrongButNot := Sprintf("Wrong. The right answer is \"%s\", but your definition is correct for %s.\n", deck.data[randomKey], termAgain)
					logData = append(logData, wrongButNot)
					Println(wrongButNot)
					anotherFound = true
				}
			}
			if !anotherFound {
				justWrong := Sprintf("Wrong. The answer is \"%s\".\n", deck.data[randomKey])
				logData = append(logData, justWrong)
				Println(justWrong)
			}
			deck.failedData[randomKey]++
		}
		anotherFound = false
	}
}

func (deck *Cached) exportCards(fn string) {
	var (
		cardRecords []string
		count       int
	)

	file, err := os.OpenFile(fn, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	for key, val := range deck.data {
		record := Sprintf("%s: %s", strings.TrimSpace(key), strings.TrimSpace(val))
		cardRecords = append(cardRecords, record)
	}

	for _, cardRecord := range cardRecords {
		if _, errz := Fprintln(file, cardRecord); errz != nil {
			log.Fatal(errz)
		}
		count++
	}

	exportSavedMessage := Sprintf("%d cards have been saved.\n", count)
	logData = append(logData, exportSavedMessage)
	Println(exportSavedMessage)
}

func (deck *Cached) importCards(fn string) {
	var count int

	file, err := os.Open(fn)
	if err != nil {
		Println("File not found.")
		logData = append(logData, "File not found.")
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		lineSlice := strings.Split(line, ": ")
		if _, ok := deck.data[lineSlice[0]]; ok {
			deck.data[lineSlice[0]] = lineSlice[1]
		} else {
			deck.data[strings.TrimSpace(lineSlice[0])] = strings.TrimSpace(lineSlice[1])
		}
		count++
	}
	if errz := scanner.Err(); errz != nil {
		log.Fatal(errz)
	}
	cardsLoadedMessage := Sprintf("%d cards have been loaded.\n", count)
	logData = append(logData, cardsLoadedMessage)
	Println(cardsLoadedMessage)
}

func logStuff(data []string) {
	var fileName string
	Println("File name:")
	Scanln(&fileName)
	logData = append(logData, "File name:", fileName)

	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	logSaved := "The log has been saved."
	Println(logSaved)
	logData = append(logData, logSaved)

	for _, line := range data {
		if _, errz := Fprintln(file, line); errz != nil {
			log.Fatal(errz)
		}
	}

	logData = []string{}
}

func showHardest(deck *Cached) {
	if len(deck.failedData) == 0 {
		noCardsErr := "There are no cards with errors."
		logData = append(logData, noCardsErr)
		Println(noCardsErr)
		return
	}

	var (
		countSlice         []int
		sameCountSlice     []string
		highest, sameCount int
	)

	for _, val := range deck.failedData {
		countSlice = append(countSlice, val)
	}

	highest = slices.Max(countSlice)

	for _, val := range countSlice {
		if val == highest {
			sameCount++
		}
	}

	for term, val := range deck.failedData {
		if val == highest {
			if sameCount == 1 {
				singleFailCard := Sprintf("The hardest card is %s. You have %d errors answering it.\n", term, highest)
				logData = append(logData, singleFailCard)
				Println(singleFailCard)
				return
			}
			sameCountSlice = append(sameCountSlice, term)
		}
	}

	multiString := Sprintf("The hardest cards are %s. You have %d errors answering them.", strings.Join(sameCountSlice, ", "), highest)
	logData = append(logData, multiString)
	Println(multiString)
}

func resetStats(deck *Cached) {
	for key, _ := range deck.failedData {
		delete(deck.failedData, key)
	}
	resetMessage := "Card statistics have been reset."
	logData = append(logData, resetMessage)
	Println(resetMessage)
}
