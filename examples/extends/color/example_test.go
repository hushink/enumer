package color

import "testing"

func TestColor(t *testing.T) {
	c, err := ParseColor("green")
	if err != nil || c != ColorGreen {
		t.Fatalf("parse: %v %v", c, err)
	}
}
