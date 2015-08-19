package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"
)

const ucdFileName = "UnicodeData.txt"
const ucdBaseUrl = "http://www.unicode.org/Public/UCD/latest/ucd/"

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

func getUcdFile(fileName string) {

	url := ucdBaseUrl + ucdFileName
	fmt.Printf("%s not found\nretrieving from %s\n", ucdFileName, url)
	running := make(chan bool)
	go progressDisplay(running)
	defer func() {
		running <- false
	}()
	response, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer response.Body.Close()
	file, err := os.Create(fileName)
	if err != nil {
		panic(err)
	}
	_, err = io.Copy(file, response.Body)
	if err != nil {
		panic(err)
	}
	file.Close()
}

type runesSlice []rune

func (p runesSlice) Len() int           { return len(p) }
func (p runesSlice) Less(i, j int) bool { return p[i] < p[j] }
func (p runesSlice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

type runesMap map[rune]struct{}

func buildIndex(fileName string) (map[string]runesSlice, map[rune]string) {
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		getUcdFile(fileName)
	}
	content, err := ioutil.ReadFile(fileName)
	if err != nil {
		panic(err)
	}

	lines := strings.Split(string(content), "\n")

	index := map[string]runesSlice{}
	names := map[rune]string{}

	for _, line := range lines {
		var uchar rune
		fields := strings.Split(line, ";")
		if len(fields) >= 2 {
			code64, _ := strconv.ParseInt(fields[0], 16, 0)
			uchar = rune(code64)
			names[uchar] = fields[1]
			// fmt.Printf("%#v", index)
			for _, word := range strings.Split(fields[1], " ") {
				var entries runesSlice
				if len(index[word]) < 1 {
					entries = runesSlice{}
				} else {
					entries = index[word]
				}
				index[word] = append(entries, uchar)
			}
		}

	}
	return index, names
}

func intersect(a, b runesMap) runesMap {
	result := runesMap{}
	for k := range a {
		if _, occurs := b[k]; occurs {
			result[k] = struct{}{}
		}
	}
	return result
}

func findRunes(query []string, index map[string]runesSlice) runesSlice {
	foundList := []runesMap{}
	for _, word := range query {
		word = strings.ToUpper(word)
		found := runesMap{}
		for _, uchar := range index[word] {
			found[uchar] = struct{}{}
		}
		foundList = append(foundList, found)
	}
	commonRunes := foundList[0]
	for _, found := range foundList[1:] {
		commonRunes = intersect(commonRunes, found)
	}
	result := runesSlice{}
	for uchar := range commonRunes {
		result = append(result, uchar)
	}
	sort.Sort(result)
	return result
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage:  runefinder <word>\texample: runefinder cat")
		os.Exit(1)
	}
	words := os.Args[1:]

	dir, _ := os.Getwd()
	path := path.Join(dir, ucdFileName)
	index, names := buildIndex(path)

	count := 0
	format := "U+%04X  %c \t%s\n"
	for _, uchar := range findRunes(words, index) {
		if uchar > 0xFFFF {
			format = "U+%5X %c \t%s\n"
		}
		fmt.Printf(format, uchar, uchar, names[uchar])
		count++
	}
	fmt.Printf("%d characters found\n", count)

}
