package huhtest

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/stretchr/testify/require"
)

var t = new(testing.T)

func ExampleNewResponder() {
	// Arrange
	var (
		howAreYouFeelingAnswer string
		areYouReadyAnswer      bool
		sleptWellAnswer        string
		activitiesAnswer       []string
	)

	myForm := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("How Are You Feeling?").
				Value(&howAreYouFeelingAnswer),
			huh.NewConfirm().
				Title("Are you ready?").
				Value(&areYouReadyAnswer),
			huh.NewSelect[string]().
				Title("Have you slept well?").
				Options(
					huh.NewOption("Well!", "well"),
					huh.NewOption("Terribly!", "terrible"),
					huh.NewOption("It was OK!", "ok"),
				).
				Value(&sleptWellAnswer),
		),
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("What are your favourite activities?").
				Options(
					huh.NewOption("Cycling", "bike"),
					huh.NewOption("Sleeping", "sleep"),
					huh.NewOption("Boating", "boat"),
					huh.NewOption("Gaming", "game"),
					huh.NewOption("Flying", "fly"),
				).
				Value(&activitiesAnswer),
		),
	)

	stdin, stdout, cancel := NewResponder().
		AddResponse("How Are You Feeling?", "Great").
		AddConfirm("Are you ready?", ConfirmAffirm).
		AddMultiSelect("activities", []int{1, 2, 4}).
		AddSelect("Have you slept well", 2).
		Start(t, 1*time.Second)

	defer cancel()

	// Act
	err := myForm.WithInput(stdin).WithOutput(stdout).Run()

	// Assert
	require.NoError(t, err)

	fmt.Println("How are you feeling?", howAreYouFeelingAnswer)
	fmt.Println("Are you ready?", areYouReadyAnswer)
	fmt.Println("Have you slept well?", sleptWellAnswer)
	fmt.Println("What are your favourite activities?", strings.Join(activitiesAnswer, ", "))

	// Output:
	// How are you feeling? Great
	// Are you ready? true
	// Have you slept well? ok
	// What are your favourite activities? sleep, boat, fly
}
