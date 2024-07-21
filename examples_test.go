package huhtest

// func ExampleResponder() {
// 	var howAreYouFeeling string
// 	var wouldYouLikeADrink bool
// 	var allOptionsThatApply []string
// 	var areYouOk1 string
// 	var areYouOk2 string
//
// 	myForm := huh.NewForm(
// 		huh.NewGroup(
// 			huh.NewInput().
// 				Title("How Are You Feeling?").
// 				Value(&howAreYouFeeling),
// 		),
// 		huh.NewGroup(
// 			huh.NewConfirm().
// 				Title("Would you like a drink?").
// 				Value(&wouldYouLikeADrink),
// 		),
// 		huh.NewGroup(
// 			huh.NewMultiSelect[string]().
// 				Title("Please pick all options that apply").
// 				Value(&allOptionsThatApply),
// 		),
// 		huh.NewGroup(
// 			huh.NewInput().
// 				Title("Are You OK? (gonna ask you twice)").
// 				Value(&areYouOk1),
// 		),
// 		huh.NewGroup(
// 			huh.NewInput().
// 				Title("Are You OK? (gonna ask you twice)").
// 				Value(&areYouOk2),
// 		),
// 	)
//
// 	formIn, formOut, closeResponder := NewResponder().
// 		AddResponse("How Are You Feeling?", "Amazing Thanks!").
// 		AddConfirms("Would you like a drink?", ConfirmNegative).
// 		Times(2).
// 		AddSelects("Please pick all options that apply", []string{"a", "b", "c"}).
// 		Once().
// 		AddResponses("Are you OK? (gonna ask you twice)", "yes", "yes for sure").
// 		Start()
//
// 	defer closeResponder()
//
// 	// Act
// 	_ = myForm.WithInput(formIn).WithOutput(formOut).Run()
//
// 	// Assert
// 	fmt.Println(howAreYouFeeling)
//
// 	// Output: Amazing Thanks!
// 	// Output: false
// 	// Output: ["A", "B", "c"]
// 	// Output: ["yes", "yes for sure"]
// }
