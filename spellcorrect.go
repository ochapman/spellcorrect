/*
* spellcorrect.go
* Derived from http://spell-correct-in-go.googlecode.com/svn/trunk/ spell-correct-in-go-read-only
* Chapman Ou <ochapman.cn@gmail.com>
* 
*/

package spellcorrect

import (
	"io/ioutil"
	"regexp"
	"strings"
)

type SpellCorrect struct {
	Data string
	Model map[string]int
}

func New(data string) *SpellCorrect {
	if data == "" {
		panic("data is empty")
	}
	return &SpellCorrect{Data: data, Model: train(data)}
}

func train(training_data string) map[string]int {
	NWORDS := make(map[string]int)
	pattern := regexp.MustCompile("[a-z]+")
	if content, err := ioutil.ReadFile(training_data); err == nil {
		for _, w := range pattern.FindAllString(strings.ToLower(string(content)), -1) {
			NWORDS[w]++;
		}
	} else {
		panic("Failed loading training data.")
	}
	return NWORDS
}


func edits1(word string, ch chan string) {
	const alphabet = "abcdefghijklmnopqrstuvwxyz"
	type Pair struct{a, b string}
	var splits []Pair
	for i := 0; i < len(word) + 1; i++ {
		splits = append(splits, Pair{word[:i], word[i:]}) }

	for _, s := range splits {
		if len(s.b) > 0 { ch <- s.a + s.b[1:] }
		if len(s.b) > 1 { ch <- s.a + string(s.b[1]) + string(s.b[0]) + s.b[2:] }
		for _, c := range alphabet { if len(s.b) > 0 { ch <- s.a + string(c) + s.b[1:] }}
		for _, c := range alphabet { ch <- s.a + string(c) + s.b }
	}
}

func edits2(word string, ch chan string) {
	ch1 := make(chan string, 1024*1024)
	go func() { edits1(word, ch1); ch1 <- "" }()
	for e1 := range ch1 {
		if e1 == "" { break }
		edits1(e1, ch)
	}
}

func best(word string, edits func(string, chan string), model map[string]int) string {
	ch := make(chan string, 1024*1024)
	go func() { edits(word, ch); ch <- "" }()
	maxFreq := 0
	correction := ""
	for word := range ch {
		if word == "" { break }
		if freq, present := model[word]; present && freq > maxFreq {
			maxFreq, correction = freq, word
		}
	}
	return correction
}

func correct(word string, model map[string]int) string {
	if _, present := model[word]; present { return word }
	if correction := best(word, edits1, model); correction != "" { return correction }
	if correction := best(word, edits2, model); correction != "" { return correction }
	return word
}

func (sc SpellCorrect) Correct(word string) string {
	return correct(word, sc.Model)
}

func (sc *SpellCorrect) Train(pattern string) {
	NWORDS := make(map[string]int)
	if pattern == "" {
		panic("No pattern!")
	}
	re := regexp.MustCompile(pattern)
	if content, err := ioutil.ReadFile(sc.Data); err == nil {
		for _, w := range re.FindAllString(strings.ToLower(string(content)), -1) {
			NWORDS[w]++;
		}
	} else {
		panic("Failed loading training data.")
	}
	sc.Model = NWORDS
}

/*
func main() {
	model := train(os.Args[1])
	//startTime := time.Nanoseconds()
	for i := 0; i < 1; i++ { fmt.Println(Correct(os.Args[2], model)) }
	//fmt.Printf("Time : %v\n", float64(time.Nanoseconds() - startTime) / float64(1e9))
}
*/
