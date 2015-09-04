package main

import (
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	ucdUrl        = "http://www.unicode.org/Public/UCD/latest/ucd/UnicodeData.txt"
	indexFileName = "runefinder-index.gob"
)

type RuneSlice []rune

func (rs RuneSlice) Len() int           { return len(rs) }
func (rs RuneSlice) Less(i, j int) bool { return rs[i] < rs[j] }
func (rs RuneSlice) Swap(i, j int)      { rs[i], rs[j] = rs[j], rs[i] }

type RuneSet map[rune]struct{}

func (rs RuneSet) Put(key rune) {
	rs[key] = struct{}{} // zero-byte struct
}

func (rs RuneSet) Contains(key rune) bool {
	_, found := rs[key]
	return found
}

func (rs RuneSet) Intersection(other RuneSet) RuneSet {
	result := RuneSet{}
	if len(other) > 0 {
		for k := range rs {
			if other.Contains(k) {
				result.Put(k)
			}
		}
	}
	return result
}

func (rs RuneSet) ToRuneSlice() RuneSlice {
	result := RuneSlice{}
	for uchar := range rs {
		result = append(result, uchar)
	}
	return result
}

func (rs RuneSet) String() string {
	sl := rs.ToRuneSlice()
	sort.Sort(sl)
	str := "❮"
	for i, uchar := range sl {
		if i > 0 {
			str += " "
		}
		str += fmt.Sprintf("U+%04X", uchar)
	}
	return str + "❯"
}

type RuneIndex struct {
	Characters map[string]RuneSet
	Names      map[rune]string
}

func progressDisplay(running <-chan bool) {
	for {
		select {
		case <-running:
			fmt.Println()
		case <-time.After(200 * time.Millisecond):
			fmt.Print(".")
		}
	}
}

func getUcdLines() []string {
	fmt.Printf("Index not found. Retrieving data from:\n%s\n", ucdUrl)
	running := make(chan bool)
	go progressDisplay(running)
	defer func() {
		running <- false
	}()
	response, err := http.Get(ucdUrl)
	if err != nil {
		log.Fatal("getUcdFile/http.Get:", err)
	}
	content, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal("buildIndex/ioutil.ReadAll:", err)
	}
	defer response.Body.Close()
	return strings.Split(string(content), "\n")
}

func buildIndex() (index RuneIndex) {
	index.Characters = map[string]RuneSet{}
	index.Names = map[rune]string{}

	for _, line := range getUcdLines() {
		var uchar rune
		fields := strings.Split(line, ";")
		if len(fields) >= 2 {
			code64, _ := strconv.ParseInt(fields[0], 16, 0)
			uchar = rune(code64)
			index.Names[uchar] = fields[1]
			words := strings.Split(strings.ToUpper(fields[1]), " ")
			for _, word := range words {
				existing, found := index.Characters[word]
				if !found {
					existing = RuneSet{}
				}
				existing.Put(uchar)
				index.Characters[word] = existing
			}
		}

	}
	return index
}

func saveIndex(index RuneIndex, indexPath string, saved chan<- bool) {
	indexFile, err := os.Create(indexPath)
	if err != nil {
		log.Printf("WARNING: Unable to save index file.")
		saved <- false
	} else {
		defer indexFile.Close()
		encoder := gob.NewEncoder(indexFile)
		encoder.Encode(index)
		saved <- true
	}
}

func getIndex(saved chan<- bool) (index RuneIndex) {
	indexDir, _ := os.Getwd()
	indexPath := path.Join(indexDir, indexFileName)
	indexFile, err := os.Open(indexPath)
	switch {
	case err == nil:
		// load existing index
		defer indexFile.Close()
		decoder := gob.NewDecoder(indexFile)
		err = decoder.Decode(&index)
		if err != nil {
			log.Fatal("getIndex/Decode:", err)
		}
		go func() { saved <- false }()
	case os.IsNotExist(err):
		// build and save index
		index = buildIndex()
		go saveIndex(index, indexPath, saved)
	default:
		log.Fatal("getIndex/os.Open:", err)
	}
	return index
}

func findRunes(query []string, index RuneIndex) RuneSlice {
	commonRunes := RuneSet{}
	for i, word := range query {
		word = strings.ToUpper(word)
		found := index.Characters[word]
		if i == 0 {
			commonRunes = found
		} else {
			commonRunes = commonRunes.Intersection(found)
		}
		if len(commonRunes) == 0 {
			break
		}
	}
	result := commonRunes.ToRuneSlice()
	sort.Sort(result)
	return result
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage:  runefinder <word>\texample: runefinder cat")
		os.Exit(1)
	}
	words := os.Args[1:]

	saved := make(chan bool)
	index := getIndex(saved)

	count := 0
	format := "U+%04X  %c \t%s\n"
	for _, uchar := range findRunes(words, index) {
		if uchar > 0xFFFF {
			format = "U+%5X %c \t%s\n"
		}
		fmt.Printf(format, uchar, uchar, index.Names[uchar])
		count++
	}
	fmt.Printf("%d characters found\n", count)
	if <-saved {
		fmt.Println("Index saved.")
	}
}
