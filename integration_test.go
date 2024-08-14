package huhtest

import (
	"testing"

	"github.com/charmbracelet/huh"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHuhTest_RespondsCorrectlyToQuestions(t *testing.T) {
	t.Parallel()

	type answers struct {
		input string

		groupInputA string
		groupInputB string
		groupInputC string

		confirmTrue  bool
		confirmFalse bool

		singleSelect1 string
		singleSelect2 string
		multiSelect   []string

		consecutiveQuestion1 string
		consecutiveQuestion2 string
	}

	var actual answers

	myForm := huh.NewForm(
		// Simple question
		huh.NewGroup(
			huh.NewInput().
				Title("How Are You Feeling?").
				Value(&actual.input),
		),
		// Questions in a group
		huh.NewGroup(
			huh.NewInput().
				Title("Group Question A?").
				Value(&actual.groupInputA),
			huh.NewInput().
				Title("Group Question B?").
				Value(&actual.groupInputB),
			huh.NewInput().
				Title("Group Question C?").
				Value(&actual.groupInputC),
		),
		// Confirm question with a true answer
		huh.NewGroup(
			huh.NewConfirm().
				Title("Would you like a drink?").
				Value(&actual.confirmTrue),
		),
		// Confirm question with a false answer
		huh.NewGroup(
			huh.NewConfirm().
				Title("Would you like a meal?").
				Value(&actual.confirmFalse),
		),
		// Select with first option
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Make a choice").
				Options(
					huh.NewOption("a", "a"),
					huh.NewOption("b", "b"),
					huh.NewOption("c", "c"),
					huh.NewOption("d", "d"),
					huh.NewOption("e", "e"),
					huh.NewOption("f", "f"),
				).
				Value(&actual.singleSelect1),
		),
		// Select with third option
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Make a second choice").
				Options(
					huh.NewOption("a", "a"),
					huh.NewOption("b", "b"),
					huh.NewOption("c", "c"),
					huh.NewOption("d", "d"),
					huh.NewOption("e", "e"),
					huh.NewOption("f", "f"),
				).
				Value(&actual.singleSelect2),
		),
		// Multi select
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Please pick all options that apply").
				Options(
					huh.NewOption("a", "a"),
					huh.NewOption("b", "b"),
					huh.NewOption("c", "c"),
					huh.NewOption("d", "d"),
					huh.NewOption("e", "e"),
					huh.NewOption("f", "f"),
				).
				Value(&actual.multiSelect),
		),
		// Simple question that gets asked twice (1)
		huh.NewGroup(
			huh.NewInput().
				Title("Are You OK? (gonna ask you twice)").
				Value(&actual.consecutiveQuestion1),
		),
		// Simple question that gets asked twice (2)
		huh.NewGroup(
			huh.NewInput().
				Title("Are You OK? (gonna ask you twice)").
				Value(&actual.consecutiveQuestion2),
		),
	)

	formInput, formOutput, closeResponder := NewResponder().
		// Simple question
		AddResponse("How Are You Feeling?", "Amazing Thanks!").
		// Questions in a group
		AddResponse("Group Question A?", "Foo").
		AddResponse("Group Question B?", "Bar").
		AddResponse("Group Question C?", "Baz").
		// Confirm question with a true answer
		AddConfirm("Would you like a drink?", ConfirmAffirm).
		// Confirm question with a false answer
		AddConfirm("meal?", ConfirmNegative).
		// Select with first option
		AddSelect("Make a choice", 0).
		// Select with third option
		AddSelect("Make a second choice", 2).
		// Multi select
		AddMultiSelect("Please pick all options that apply", []int{2, 3, 5}).
		// Simple question that gets asked twice (2)
		AddResponse("Are You OK? (gonna ask you twice)", "yes").
		AddResponse("Are You OK? (gonna ask you twice)", "yes for sure").
		// Better for debugging
		Debug().
		Start(t, defaultTimeout)

	defer closeResponder()

	// Act
	err := myForm.WithInput(formInput).WithOutput(formOutput).Run()

	// Assert
	require.NoError(t, err)

	expected := answers{
		input:                "Amazing Thanks!",
		groupInputA:          "Foo",
		groupInputB:          "Bar",
		groupInputC:          "Baz",
		confirmTrue:          true,
		confirmFalse:         false,
		singleSelect1:        "a",
		singleSelect2:        "c",
		multiSelect:          []string{"c", "d", "f"},
		consecutiveQuestion1: "yes",
		consecutiveQuestion2: "yes for sure",
	}

	assert.Equal(t, expected, actual)
}
