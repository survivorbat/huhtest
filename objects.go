package huhtest

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// questionMatchType dictates how we should compare whether an incoming line
// with a given question. By default we match against questions exactly,
// but we also offer the option to match by substring and regexp.
type questionMatchType string

const (
	// questionMatchExact will expect a given output to be an exact copy of the question. This usually
	// doesn't work because huh questions contain formatting/
	questionMatchExact questionMatchType = "exact"

	// questionMatchSubstring uses strings.Contains to determine whether a line
	// matches a question
	questionMatchSubstring questionMatchType = "substring"

	// questionMatchRegexp compiles the given response into regexp to determine
	// whether it matches
	questionMatchRegexp questionMatchType = "regexp"
)

// defaultQuestionMatchType is set to substring, as we expect huh output to contain formatting and thus
// not be suitable for exact matching
const defaultQuestionMatchType = questionMatchSubstring

// newResponses makes it easier to instantiate all the maps inside of the struct
func newResponses() *responses {
	return &responses{
		exactQuestions:     make(map[string]*response),
		substringQuestions: make(map[string]*response),
		regexQuestions:     make(map[string]*response),
		regexCache:         make(map[string]*regexp.Regexp),
	}
}

// responses is a collection of responses that the user has registered in a Responder,
// with convenient methods that make it easier to retrieve a desired response.
//
// It keeps 3 separate maps with questions that are matched using different questionMatchType options.
type responses struct {
	// exactQuestions is for questionMatchExact questions and should be evaluated first
	exactQuestions map[string]*response

	// substringQuestions is for questionMatchSubstring questions and should be evaluated second,
	// as an exact match should have priority over partial matches. This is expected to use strings.Contains
	// for matching.
	substringQuestions map[string]*response

	// regexQuestions saves the responses and the string version of the regex, because regexp.Regexp objects
	// can't be map keys.
	regexQuestions map[string]*response

	// regexCache keeps track of the actual Regexp objects so that we don't have to
	// compile them on every find call..
	regexCache map[string]*regexp.Regexp
}

// find traverses the 3 question types for a match with the given line. The order is from easy to
// difficult, starting with exact matches and ending with regexp. It returns a response if found,
// the question it matched with and a boolean that indicates whether a question was found.
func (q *responses) find(line string) (*response, string, bool) {
	if response, ok := q.exactQuestions[line]; ok {
		return response, line, true
	}

	for question, response := range q.substringQuestions {
		if strings.Contains(line, question) {
			return response, question, true
		}
	}

	for question, response := range q.regexQuestions {
		if q.regexCache[question].MatchString(line) {
			return response, question, true
		}
	}

	return nil, "", false
}

// add sorts a new question into the relevant maps, and will compile a regexp into the cache
// list if one is given. Since this code is unexported, we've opted to let it panic on an
// unknown questionMatchType instead or eturning an error, as it should be near impossible to
// trigger that path.
func (q *responses) add(question string, matchType questionMatchType, res response) {
	switch matchType {
	case questionMatchExact:
		if existing, ok := q.exactQuestions[question]; ok {
			res.answers = append(existing.answers, res.answers...)
		}

		q.exactQuestions[question] = &res

	case questionMatchSubstring:
		if existing, ok := q.substringQuestions[question]; ok {
			res.answers = append(existing.answers, res.answers...)
		}

		q.substringQuestions[question] = &res

	case questionMatchRegexp:
		if existing, ok := q.regexQuestions[question]; ok {
			res.answers = append(existing.answers, res.answers...)
		}

		q.regexQuestions[question] = &res
		q.regexCache[question] = regexp.MustCompile(question)

	default:
		panic("unknown question match type")
	}
}

// response contains a list of answers that should be returned in order. It also keeps
// track of how many times it;s been called and how many times we expect it to be called.
type response struct {
	// answers will be walked through on consecutive pickAnswer calls, with the final answer
	// being repeated
	answers []string

	// submitCharacter is used if non-empty, as some questions may get tangled if we use the defaultSubmit
	submitCharacterOverride string

	actualTimes   int
	expectedTimes int
}

// errRanOutOfResponses may be returned by pickAnswer if an expectedTimes is set.
var errRanOutOfResponses = errors.New("ran out of responses")

// pickAnswer contains the logic to pick the next answer to return to a question. If
// the expectedTimes is set in the response object, it will also start returning errRanOutOfResponses as soon
// as that number is reached. If we ran out of answers, we'll keep repeating the last answer in the list, even on error.
func (q *response) pickAnswer() (string, error) {
	defer func() { q.actualTimes++ }()

	if q.expectedTimes != 0 && q.actualTimes >= q.expectedTimes {
		return q.lastAnswer(), fmt.Errorf("called %d/%d times: %w", q.actualTimes+1, q.expectedTimes, errRanOutOfResponses)
	}

	if len(q.answers)-1 >= q.actualTimes {
		return q.answers[q.actualTimes], nil
	}

	return q.lastAnswer(), nil
}

// lastAnswer is a convenience method for getting the final answer in the answers slice.
func (q *response) lastAnswer() string {
	return q.answers[len(q.answers)-1]
}

// submitCharacter is used to catch any special submit situations, such as with select questions
// that only require a \r and not the \n. If no override character has been defined, defaultSubmit is returned.
func (q *response) submitCharacter() string {
	if q.submitCharacterOverride != "" {
		return q.submitCharacterOverride
	}

	return defaultSubmit
}
