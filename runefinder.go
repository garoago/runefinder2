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
	for k := range rs {
		if other.Contains(k) {
			result.Put(k)
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

func buildIndex(fileName string) (map[string]RuneSet, map[rune]string) {
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		getUcdFile(fileName)
	}
	content, err := ioutil.ReadFile(fileName)
	if err != nil {
		panic(err)
	}
	lines := strings.Split(string(content), "\n")

	index := map[string]RuneSet{}
	names := map[rune]string{}

	for _, line := range lines {
		var uchar rune
		fields := strings.Split(line, ";")
		if len(fields) >= 2 {
			code64, _ := strconv.ParseInt(fields[0], 16, 0)
			uchar = rune(code64)
			names[uchar] = fields[1]
			for _, word := range strings.Split(fields[1], " ") {
				existing, ok := index[word]
				if !ok {
					existing = RuneSet{}
				}
				existing.Put(uchar)
				index[word] = existing
			}
		}

	}
	return index, names
}

func getIndex() (map[string]RuneSet, map[rune]string) {
	dir, _ := os.Getwd()
	path := path.Join(dir, ucdFileName)
	return buildIndex(path)
}

func findRunes(query []string, index map[string]RuneSet) RuneSlice {
	commonRunes := RuneSet{}
	for i, word := range query {
		word = strings.ToUpper(word)
		found := index[word]
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

	index, names := getIndex()

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
