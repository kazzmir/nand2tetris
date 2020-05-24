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

func TestSimpleMath(test *testing.T){
    text := `
class x {
  function int bar(int m) {
    return m + 2;
  }

  function void foo(){
    var int z, y, a, b, c;
    let y = 1;
    let z = x.bar(y);
    let a = y + z;
    let b = z - z;
    let c = y / (z + 1);
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
        "function x.foo 5",
        "push constant 1",
        "pop local 1",
        "push local 1",
        "call x.bar 1",
        "pop local 0",
        "push local 1",
        "push local 0",
        "add",
        "pop local 2",
        "push local 0",
        "push local 0",
        "sub",
        "pop local 3",
        "push local 1",
        "push local 0",
        "push constant 1",
        "add",
        "call Math.divide 2",
        "pop local 4",
        "push constant 0",
        "return",
    }

    if !compareCode(generated, expected) {
        test.Fatalf("unexpected generated code: actual %v vs expected %v\n", generated, expected)
    }
}

func TestField(test *testing.T){
    text := `
class x {
  field int field1;

  method int method1(int a){
    return field1 + a;
  }
}
`
    generated, err := doCodeGen(text)
    if err != nil {
        test.Fatalf("could not generate code: %v", err)
    }

    expected := []string{
        "function x.method1 0",
        "push argument 0",
        "pop pointer 0",
        "push this 0",
        "push argument 1",
        "add",
        "return",
    }

    if !compareCode(generated, expected) {
        test.Fatalf("unexpected generated code: actual %v vs expected %v\n", generated, expected)
    }
}
