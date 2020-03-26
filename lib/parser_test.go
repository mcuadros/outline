package lib

import (
	"bytes"
	"testing"

	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/stretchr/testify/assert"
)

var differ = diffmatchpatch.New()

const twoFuncsTabs = `outline: twoFuncs
	path: twoFuncs
	functions:
		difference(a,b int) int
		sum(a,b int) int
			add two things together`

const twoFuncsSpaces = `outline: twoFuncs
	path: twoFuncs
  functions:
    difference(a,b int) int
    sum(a,b int) int
			add two things together`

var twoFuncs = &Doc{
	Name: "twoFuncs",
	Path: "twoFuncs",
	Functions: []*Function{
		{
			Signature: "difference(a,b int) int",
		},
		{
			Signature:   "sum(a,b int) int",
			Description: "add two things together",
		},
	},
}

const timeSpaces = `here's some leading gak that shouldn't get read into the doc

outline: time
  functions:
    duration(string) duration
      parse a duration
    time(string, format=..., location=...) time
      parse a time
    now() time
      new time instance set to current time
      implementations are able to make this a constant
    zero() time
      a constant

  types:
    duration
      a period of time
      methods:
        add(d duration) int
          params:
             d duration
      fields:
        hours float
          number of hours starting at zero
        minutes float
        nanoseconds int
        seconds float
      operators:
        duration - time = duration
        duration + time = time
        duration == duration = boolean
        duration < duration = booleans
    time
      fields:
      operators:
        time == time = boolean
        time < time = boolean`

var time = &Doc{
	Name: "time",
	Functions: []*Function{
		{Signature: "duration(string) duration",
			Description: "parse a duration"},
		{Signature: "time(string, format=..., location=...) time",
			Description: "parse a time"},
		{Signature: "now() time",
			Description: "new time instance set to current time implementations are able to make this a constant"},
		{Signature: "zero() time",
			Description: "a constant"},
	},
	Types: []*Type{
		{Name: "duration",
			Description: "a period of time",
			Methods: []*Function{
				{Signature: "add(d duration) int",
					Params: []*Param{
						{Name: "d", Type: "duration"},
					},
				},
			},
			Fields: []*Field{
				{Name: "hours float", Description: "number of hours starting at zero"},
				{Name: "minutes float"},
				{Name: "nanoseconds int"},
				{Name: "seconds float"},
			},
			Operators: []*Operator{
				{Opr: "duration - time = duration"},
				{Opr: "duration + time = time"},
				{Opr: "duration == duration = boolean"},
				{Opr: "duration < duration = booleans"},
			},
		},
		{Name: "time",
			Fields: []*Field{},
			Operators: []*Operator{
				{Opr: "time == time = boolean"},
				{Opr: "time < time = boolean"},
			},
		},
	},
}

const docWithDescriptionTabs = `outline: doc
	this is a document description.
	It's written across two lines
	functions:
		sum(a int, b int) int`

var docWithDescription = &Doc{
	Name:        "doc",
	Description: "this is a document description. It's written across two lines",
	Functions: []*Function{
		{Signature: "sum(a int, b int) int"},
	},
}

const huhSpaces = `  outline: huh
  huh is a package that has no meaning or purpose
  functions:
    foo(bar string) int
      foo a bar, which is to to a bar and remove 'd' from 'food'
      params:
        bar string
          the name of a bar
    date() date
      make a date`

var huh = &Doc{
	Name:        "huh",
	Description: "huh is a package that has no meaning or purpose",
	Functions: []*Function{
		{Signature: "foo(bar string) int",
			Description: "foo a bar, which is to to a bar and remove 'd' from 'food'",
			Params: []*Param{
				{Name: "bar", Type: "string", Description: "the name of a bar"},
			},
		},
		{Signature: "date() date",
			Description: "make a date"},
	},
}

func TestParse(t *testing.T) {
	cases := []struct {
		in  string
		exp *Doc
		err string
	}{
		{"outline: foo\n", &Doc{Name: "foo"}, ""},
		{twoFuncsTabs, twoFuncs, ""},
		{twoFuncsSpaces, twoFuncs, ""},
		{timeSpaces, time, ""},
		{docWithDescriptionTabs, docWithDescription, ""},
		{huhSpaces, huh, ""},
	}

	for i, c := range cases {
		b := bytes.NewBufferString(c.in)
		got, err := ParseFirst(b)
		if !(err == nil && c.err == "" || err != nil && err.Error() == c.err) {
			t.Errorf("case %d error mismatch. expected: %s, got: %s", i, c.err, err)
			continue
		}

		if got == nil {
			t.Errorf("case %d doc returned nil", i)
			continue
		}

		gotB, _ := got.MarshalIndent(0, "  ")
		expB, _ := c.exp.MarshalIndent(0, "  ")
		if string(expB) != string(gotB) {
			t.Errorf("case %d equality mismatch. expected:\n%s\ngot:\n%s\n", i, string(expB), string(gotB))
			t.Log("\n", gotB, "\n", expB)
			diffs := differ.DiffMain(string(gotB), string(expB), true)
			t.Log(differ.DiffPrettyText(diffs))
		}
	}
}

const funcExamples = `outline: examples
  huh is a package that has no meaning or purpose
  functions:
    foo(bar string) int
      foo a bar, which is to to a bar and remove 'd' from 'food'
      examples:
        foo.star Foo Example
        bar.star 
          Description of the bar example
      params:
        bar string
          the name of a bar`

func TestFunctionExamples(t *testing.T) {
	b := bytes.NewBufferString(funcExamples)
	doc, err := ParseFirst(b)
	assert.NoError(t, err)

	assert.Equal(t, doc.Functions.Len(), 1)

	foo := doc.Functions[0]
	assert.Len(t, foo.Examples, 2)
	assert.Equal(t, foo.Examples[0].Name, "Foo Example")
	assert.Equal(t, foo.Examples[0].Filename, "foo.star")
	assert.Equal(t, foo.Examples[0].Description, "")

	assert.Equal(t, foo.Examples[1].Name, "")
	assert.Equal(t, foo.Examples[1].Filename, "bar.star")
	assert.Equal(t, foo.Examples[1].Description, "Description of the bar example")
}

const typeExamples = `outline: examples
types:
  duration
    a period of time
    examples:
      foo.star Foo Example
    methods:
      add(d duration) int
        params:
          d duration`

func TestTypeExamples(t *testing.T) {
	b := bytes.NewBufferString(typeExamples)
	doc, err := ParseFirst(b)
	assert.NoError(t, err)

	assert.Equal(t, doc.Types.Len(), 1)

	foo := doc.Types[0]
	assert.Len(t, foo.Examples, 1)
	assert.Equal(t, foo.Examples[0].Name, "Foo Example")
	assert.Equal(t, foo.Examples[0].Filename, "foo.star")
	assert.Equal(t, foo.Examples[0].Description, "")
}

const typeDescriptionMultiLine = `outline: examples
types:
  duration
    line 1.
    line 2.

    line 3.
    examples:
      foo.star Foo Example`

func TestDescriptionMultiLine(t *testing.T) {
	b := bytes.NewBufferString(typeDescriptionMultiLine)
	doc, err := ParseFirst(b)
	assert.NoError(t, err)

	assert.Equal(t, doc.Types.Len(), 1)

	foo := doc.Types[0]
	assert.Equal(t, foo.Description, "line 1.\nline 2.\n\nline 3.")
	assert.Len(t, foo.Examples, 1)
	assert.Equal(t, foo.Examples[0].Name, "Foo Example")
	assert.Equal(t, foo.Examples[0].Filename, "foo.star")
	assert.Equal(t, foo.Examples[0].Description, "")
}
