package huhtest

import (
	"bufio"
	"io"
	"slices"
	"strings"
	"time"

	testingi "github.com/mitchellh/go-testing-interface"
)

// Reference: https://www.alanwood.net/demos/ansi.html
const (
	// defaultSubmit is appended to all responses to move to the next one. These represent \r\n.
	defaultSubmit = "\x0D\x0A"

	// selectSubmit is a special case where the defaultSubmit messes up the input in select statements
	selectSubmit = "\x0D"

	// selectOption is used in a select and multiselect to mark or unmark an item
	selectOption = "\x20"

	// arrowDown is used in a select and multiselect to move downwards
	arrowDown = "\x1b[B"

	// arrowRight is used in a confirm to move between yes and no
	arrowRight = "\x1b[C"
)

// readableReplacer is used primarily for logging to represent awkward
// characters with a readable representation
var readableReplacer = strings.NewReplacer(
	defaultSubmit, "<submit>",
	arrowDown, "<down>",
	arrowRight, "<right>",
)

// NewResponder instantiates a Responder that allows you to build responses
// to a commandline application. The Start() command should be called at
// the end of a chain to start a goroutine that will read and write using the returned
// io.Pipe objects.
//
// For example:
//
//	stdIn, stdOut, cancel := NewResponder().
//	  AddResponse(...),
//	  AddConfirm(...),
//	  Start()
//	defer cancel()
//
//	myForm.WithInput(stdIn).WithOutput(stdOut).Run()
//
// Check out the individual method descriptions to learn more.
func NewResponder() *Responder {
	return &Responder{
		latestQuestionMatchType: defaultQuestionMatchType,
		latestResponse:          new(response),

		responses: newResponses(),
	}
}

// NewResponderWith is a lot like NewResponder, but initialises the given questions and responses automatically.
// This only works for simple responses, select, multi-select and confirms have to be added manually afterwards.
// Please refer to the documentation of NewResponder to learn more about its usages.
func NewResponderWith(questionResponses map[string][]string) *Responder {
	responder := NewResponder()

	for question, answers := range questionResponses {
		responder.addResponses(question, answers...)
	}

	return responder
}

// Responder is a builder that allows you to put together a list of responses
// to questions asked in a form. Check out NewResponder for more information.
type Responder struct {
	// Modifiers can be applied to a question after the question has been registered,
	// so we keep the data here and only save it when the next question gets added, or the Start
	// method is called.

	latestQuestion          string
	latestQuestionMatchType questionMatchType
	latestResponse          *response

	// debug can be flipped to increase debugging in the Start method
	debug bool

	responses *responses
}

/**
* Types of responses
 */

// AddResponse adds a text-based response to the responder that will be returned if the question matches.
// If the same question comes up multiple times, the same response will be returned by default. Use Times()
// or Once() to modify this behaviour and register an error.
//
// Multiple answers to the same question can be added by repeating this call.
func (r *Responder) AddResponse(question string, answer string) *Responder {
	return r.addResponses(question, answer)
}

// AddResponse adds multiple-based responses to the responder that will be returned if the question matches.
// If the same question comes up multiple times, the next response in the list will be picked. If we
// run out of responses, the last response will be returned.
//
// NOTICE: This method is currently not exported, might consider doing this later
func (r *Responder) addResponses(question string, answers ...string) *Responder {
	r.saveResponse()

	r.latestQuestion = question
	r.latestResponse.answers = append(r.latestResponse.answers, answers...)

	return r
}

// AddSelect adds a response that will navigate a multiple-choice list and pick the index of the given option.
// If the same question comes up multiple times, the same response will be returned by default. Use Times()
// or Once() to modify this behaviour and register an error.
//
// Multiple answers to the same question can be added by repeating this call.
func (r *Responder) AddSelect(question string, option int) *Responder {
	return r.addSelects(question, option)
}

// AddSelect adds a response that will navigate a multiple-choice list and pick the index of the given option.
// If the same question comes up multiple times, the next response in the list will be picked. If we
// run out of responses, the last response will be returned.
//
// NOTICE: This method is currently not exported, might consider doing this later
func (r *Responder) addSelects(question string, options ...int) *Responder {
	r.saveResponse()

	r.latestQuestion = question
	r.latestResponse.submitCharacterOverride = selectSubmit

	for _, optionIndex := range options {
		r.latestResponse.answers = append(r.latestResponse.answers, strings.Repeat(arrowDown, optionIndex))
	}

	return r
}

// AddMultiSelect adds a response that will navigate a multiple-choice list and pick the indexes of the given options.
// If the same question comes up multiple times, the same response will be returned by default. Use Times()
// or Once() to modify this behaviour and register an error.
//
// Multiple answers to the same question can be added by repeating this call.
func (r *Responder) AddMultiSelect(question string, options []int) *Responder {
	return r.addMultiSelects(question, options)
}

// AddMultiSelect adds a response that will navigate a multiple-choice list and pick the indexes of the given options.
// If the same question comes up multiple times, the next response in the list will be picked. If we
// run out of responses, the last response will be returned.
//
// NOTICE: This method is currently not exported, might consider doing this later
func (r *Responder) addMultiSelects(question string, options ...[]int) *Responder {
	r.saveResponse()

	r.latestQuestion = question

	var answer strings.Builder

	for _, option := range options {
		for index := range option[len(option)-1] + 1 {
			if slices.Contains(option, index) {
				answer.WriteString(selectOption)
			}

			answer.WriteString(arrowDown)
		}

		// Remove trailing down arrows
		input := strings.TrimSuffix(answer.String(), arrowDown)

		r.latestResponse.answers = append(r.latestResponse.answers, input)

		answer.Reset()
	}

	return r
}

// ConfirmResponse could have been a boolean, but I wanted to be more semantic by using Affirm and Negative like huh does it.
type ConfirmResponse string

const (
	// ConfirmAffirm is the 'yes' answer in a Confirm question
	ConfirmAffirm ConfirmResponse = "yes"

	// ConfirmNegative is the 'no' answer in a Confirm question
	ConfirmNegative ConfirmResponse = "no"
)

// AddConfirm adds a confirm response to the Responder. If the same question comes up multiple times, the same response will be returned by default. Use Times()
// or Once() to modify this behaviour and register an error.
//
// Multiple answers to the same question can be added by repeating this call.
func (r *Responder) AddConfirm(question string, answer ConfirmResponse) *Responder {
	return r.addConfirms(question, answer)
}

// AddConfirm adds a confirm response to the Responder. If the same question comes up multiple times, the next response in the list will be picked. If we
// run out of responses, the last response will be returned.
//
// Multiple answers to the same question can be added by repeating this call.
//
// NOTICE: This method is currently not exported, might consider doing this later
func (r *Responder) addConfirms(question string, answers ...ConfirmResponse) *Responder {
	r.saveResponse()

	r.latestQuestion = question

	for _, answer := range answers {
		switch answer {
		case ConfirmNegative:
			r.latestResponse.answers = append(r.latestResponse.answers, " ")
		case ConfirmAffirm:
			r.latestResponse.answers = append(r.latestResponse.answers, arrowRight+" ")
		}
	}

	return r
}

/**
* Helpers
 */

// saveResponse should be called when adding a new response to finalise the previous response. We have to do this because
// a question can be modified after calling AddResponse, so we should only 'save' a response to the list after we're
// done composing it.
func (r *Responder) saveResponse() {
	// Guard against the small chance of Start() being immediately called after instantiation
	if r.latestQuestion == "" {
		return
	}

	r.responses.add(r.latestQuestion, r.latestQuestionMatchType, *r.latestResponse)

	r.latestQuestion = ""
	r.latestQuestionMatchType = defaultQuestionMatchType
	r.latestResponse = new(response)
}

/**
* Modifiers that change the previously registered response
 */

// MatchExact changes question matching to exactly match the output. This is only useful if you
// are 100% sure that the output line won't contain any formatting or flair,
func (r *Responder) MatchExact() *Responder {
	r.latestQuestionMatchType = questionMatchExact
	return r
}

// MatchRegexp changes question matching to treat the question as a regex
func (r *Responder) MatchRegexp() *Responder {
	r.latestQuestionMatchType = questionMatchRegexp
	return r
}

// RespondOnce will make the test error if the question is posed more than 1 once, but it will still return
// an answer.
func (r *Responder) RespondOnce() *Responder {
	r.latestResponse.expectedTimes = 1
	return r
}

// RespondTimes will make the test error if the question is posed more than the specified amount of times, but it will still return an answer.
func (r *Responder) RespondTimes(times int) *Responder {
	r.latestResponse.expectedTimes = times
	return r
}

/**
 * 'Other' methods
 */

// Closer is returned from Start and should be called in a defer after calling Start
type Closer func()

// Start will kick off the goroutine that will listen for inputs in the returned io.PipeWriter. It will
// then attempt to answer the incoming line with a registered response on the io.PipeReader. You're required
// to provide a timeout that will stop the reader and writer to prevent it from locking forever.
//
// To stop the responder, you can call the returned cancel/close function that will close the readers and
// writers
//
// Usage:
//
//	stdIn, stdOut, cancel := NewResponder().
//	  AddResponse(...),
//	  AddConfirms(...),
//	  Start()
//	defer cancel()
//
//	myForm.WithInput(stdIn).WithOutput(stdOut).Run()
func (r *Responder) Start(t testingi.T, timeout time.Duration) (*io.PipeReader, *io.PipeWriter, Closer) {
	t.Helper()

	r.saveResponse()

	formStdIn, answerInput := io.Pipe()
	questionOutput, formStdOut := io.Pipe()

	// Avoids having to put if-statements everywhere
	log := func(input ...any) {
		// If the test has already failed, we could cause a panic
		if r.debug && !t.Failed() {
			t.Log(input...)
		}
	}

	go func() {
		lineReader := bufio.NewScanner(questionOutput)

		for lineReader.Scan() {
			line := lineReader.Text()

			log("Got line:", line)

			response, question, ok := r.responses.find(line)

			if ok {
				log("Matches question:", question)

				answer, err := response.pickAnswer()
				if err != nil {
					t.Error(err)
				}

				log("Replying:", readableReplacer.Replace(answer))

				_, err = answerInput.Write([]byte(answer))
				if err != nil {
					t.Error(err)
				}

				time.Sleep(20 * time.Millisecond)

				log("Sending submit character...")
				if _, err := answerInput.Write([]byte(response.submitCharacter())); err != nil {
					t.Error(err)
				}

				continue
			}
		}
	}()

	closer := func() {
		answerInput.Close()
		questionOutput.Close()

		formStdIn.Close()
		formStdOut.Close()
	}

	go func() {
		time.Sleep(timeout)
		t.Error("Deadline reached, closing readers and writers")
		closer()
	}()

	return formStdIn, formStdOut, closer
}

// Debug turns on logging for debugging forms
func (r *Responder) Debug() *Responder {
	r.debug = true
	return r
}
