package huhtest

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	dummyResponse  = &response{answers: []string{"foo"}}
	dummyResponse2 = &response{answers: []string{"bar"}}
)

func TestResponses_Find_ReturnsExpectedResponses(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		responses *responses
		line      string

		expectedQuestion string
		expectedOk       bool
	}{
		"exact response is found correctly": {
			responses: &responses{
				exactQuestions: map[string]*response{"Hello World?": dummyResponse},
			},
			line: "Hello World?",

			expectedQuestion: "Hello World?",
			expectedOk:       true,
		},
		"substring response is found correctly": {
			responses: &responses{
				substringQuestions: map[string]*response{"Hello": dummyResponse},
			},
			line: "Hello World?",

			expectedQuestion: "Hello",
			expectedOk:       true,
		},
		"regexp response is found correctly": {
			responses: &responses{
				regexQuestions: map[string]*response{`Hel{2}o [Ww]orl.\?`: dummyResponse},
				regexCache:     map[string]*regexp.Regexp{`Hel{2}o [Ww]orl.\?`: regexp.MustCompile(`Hel{2}o [Ww]orl.\?`)},
			},
			line: "Hello World?",

			expectedQuestion: `Hel{2}o [Ww]orl.\?`,
			expectedOk:       true,
		},
		"mismatch on exact returns false": {
			responses: &responses{
				exactQuestions: map[string]*response{"Hello world": dummyResponse},
			},
			line: "Hello World?",

			expectedOk: false,
		},
	}

	for name, testData := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			// Act
			result, question, ok := testData.responses.find(testData.line)

			// Assert
			require.Equal(t, testData.expectedOk, ok)

			if ok {
				require.Equal(t, testData.expectedQuestion, question)
				require.Equal(t, dummyResponse, result)
				return
			}

			require.Nil(t, result)
		})
	}
}

func TestResponses_Add_AddsExactQuestionToExactMap(t *testing.T) {
	t.Parallel()
	// Arrange
	response := newResponses()
	question := "How many fingers?"

	// Act
	response.add(question, questionMatchExact, *dummyResponse)

	// Assert
	require.NotEmpty(t, response.exactQuestions[question])
}

func TestResponses_Add_CombinesConsecutiveCallsOnExact(t *testing.T) {
	t.Parallel()
	// Arrange
	response := newResponses()
	question := "How many fingers?"

	response.add(question, questionMatchExact, *dummyResponse)

	// Act
	response.add(question, questionMatchExact, *dummyResponse2)

	// Assert
	require.NotEmpty(t, response.exactQuestions[question])

	answers := response.exactQuestions[question].answers
	require.Equal(t, []string{"foo", "bar"}, answers)
}

func TestResponses_Add_AddsSubstringQuestionToContainsMap(t *testing.T) {
	t.Parallel()
	// Arrange
	response := newResponses()
	question := "How many fingers?"

	// Act
	response.add(question, questionMatchSubstring, *dummyResponse)

	// Assert
	require.NotEmpty(t, response.substringQuestions[question])
}

func TestResponses_Add_CombinesConsecutiveCallsOnSubstring(t *testing.T) {
	t.Parallel()
	// Arrange
	response := newResponses()
	question := "How many fingers?"

	response.add(question, questionMatchSubstring, *dummyResponse)

	// Act
	response.add(question, questionMatchSubstring, *dummyResponse2)

	// Assert
	require.NotEmpty(t, response.substringQuestions[question])

	answers := response.substringQuestions[question].answers
	require.Equal(t, []string{"foo", "bar"}, answers)
}

func TestResponses_Add_AddsRegexQuestionToRegexMap(t *testing.T) {
	t.Parallel()
	// Arrange
	response := newResponses()
	question := "How many fingers?"

	// Act
	response.add(question, questionMatchRegexp, *dummyResponse)

	// Assert
	require.NotEmpty(t, response.regexQuestions[question])
}

func TestResponses_Add_CombinesConsecutiveCallsOnRegexp(t *testing.T) {
	t.Parallel()
	// Arrange
	response := newResponses()
	question := "How many fingers?"

	response.add(question, questionMatchRegexp, *dummyResponse)

	// Act
	response.add(question, questionMatchRegexp, *dummyResponse2)

	// Assert
	require.NotEmpty(t, response.regexQuestions[question])

	answers := response.regexQuestions[question].answers
	require.Equal(t, []string{"foo", "bar"}, answers)
}

func TestResponses_Add_PanicsOnUnrecognisedQuestionType(t *testing.T) {
	t.Parallel()
	// Arrange
	responses := newResponses()

	// Act
	result := func() { responses.add("", "", *dummyResponse) }

	// Assert
	require.PanicsWithValue(t, "unknown question match type", result)
}

func TestResponse_PickAnswer_ReturnsErrorOnRanOutOfAttempts(t *testing.T) {
	t.Parallel()

	// Arrange
	res := response{
		answers:       []string{"a", "b", "c"},
		expectedTimes: 3,
	}

	result := make([]string, 6)
	errs := make([]error, 6)

	// Act
	for index := range 6 {
		result[index], errs[index] = res.pickAnswer()
	}

	// Assert
	expectedErrors := []error{
		nil,
		nil,
		nil,
		fmt.Errorf("called 4/3 times: %w", errRanOutOfResponses),
		fmt.Errorf("called 5/3 times: %w", errRanOutOfResponses),
		fmt.Errorf("called 6/3 times: %w", errRanOutOfResponses),
	}

	expectedAnswers := []string{"a", "b", "c", "c", "c", "c"}

	assert.Equal(t, expectedErrors, errs)
	assert.Equal(t, expectedAnswers, result)
}

func TestResponse_PickAnswer_ReturnsExpectedAnswers(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		answers []string
		calls   int

		expected []string
	}{
		"one answer is always the same": {
			answers:  []string{"a"},
			expected: []string{"a", "a", "a"},
			calls:    3,
		},
		"multiple answers get walked through before the final one is repeated": {
			answers:  []string{"a", "b", "c", "d"},
			expected: []string{"a", "b", "c", "d", "d", "d", "d"},
			calls:    7,
		},
	}

	for name, testData := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			// Arrange
			res := response{
				answers: testData.answers,
			}

			result := make([]string, testData.calls)
			errs := make([]error, testData.calls)

			// Act
			for index := range testData.calls {
				result[index], errs[index] = res.pickAnswer()
			}

			// Assert
			for _, err := range errs {
				require.NoError(t, err)
			}

			assert.Equal(t, testData.expected, result)
			assert.Equal(t, testData.calls, res.actualTimes)
		})
	}
}

func TestResponse_LastAnswer_ReturnsLastAnswer(t *testing.T) {
	t.Parallel()
	// Arrange
	res := response{
		answers: []string{"a", "b", "c", "d", "e"},
	}

	// Act
	result := res.lastAnswer()

	// Assert
	assert.Equal(t, "e", result)
}

func TestResponse_SubmitCharacter_ReturnsExpectedCharacter(t *testing.T) {
	t.Parallel()

	tests := map[string]string{
		"":              defaultSubmit,
		"anything else": "anything else",
	}

	for submitCharacterOverride, expected := range tests {
		t.Run(submitCharacterOverride, func(t *testing.T) {
			t.Parallel()
			res := response{
				submitCharacterOverride: submitCharacterOverride,
			}

			// Act
			result := res.submitCharacter()

			// Assert
			assert.Equal(t, expected, result)
		})
	}
}
