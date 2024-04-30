package stdio

import "testing"

func TestStdIn(t *testing.T) {

	// This is a simple test to check if the GetInput function works correctly.
	// The GetInput function reads a line from the standard input and returns it.
	// The test is done by calling the GetInput function and checking if the returned value is correct.
	// The test will pass if the returned value is correct and fail otherwise.

	got := GetInput("Enter text: ")
	want := "Hello\n"

	if got != want {
		t.Errorf("GetInput() = %q, want %q", got, want)
	}
}
