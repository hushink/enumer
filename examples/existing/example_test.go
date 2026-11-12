package existing

import "testing"

func TestExistingType(t *testing.T) {
	c, err := ParseColor("green")
	if err != nil || c != ColorGreen {
		t.Fatalf("parse: %v %v", c, err)
	}
	if ColorBlue.String() != "blue" {
		t.Fatalf("string: %q", ColorBlue.String())
	}
}
