package naming

import (
	"reflect"
	"testing"
)

func TestSplitWords(t *testing.T) {
	cases := []struct {
		in   string
		want []string
	}{
		{"http_status", []string{"http", "status"}},
		{"HTTPStatus", []string{"HTTP", "Status"}},
		{"redOrange", []string{"red", "Orange"}},
		{"in-progress", []string{"in", "progress"}},
		{"v1", []string{"v1"}},
		{"iPhone", []string{"i", "Phone"}},
		{"", []string{}},
	}
	for _, tc := range cases {
		if got := SplitWords(tc.in); !reflect.DeepEqual(got, tc.want) {
			t.Errorf("SplitWords(%q) = %#v, want %#v", tc.in, got, tc.want)
		}
	}
}

func TestCamel(t *testing.T) {
	init := InitialismSet(nil, false)
	cases := map[string]string{
		"red":         "Red",
		"http_status": "HTTPStatus",
		"api_url":     "APIURL",
		"uuid":        "UUID",
		"blue-green":  "BlueGreen",
		"not_found":   "NotFound",
		"in_progress": "InProgress",
	}
	for in, want := range cases {
		if got := Camel(in, init); got != want {
			t.Errorf("Camel(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestCamelWithoutInitialisms(t *testing.T) {
	init := InitialismSet(nil, true)
	if got := Camel("http_status", init); got != "HttpStatus" {
		t.Fatalf("Camel = %q", got)
	}
}

func TestSnakeKebabTitle(t *testing.T) {
	if got := Snake("HTTPStatus"); got != "http_status" {
		t.Fatalf("Snake = %q", got)
	}
	if got := Kebab("in_progress"); got != "in-progress" {
		t.Fatalf("Kebab = %q", got)
	}
	if got := Title("http_status"); got != "Http Status" {
		t.Fatalf("Title = %q", got)
	}
}

func TestInitialismSetExtra(t *testing.T) {
	init := InitialismSet([]string{"grpc"}, false)
	if !init["GRPC"] {
		t.Fatal("extra initialism missing")
	}
	if got := Camel("grpc_status", init); got != "GRPCStatus" {
		t.Fatalf("Camel = %q", got)
	}
}
