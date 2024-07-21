# 🤔 Huh Test

_Library is still under development, but feel free to try it out_

`huhtest` is a work-in-progress test library for your [huh](https://github.com/charmbracelet/huh) forms
If you're - for some reason - eager to test your huh-based interactive CLI applications then you've come to
the right place.

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
    Start(t, 1 * time.Second)

  defer cancel()

  // Act
  err := myForm.WithInput(formInput).WithOutput(formOutput).Run()

  // Assert
  // ...
}
```

## 🔭 Plans

- Custom keymap support for select fields
- Get multiple inputs in a group working
- Verify whether tests panic if calling t.Error or t.Log after timeout ends
