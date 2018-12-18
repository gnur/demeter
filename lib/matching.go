package lib

import (
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

var onlyLower = regexp.MustCompile("[^a-z]+")
var leadingNumbers = regexp.MustCompile("^ *[0-9]+")
var betweenParentheses = regexp.MustCompile(`\(.*\)`)
var betweenBockHooks = regexp.MustCompile(`\[.*\]`)

var leadingZeroes = regexp.MustCompile(`^ *(0)([0-9]+) `)
var alphaNumeric = regexp.MustCompile(`[^a-z0-9]+`)

//var year = regexp.MustCompile(`(19[0-9]{2})|(20[0-9]{2})`)

var uselessWords = []string{
	"le", "la", "et",
	"de", "het", "en",
	"the", "and", "a", "an",
}

func hashBook(author, title string) string {
	author = strings.ToLower(author)
	author = strings.Replace(author, "-", " ", -1)
	title = strings.ToLower(title)

	authorParts := strings.Split(author, " ")
	lastName := authorParts[len(authorParts)-1]

	//remove author from title
	title = strings.Replace(title, author, "", -1)
	title = strings.Replace(title, lastName, "", -1)

	//remove leading numbers
	title = leadingNumbers.ReplaceAllString(title, "")

	//concatenate to half further actions
	title = lastName + " " + title

	title = removeAccents(title)

	//make sure no whitespace is on either end
	title = strings.TrimSpace(title)

	//remove everything between parenthesis
	title = betweenParentheses.ReplaceAllString(title, " ")

	//remove everything between blockhooks
	title = betweenBockHooks.ReplaceAllString(title, " ")

	//remove ': a novel'
	title = strings.Replace(title, ": a novel", " ", -1)

	//remove leading zeroes from numbers
	title = leadingZeroes.ReplaceAllString(title, " $2 ")

	//remove all non [a-z0-9]
	title = alphaNumeric.ReplaceAllString(title, "")

	return title
}

func removeAccents(in string) string {
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	s, _, err := transform.String(t, in)
	if err != nil {
		return in
	}
	return s
}
