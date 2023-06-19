package main

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

const packageDef = "package foo\n"

func TestExportedMethods(t *testing.T) {
	testSrc := `
import (
	"fmt"
)

type Client string
type AnotherType int

func ExportedNonMethod(a string, b *int) error {
	return fmt.Errorf("this should be skipped")
}

func (m *Client) ExportedPointerReceiverMethod(x, y int, z *AnotherType) ([]byte, error){
	s := x+y+int(*z)
	b := make([]byte, s)
	return b
}

func (m Client) ExportedReceiverMethod(w *io.Writer, n []int) *[]string {
	a := []string{"foo", "bar"}
	return &a
}

func (t *Client) unexportedMethod () (*Client, string, int) {
	return t, "foo", 42
}

func (t *Client) AnotherExportedPointerReceiverMethod() {
	_ = t + "foo"
}
`
	expToBeIncluded := []string{"ExportedPointerReceiverMethod", "AnotherExportedPointerReceiverMethod"}
	b := bytes.Buffer{}
	f, err := parser.ParseFile(token.NewFileSet(), "testSrc", packageDef+testSrc, 0)
	if err != nil {
		t.Fatal(err)
	}
	count, err := generate(&b, map[string]*ast.File{"testSrc": f})
	if err != nil {
		t.Fatal(err)
	}
	if count != len(expToBeIncluded) {
		t.Errorf("Expected `exportedMethods` to return %d methods, got %d", len(expToBeIncluded), count)
	}
	s := b.String()
	for _, expFunc := range expToBeIncluded {
		if !strings.Contains(s, expFunc) {
			t.Errorf("Expected 'generate' to include a proxy function for %s, but it didn't", expFunc)
		}
	}

}

func TestGenFuncFromMethod(t *testing.T) {
	testCases := []struct {
		name         string
		inSrc        string
		expParamsStr string
		expArgsStr   string
		expResStr    string
	}{
		{
			name:         "1. One 'int' pram and one 'error' return",
			inSrc:        `func (c *Client) SampleMethod(x int) error {}`,
			expParamsStr: "(x int)",
			expArgsStr:   "(x)",
			expResStr:    " error",
		},
		{
			name:         "2. One '*Foo' param, one `*Client` return and one '[]byte' return",
			inSrc:        `func (c *Client) SampleMethod(f *Foo) (*Client, []byte) {}`,
			expParamsStr: "(f *Foo)",
			expArgsStr:   "(f)",
			expResStr:    " (*Client, []byte)",
		},
		{
			name:         "3. One slice param, one variadic `int` param, no return",
			inSrc:        `func (c *Client) SampleMethod(b []byte, n ...int) {}`,
			expParamsStr: "(b []byte, n ...int)",
			expArgsStr:   "(b, n...)",
			expResStr:    "",
		},
		{
			name:         "4. One `func` param, one variadic `[]int` param, one `func() int` return",
			inSrc:        `func (c *Client) SampleMethod(f func(int, int, string) string, n ...[]int) func() int {}`,
			expParamsStr: "(f func(int, int, string) string, n ...[]int)",
			expArgsStr:   "(f, n...)",
			expResStr:    " func() int",
		},
		{
			name:         "5. Two 'string' params, one variadic `func(int) error` param, two `int' returns",
			inSrc:        `func (c *Client) SampleMethod(a, b string, f ...[]func(int) error) (int, int) {}`,
			expParamsStr: "(a, b string, f ...[]func(int) error)",
			expArgsStr:   "(a, b, f...)",
			expResStr:    " (int, int)",
		},
		{
			name:         "6. One `*[]byte` param, one `[]*Client` return and one 'error' return",
			inSrc:        `func (b *Client) SampleMethod(b *[]byte) ([]*Client, error) {}`,
			expParamsStr: "(b *[]byte)",
			expArgsStr:   "(b)",
			expResStr:    " ([]*Client, error)",
		},
		{
			name:         "7. Very complex method definition",
			inSrc:        `func (b *Client) SampleMethod(f *[]*func([]byte, *[][][]float64, func(int) *Client) (*[]byte, error)) (*[]*func([]byte, *[][][]float64), *[][]*[]float64, error) {}`,
			expParamsStr: "(f *[]*func([]byte, *[][][]float64, func(int) *Client) (*[]byte, error))",
			expArgsStr:   "(f)",
			expResStr:    " (*[]*func([]byte, *[][][]float64), *[][]*[]float64, error)",
		},
		{
			name:         "8. Very simple method definition",
			inSrc:        `func (b *Client) SampleMethod() {}`,
			expParamsStr: "()",
			expArgsStr:   "()",
			expResStr:    "",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			f, err := parser.ParseFile(token.NewFileSet(), "", packageDef+tc.inSrc, 0)
			if err != nil {
				t.Fatal(err)
			}
			methods := ptrRcvMethods(f, "Client")
			if len(methods) == 0 {
				t.Fatal("Expected at least one method to work with")
			}
			par := methods[0].Type.Params
			res := methods[0].Type.Results
			gotParamsStr := genTypedParams(par)
			if gotParamsStr != tc.expParamsStr {
				t.Errorf("Unexpected string returned from `genTypedParams`.\nExpected:\n%q\nGot:\n%q", tc.expParamsStr, gotParamsStr)
			}
			gotArgsStr := genFuncArgs(par)
			if gotArgsStr != tc.expArgsStr {
				t.Errorf("Unexpected string returned from `genFuncArgs`.\nExpected:\n%q\nGot:\n%q", tc.expArgsStr, gotArgsStr)
			}
			gotResStr := genFuncReturnTypes(res)
			if gotResStr != tc.expResStr {
				t.Errorf("Unexpected string returned from `genFuncReturnTypes`.\nExpected:\n%q\nGot:\n%q", tc.expResStr, gotResStr)
			}

		})
	}
}
