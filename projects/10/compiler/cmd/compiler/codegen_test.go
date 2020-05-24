package main

import (
    "testing"
    "strings"
)

func filterEmpty(values []string) []string {
    var out []string

    for _, line := range values {
        if strings.TrimSpace(line) != "" {
            out = append(out, line)
        }
    }

    return out
}

func doCodeGen(text string) ([]string, error) {
    ast, err := parse(strings.NewReader(text))
    if err != nil {
        return nil, err
    }

    var out strings.Builder

    err = GenerateCode(ast, &out)
    if err != nil {
        return nil, err
    }

    /* there will be an extra blank line at the end, so we filter it out */
    return filterEmpty(strings.Split(out.String(), "\n")), nil
}

func compareCode(actual []string, expected []string) bool {
    if len(actual) != len(expected) {
        return false
    }

    for i, actualLine := range actual {
        if actualLine != expected[i] {
            return false
        }
    }

    return true
}

func TestSimpleLet(test *testing.T){
    text := `class x {
    function void foo() {
        var int x, y;
        let y = 1;
        let x = y;
        return;
    }
}
`
    generated, err := doCodeGen(text)
    if err != nil {
        test.Fatalf("could not generate code: %v", err)
    }

    expected := []string{
        "function x.foo 2",
        "push constant 1",
        "pop local 1",
        "push local 1",
        "pop local 0",
        "push constant 0",
        "return",
    }

    if !compareCode(generated, expected) {
        test.Fatalf("unexpected generated code: actual %v vs expected %v\n", generated, expected)
    }
}

func TestFunctionCall(test *testing.T){
    text := `
class x {
  function int bar(int m) {
    return m + 2;
  }

  function void foo(){
    var int z, y;
    let y = 1;
    let z = x.bar(y);
    return;
  }
}
`
    generated, err := doCodeGen(text)
    if err != nil {
        test.Fatalf("could not generate code: %v", err)
    }

    expected := []string{
        "function x.bar 0",
        "push argument 0",
        "push constant 2",
        "add",
        "return",
        "function x.foo 2",
        "push constant 1",
        "pop local 1",
        "push local 1",
        "call x.bar 1",
        "pop local 0",
        "push constant 0",
        "return",
    }

    if !compareCode(generated, expected) {
        test.Fatalf("unexpected generated code: actual %v vs expected %v\n", generated, expected)
    }
}
