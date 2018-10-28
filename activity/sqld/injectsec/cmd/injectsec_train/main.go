// Copyright 2018 The InjectSec Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"regexp"
	"sort"
	"strings"

	dat "github.com/project-flogo/microgateway/activity/sqld/injectsec/data"
	"github.com/project-flogo/microgateway/activity/sqld/injectsec/gru"
)

var (
	rnd *rand.Rand
	// FuzzFiles are the files to train on
	FuzzFiles = []string{
		"./data/Generic-BlindSQLi.fuzzdb.txt",
		"./data/Generic-SQLi.txt",
	}
)

// Example is a training example
type Example struct {
	Data   []byte
	Attack bool
}

// Examples are a set of examples
type Examples []Example

// Permute puts the examples into random order
func (e Examples) Permute() {
	length := len(e)
	for i := range e {
		j := i + rand.Intn(length-i)
		e[i], e[j] = e[j], e[i]
	}
}

func generateTrainingData() (training, validation Examples) {
	generators := dat.TrainingDataGenerator(rnd)
	for _, generator := range generators {
		if generator.SkipTrain == true {
			continue
		}
		if generator.Regex != nil {
			parts := dat.NewParts()
			generator.Regex(parts)
			for i := 0; i < 128; i++ {
				line, err := parts.Sample(rnd)
				if err != nil {
					panic(err)
				}
				training = append(training, Example{[]byte(strings.ToLower(line)), true})
			}
		}
	}

	var symbols []rune
	for s := '0'; s <= '9'; s++ {
		symbols = append(symbols, s)
	}
	for i := 0; i < 128; i++ {
		example, size := "", 1+rnd.Intn(8)
		for j := 0; j < size; j++ {
			example += string(symbols[rnd.Intn(len(symbols))])
		}
		training = append(training, Example{[]byte(strings.ToLower(example)), false})
	}

	for s := 'a'; s <= 'z'; s++ {
		symbols = append(symbols, s)
	}
	for i := 0; i < 2048; i++ {
		left, size := "", 1+rnd.Intn(8)
		if rnd.Intn(2) == 0 {
			left += " "
		}
		for j := 0; j < size; j++ {
			left += string(symbols[rnd.Intn(len(symbols))])
		}
		right, size := "", 1+rnd.Intn(8)
		for j := 0; j < size; j++ {
			right += string(symbols[rnd.Intn(len(symbols))])
		}
		if rnd.Intn(2) == 0 {
			right += " "
		}
		example := ""
		switch rnd.Intn(3) {
		case 0:
			example = left + "or" + right
		case 1:
			example = left + "or"
		case 2:
			example = "or" + right
		}
		training = append(training, Example{[]byte(strings.ToLower(example)), false})
	}

	var symbolsNumeric, symbolsAlphabet []rune
	for s := '0'; s <= '9'; s++ {
		symbolsNumeric = append(symbolsNumeric, s)
	}
	for s := 'a'; s <= 'z'; s++ {
		symbolsAlphabet = append(symbolsAlphabet, s)
	}
	length := len(training)
	for i := 0; i < length; i++ {
		words, example, ws := rnd.Intn(3)+1, "", ""
		for w := 0; w < words; w++ {
			example += ws
			size, typ := 1+rnd.Intn(16), rnd.Intn(3)
			switch typ {
			case 0:
				for j := 0; j < size; j++ {
					example += string(symbolsNumeric[rnd.Intn(len(symbolsNumeric))])
				}
			case 1:
				for j := 0; j < size; j++ {
					example += string(symbolsAlphabet[rnd.Intn(len(symbolsAlphabet))])
				}
			case 2:
				for j := 0; j < size; j++ {
					example += string(symbols[rnd.Intn(len(symbols))])
				}
			}
			ws = " "
		}
		training = append(training, Example{[]byte(strings.ToLower(example)), false})
	}

	training.Permute()
	validation = training[:2000]
	training = training[2000:]

	for _, generator := range generators {
		if generator.SkipTrain == true {
			continue
		}
		if generator.Case == "" {
			training = append(training, Example{[]byte(strings.ToLower(generator.Form)), true})
		} else {
			training = append(training, Example{[]byte(strings.ToLower(generator.Case)), true})
		}
	}

	return
}

func printChunks() {
	chunks := make(map[string]int, 0)
	for _, file := range FuzzFiles {
		in, err := os.Open(file)
		if err != nil {
			panic(err)
		}
		reader := bufio.NewReader(in)
		line, err := reader.ReadString('\n')
		for err == nil {
			line = strings.ToLower(strings.TrimSuffix(line, "\n"))
			symbols, buffer := []rune(line), make([]rune, 0, 32)
			for _, v := range symbols {
				if v >= 'a' && v <= 'z' {
					buffer = append(buffer, v)
				} else if len(buffer) > 1 {
					chunks[string(buffer)]++
					buffer = buffer[:0]
				} else {
					buffer = buffer[:0]
				}
			}
			line, err = reader.ReadString('\n')
		}
	}
	type Chunk struct {
		Chunk string
		Count int
	}
	ordered, i := make([]Chunk, len(chunks)), 0
	for k, v := range chunks {
		ordered[i] = Chunk{
			Chunk: k,
			Count: v,
		}
		i++
	}
	sort.Slice(ordered, func(i, j int) bool {
		return ordered[i].Count > ordered[j].Count
	})
	for _, v := range ordered {
		fmt.Println(v)
	}
	fmt.Println(len(chunks))
}

var (
	help   = flag.Bool("help", false, "print help")
	chunks = flag.Bool("chunks", false, "generate chunks")
	print  = flag.Bool("print", false, "print training data")
	parts  = flag.Bool("parts", false, "test parts")
	data   = flag.String("data", "", "use data for training")
	epochs = flag.Int("epochs", 1, "the number of epochs for training")
)

func main() {
	flag.Parse()
	if *help {
		flag.Usage()
		return
	}

	rnd = rand.New(rand.NewSource(1))

	if *chunks {
		printChunks()
		return
	}

	if *print {
		generators := dat.TrainingDataGenerator(rnd)
		for _, generator := range generators {
			fmt.Println(generator.Form)
			if generator.Regex != nil {
				parts := dat.NewParts()
				generator.Regex(parts)
				for i := 0; i < 10; i++ {
					fmt.Println(parts.Sample(rnd))
				}
			}
			fmt.Println()
		}
		return
	}

	if *parts {
		generators, count, attempts, nomatch := dat.TrainingDataGenerator(rnd), 0, 0, 0
		for _, generator := range generators {
			if generator.Regex != nil {
				parts := dat.NewParts()
				generator.Regex(parts)
				exp, err := parts.Regex()
				if err == nil {
					regex, err := regexp.Compile(exp)
					if err != nil {
						panic(err)
					}
					form := strings.ToLower(generator.Form)
					if generator.Case != "" {
						form = strings.ToLower(generator.Case)
					}
					attempts++
					if !regex.MatchString(form) {
						nomatch++
						fmt.Println(exp)
						fmt.Println(form)
						fmt.Println()
					}
				}
				count++
			}
		}
		fmt.Println(count, attempts, nomatch)
		return
	}

	os.Mkdir("output", 0744)
	results, err := os.Create("output/results.txt")
	if err != nil {
		panic(err)
	}
	defer results.Close()

	printResults := func(a ...interface{}) {
		s := fmt.Sprint(a...)
		fmt.Println(s)
		results.WriteString(s + "\n")
	}

	training, validation := generateTrainingData()
	if *data != "" {
		in, err1 := os.Open(*data)
		if err1 != nil {
			panic(err1)
		}
		defer in.Close()
		var custom Examples
		reader := csv.NewReader(in)
		line, err1 := reader.Read()
		for err1 != nil {
			example := Example{
				Data:   []byte(line[0]),
				Attack: line[1] == "attack",
			}
			custom = append(custom, example)
			line, err1 = reader.Read()
		}
		custom.Permute()
		cutoff := (80 * len(custom)) / 100
		training = append(training, custom[:cutoff]...)
		validation = append(validation, custom[cutoff:]...)
	}

	fmt.Println(len(training))

	networkRnd := rand.New(rand.NewSource(1))
	network := gru.NewGRU(networkRnd)

	for epoch := 0; epoch < *epochs; epoch++ {
		training.Permute()
		for i, example := range training {
			cost := network.Train(example.Data, example.Attack)
			if i%100 == 0 {
				fmt.Println(cost)
			}
		}

		file := fmt.Sprintf("output/w%v.w", epoch)
		printResults(file)
		err = network.WriteFile(file)
		if err != nil {
			panic(err)
		}

		correct, attacks, nattacks := 0, 0, 0
		for i := range validation {
			example := validation[i]
			attack := network.Test(example.Data)
			if example.Attack == attack {
				correct++
			} else {
				printResults(string(example.Data), example.Attack, attack)
			}
			if example.Attack {
				attacks++
			} else {
				nattacks++
			}
		}
		printResults(attacks, nattacks, correct, len(validation))
	}
}
