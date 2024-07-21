package huhtest

import (
	"bufio"
	"io"
	"strings"
	"testing"
	"time"

	testingi "github.com/mitchellh/go-testing-interface"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Functions

// simulateCLI is a test helper that serves as a stand-in for an actual huh form. Questions are sent
// to the io.PipeWriter in order and answers are captured line-by-line from the io.PipeReader.
//
// Returns the captured answers
func simulateCLI(t *testing.T, questions []string, stdout *io.PipeWriter, stdin *io.PipeReader) []string {
	t.Helper()

	actualAnswers := make([]string, len(questions))

	reader := bufio.NewReader(stdin)

	for index, question := range questions {
		t.Logf("Posing question: %s", question)

		_, err := stdout.Write([]byte(question + "\n\r"))
		require.NoError(t, err)

		line, err := reader.ReadString('\r')
		require.NoError(t, err)

		// Strip trailing submit character
		line = strings.TrimSuffix(line, "\r")

		actualAnswers[index] = readableReplacer.Replace(line)
	}

	return actualAnswers
}

// Tests

const defaultTimeout = 1 * time.Second

// Tests both the public API and the unexported methods that might one day be made public
func TestResponder_Start_ReturnsExpectedAnswers(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		responder       *Responder
		questions       []string
		expectedAnswers []string
	}{
		"one simple question": {
			responder: NewResponder().
				AddResponse("alright?", "Yes, for sure!"),
			questions:       []string{"You doing alright?"},
			expectedAnswers: []string{"Yes, for sure!"},
		},
		"three simple questions": {
			responder: NewResponder().
				AddResponse("alright?", "Yes, for sure!").
				AddResponse("dinner?", "Let's find out!").
				AddResponse("fair?", "Don't think so"),
			questions:       []string{"You doing alright?", "What's for dinner?", "Are you going to the fair?"},
			expectedAnswers: []string{"Yes, for sure!", "Let's find out!", "Don't think so"},
		},
		"one question that gets asked 3 times with different answers": {
			responder: NewResponder().
				addResponses("alright?", "Yes, for sure!", "Positive!", "Yes sir!"),
			questions:       []string{"You doing alright?", "You doing alright?", "You doing alright?"},
			expectedAnswers: []string{"Yes, for sure!", "Positive!", "Yes sir!"},
		},
		"one question that gets asked 3 times gets repeated answers": {
			responder: NewResponder().
				addResponses("alright?", "Yes, for sure!"),
			questions:       []string{"You doing alright?", "You doing alright?", "You doing alright?"},
			expectedAnswers: []string{"Yes, for sure!", "Yes, for sure!", "Yes, for sure!"},
		},

		"one affirmative confirm question": {
			responder: NewResponder().
				addConfirms("alright?", ConfirmAffirm),
			questions:       []string{"You doing alright?"},
			expectedAnswers: []string{"<right> "},
		},
		"one negative confirm question": {
			responder: NewResponder().
				addConfirms("alright?", ConfirmNegative),
			questions:       []string{"You doing alright?"},
			expectedAnswers: []string{" "},
		},
		"multiple confirm questions": {
			responder: NewResponder().
				addConfirms("right?", ConfirmAffirm, ConfirmNegative, ConfirmNegative),
			questions:       []string{"You doing alright?", "is it alright?", "right?"},
			expectedAnswers: []string{"<right> ", " ", " "},
		},

		"one select question": {
			responder: NewResponder().
				AddSelect("how?", 5),
			questions:       []string{"how?"},
			expectedAnswers: []string{"<down><down><down><down><down>"},
		},
		"multiple select questions": {
			responder: NewResponder().
				addSelects("how?", 0, 1, 3),
			questions: []string{"how?", "how?", "how?"},
			expectedAnswers: []string{
				"",
				"<down>",
				"<down><down><down>",
			},
		},

		"one multiselect question": {
			responder: NewResponder().
				AddMultiSelect("how?", []int{2, 3}),
			questions:       []string{"how?"},
			expectedAnswers: []string{"<down><down> <down> "},
		},
		"multiple multiselect questions": {
			responder: NewResponder().
				addMultiSelects("how?", []int{2, 3}, []int{3, 4}),
			questions: []string{"how?", "how?"},
			expectedAnswers: []string{
				"<down><down> <down> ",
				"<down><down><down> <down> ",
			},
		},

		"one exact match": {
			responder: NewResponder().
				AddResponse("You doing alright?", "Splendid").
				MatchExact().
				AddResponse("alright?", "Yes"),
			questions:       []string{"You doing alright?"},
			expectedAnswers: []string{"Splendid"},
		},
		"one regexp match": {
			responder: NewResponder().
				AddResponse(`Y[ou]{2} doin. alr.ght\?a?`, "Splendid").
				MatchRegexp().
				AddResponse("alright?", "Yes, for sure!").
				MatchExact(),
			questions:       []string{"You doing alright?"},
			expectedAnswers: []string{"Splendid"},
		},
	}

	for name, testData := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			// Act
			stdin, stdout, closer := testData.responder.Start(t, defaultTimeout)

			// Assert
			defer closer()

			actualAnswers := simulateCLI(t, testData.questions, stdout, stdin)

			assert.Equal(t, testData.expectedAnswers, actualAnswers)
		})
	}
}

func TestResponder_Start_FailsTestIfCalledNotOnceOnOnce(t *testing.T) {
	t.Parallel()
	// Arrange
	responder := NewResponder().
		AddResponse("right?", "left!").
		RespondOnce()

	questions := []string{"right?", "right?"}
	expectedAnswers := []string{"left!", "left!"}

	dummyT := new(testingi.RuntimeT)

	// Act
	stdin, stdout, closer := responder.Start(dummyT, defaultTimeout)

	// Assert
	defer closer()

	actualAnswers := simulateCLI(t, questions, stdout, stdin)

	// It should still return the answers
	assert.Equal(t, expectedAnswers, actualAnswers)

	assert.True(t, dummyT.Failed(), "Test should have failed")
}

func TestResponder_Start_FailsTestIfCalledMoreThanTimes(t *testing.T) {
	t.Parallel()
	// Arrange
	responder := NewResponder().
		AddResponse("right?", "left!").
		RespondTimes(3)

	questions := []string{"right?", "right?", "right?", "right?", "right?"}
	expectedAnswers := []string{"left!", "left!", "left!", "left!", "left!"}

	dummyT := new(testingi.RuntimeT)

	// Act
	stdin, stdout, closer := responder.Start(dummyT, defaultTimeout)

	// Assert
	defer closer()

	actualAnswers := simulateCLI(t, questions, stdout, stdin)

	// It should still return the answers
	assert.Equal(t, expectedAnswers, actualAnswers)

	assert.True(t, dummyT.Failed(), "Test should have failed")
}

func TestResponder_Start_TimeoutClosesPipesAndFailsTest(t *testing.T) {
	t.Parallel()
	// Arrange
	responder := NewResponder()

	dummyT := new(testingi.RuntimeT)

	// Act
	formInput, formOutput, cancel := responder.Start(dummyT, 0)

	// Assert
	defer cancel()

	_, readErr := formInput.Read([]byte{})
	require.ErrorIs(t, readErr, io.ErrClosedPipe)

	_, writeErr := formOutput.Write([]byte{})
	require.ErrorIs(t, writeErr, io.ErrClosedPipe)

	assert.True(t, dummyT.Failed(), "Test should have failed")
}

func TestResponder_Start_CancelFunctionClosesPipes(t *testing.T) {
	t.Parallel()
	// Arrange
	responder := NewResponder()

	// Act
	formInput, formOutput, cancel := responder.Start(new(testing.T), defaultTimeout)

	// Assert
	cancel()

	_, readErr := formInput.Read([]byte{})
	require.ErrorIs(t, readErr, io.ErrClosedPipe)

	_, writeErr := formOutput.Write([]byte{})
	require.ErrorIs(t, writeErr, io.ErrClosedPipe)
}
