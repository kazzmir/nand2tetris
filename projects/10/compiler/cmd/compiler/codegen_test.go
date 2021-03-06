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

func TestMethodCallSelf(test *testing.T){
    text := `
class p {
    method int foo(int a){
        return a;
    }

    method void bar(){
        do foo(2);
        return;
    }
}
`
    generated, err := doCodeGen(text)
    if err != nil {
        test.Fatalf("could not generate code: %v", err)
    }

    expected := []string{
        "function p.foo 0",
        "push argument 0",
        "pop pointer 0",
        "push argument 1",
        "return",
        "function p.bar 0",
        "push argument 0",
        "pop pointer 0",
        "push pointer 0",
        "push constant 2",
        "call p.foo 2",
        "pop temp 0",
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

func TestWhile(test *testing.T){
    text := `
class y {
   function void main() {
     var int length;
     var int i;
     
     let i = 0;
     while (i < length) {
        let i = i + 1;
     }
     
     return;
   }
}
`
    generated, err := doCodeGen(text)
    if err != nil {
        test.Fatalf("could not generate code: %v", err)
    }

    expected := []string{
        "function y.main 2",
        "push constant 0",
        "pop local 1",
        "label while_start_0",
        "push local 1",
        "push local 0",
        "lt",
        "not",
        "if-goto while_end_1",
        "push local 1",
        "push constant 1",
        "add",
        "pop local 1",
        "goto while_start_0",
        "label while_end_1",
        "push constant 0",
        "return",
    }

    if !compareCode(generated, expected) {
        test.Fatalf("unexpected generated code: actual %v vs expected %v\n", generated, expected)
    }
}

func TestDoOutputString(test *testing.T){
    text := `
class y {
    function void main() {
      do Output.printString("abc");
      return;
    }
}
`
    generated, err := doCodeGen(text)
    if err != nil {
        test.Fatalf("could not generate code: %v", err)
    }

    expected := []string{
        "function y.main 0",
        "push constant 3",
        "call String.new 1",
        "push constant 97",
        "call String.appendChar 2",
        "push constant 98",
        "call String.appendChar 2",
        "push constant 99",
        "call String.appendChar 2",
        "call Output.printString 1",
        "pop temp 0",
        "push constant 0",
        "return",
    }

    if !compareCode(generated, expected) {
        test.Fatalf("unexpected generated code: actual %v vs expected %v\n", generated, expected)
    }
}

func TestArray(test *testing.T){
    text := `
class y {
    function void main() {
        var Array a;
        var int sum;

        let a = Array.new(3);
        let sum = sum + a[1];
        let a[2] = sum + a[0];
        return;
    }
}
`
    generated, err := doCodeGen(text)
    if err != nil {
        test.Fatalf("could not generate code: %v", err)
    }

    expected := []string{
        "function y.main 2",
        "push constant 3",
        "call Array.new 1",
        "pop local 0",
        "push local 1",
        "push local 0",
        "push constant 1",
        "add",
        "pop pointer 1",
        "push that 0",
        "add",
        "pop local 1",
        "push local 1",
        "push local 0",
        "push constant 0",
        "add",
        "pop pointer 1",
        "push that 0",
        "add",
        "push local 0",
        "push constant 2",
        "add",
        "pop pointer 1",
        "pop that 0",
        "push constant 0",
        "return",
    }

    if !compareCode(generated, expected) {
        test.Fatalf("unexpected generated code: actual %v vs expected %v\n", generated, expected)
    }
}

func TestBool(test *testing.T){
    text := `
class y {
    function void main() {
        var String a;
        var boolean b;
        let a = null;
        let b = true;
        let b = false;
        return;
    }
}
`
    generated, err := doCodeGen(text)
    if err != nil {
        test.Fatalf("could not generate code: %v", err)
    }

    expected := []string{
        "function y.main 2",
        "push constant 0",
        "pop local 0",
        "push constant 0",
        "not",
        "pop local 1",
        "push constant 0",
        "pop local 1",
        "push constant 0",
        "return",
    }

    if !compareCode(generated, expected) {
        test.Fatalf("unexpected generated code: actual %v vs expected %v\n", generated, expected)
    }
}

func TestStatic(test *testing.T){
    text := `
class y {
    static int z;
    function void main() {
        var int a;
        let z = 2;
        let a = z + 1;
        return;
    }
}
`
    generated, err := doCodeGen(text)
    if err != nil {
        test.Fatalf("could not generate code: %v", err)
    }

    expected := []string{
        "function y.main 1",
        "push constant 2",
        "pop static 0",
        "push static 0",
        "push constant 1",
        "add",
        "pop local 0",
        "push constant 0",
        "return",
    }

    if !compareCode(generated, expected) {
        test.Fatalf("unexpected generated code: actual %v vs expected %v\n", generated, expected)
    }
}

func TestMethodCall(test *testing.T){
    text := `
class y {
    function void main() {
      var SquareGame game;
      let game = SquareGame.new();
      do game.run();
      do game.dispose();
      return;
    }
}
`
    generated, err := doCodeGen(text)
    if err != nil {
        test.Fatalf("could not generate code: %v", err)
    }

    expected := []string{
        "function y.main 1",
        "call SquareGame.new 0",
        "pop local 0",
        "push local 0",
        "call SquareGame.run 1",
        "pop temp 0",
        "push local 0",
        "call SquareGame.dispose 1",
        "pop temp 0",
        "push constant 0",
        "return",
    }

    if !compareCode(generated, expected) {
        test.Fatalf("unexpected generated code: actual %v vs expected %v\n", generated, expected)
    }
}

func TestIf(test *testing.T){
    text := `
class y {
    function void test() {
        var int i;
        if (false) {
            let i = 1;
        } else {
            let i = 2;
        }

        if (true) {
            let i = 3;
        }

        return;
    }
}
`
    generated, err := doCodeGen(text)
    if err != nil {
        test.Fatalf("could not generate code: %v", err)
    }

    expected := []string{
        "function y.test 1",
        "push constant 0",
        "not",
        "if-goto if_else_0",
        "push constant 1",
        "pop local 0",
        "goto if_done_1",
        "label if_else_0",
        "push constant 2",
        "pop local 0",
        "label if_done_1",
        "push constant 0",
        "not",
        "not",
        "if-goto if_done_3",
        "push constant 3",
        "pop local 0",
        "label if_done_3",
        "push constant 0",
        "return",
    }

    if !compareCode(generated, expected) {
        test.Fatalf("unexpected generated code: actual %v vs expected %v\n", generated, expected)
    }
}

func TestConstructor(test *testing.T){
    text := `
class p {
    field int a, b;

    constructor p new(){
      let a = 2;
      let b = 3;
      return this;
    }
}
`
    generated, err := doCodeGen(text)
    if err != nil {
        test.Fatalf("could not generate code: %v", err)
    }

    expected := []string{
        "function p.new 0",
        "push constant 2",
        "call Memory.alloc 1",
        "pop pointer 0",
        "push constant 2",
        "pop this 0",
        "push constant 3",
        "pop this 1",
        "push pointer 0",
        "return",
    }

    if !compareCode(generated, expected) {
        test.Fatalf("unexpected generated code: actual %v vs expected %v\n", generated, expected)
    }
}
