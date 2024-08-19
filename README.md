# 🤔 Huh Test

`huhtest` is a work-in-progress test library for your [huh](https://github.com/charmbracelet/huh) forms.
If you're - for some reason - eager to test your huh-based interactive CLI applications then you've come to
the right place.
It works by matching messages in a form's output (stdout) and then sending pre-programmed text to the form's input (stdin).

It's not 100% bug-free, as some combinations of groups and selects seem to have off-by-one errors.

## ⬇️ Installation

`go get github.com/survivorbat/huhtest`

## 📋 Usage

Check out [this example](./examples_test.go)

## 🧪 Testing

To make sure this thing actually works, we have both unit tests and integration tests, the former
checks the output of `huhtest` directly, the latter actually uses `huh` to check whether the inputs
are properly processed.

## 🐞 Debugging

There's a `.Debug()` method available that enabled extra logging in the `Responser`. If you
encounter a bug or are suspicious about something not working, turn it on to see exactly what it's doing.

## 🔭 Plans

- Custom keymap support for select fields
- Select fields should preferably be selected by output text and not index
