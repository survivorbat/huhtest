# 🤔 Huh Test

`huhtest` is a work-in-progress test library for your [huh](https://github.com/charmbracelet/huh) forms
If you're - for some reason - eager to test your huh-based interactive CLI applications then you've come to
the right place.
It works by matching messages in a form's output (stdout) and then sending pre-programmed text to the form's input (stdin).

## ⬇️ Installation

`go get github.com/survivorbat/huhtest`

## 📋 Usage

```go
package main

import (
  "time"

  "github.com/survivorbat/huhtest"
)

func TestMyForm() {
  // Arrange
  myForm := huh.NewForm(/* ... */)

  stdin, stdout, cancel := huhtest.NewResponder().
    AddResponse("How Are You Feeling?", "Amazing Thanks!").
    AddConfirm("Would you like a drink?", ConfirmAffirm).
    AddSelect("Make a second choice", 2).
    AddMultiSelect("Choose all options that apply", []int{2, 5, 6}),
    Start(t, 1 * time.Second)

  defer cancel()

  // Act
  err := myForm.WithInput(formInput).WithOutput(formOutput).Run()

  // Assert
  // ...
}
```

## 🧪 Testing

To make sure this thing actually works, we have both unit tests and integration tests, the former
checks the output of `huhtest` directly, the latter actually uses `huh` to check whether the inputs
are properly processed.

## 🐞 Debugging

There's a `.Debug()` method available that enabled extra logging in the `Responser`. If you
encounter a bug or are suspicious about something not working, turn it on to see exactly what it's doing.

## 🔭 Plans

- Custom keymap support for select fields
