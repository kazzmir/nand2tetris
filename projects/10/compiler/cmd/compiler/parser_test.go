package main

import (
    "testing"
    "strings"
    "fmt"
)

func TestClassParse(test *testing.T){
    text := "class Foo { }"
    class, err := parse(strings.NewReader(text))
    if err != nil {
        test.Fatalf("could not parse %v: %v", text, err)
    }

    if class.Kind() != ASTKindClass {
        test.Fatalf("did not get a class ast: %v", class)
    }
}

func TestVar(test *testing.T){
    text := "var int a, b, c;"

    tokens := make(chan Token, 1000)

    var lexerError error

    go func(){
        lexerError = standardLexer(strings.NewReader(text), tokens)
    }()

    stream := &TokenStream{
        tokens: tokens,
        hasNext: false,
    }

    var_, err := parseVarDeclaration(stream)

    if lexerError != nil {
        test.Fatalf("failed to lex var: %v", lexerError)
    }

    if err != nil {
        test.Fatalf("could not parse var: %v", err)
    }

    if var_.Kind() != ASTKindVar {
        test.Fatalf("did not parse var: %v", var_)
    }

    if len(var_.Names) != 3 {
        test.Fatalf("did not parse 3 variable names: %v", len(var_.Names))
    }
}

func TestLetExpression(test *testing.T){
    text := "let game = SquareGame.new();"

    tokens := make(chan Token, 1000)

    var lexerError error

    go func(){
        lexerError = standardLexer(strings.NewReader(text), tokens)
    }()

    stream := &TokenStream{
        tokens: tokens,
        hasNext: false,
    }

    let, err := parseLet(stream)

    if lexerError != nil {
        test.Fatalf("failed to lex var: %v", lexerError)
    }

    if err != nil {
        test.Fatalf("could not parse let: %v", err)
    }

    if let.Kind() != ASTKindLet {
        test.Fatalf("did not parse let: %v", let)
    }
}

func testExpression(text string) error {
    tokens := make(chan Token, 1000)

    var lexerError error

    go func(){
        lexerError = standardLexer(strings.NewReader(text), tokens)
    }()

    stream := &TokenStream{
        tokens: tokens,
        hasNext: false,
    }

    expression, err := parseExpression(stream)

    if lexerError != nil {
        return lexerError
    }

    if err != nil {
        return err
    }

    /* make sure there are no tokens left */
    for {
        token, empty := stream.Consume()
        if empty != nil {
            break
        }
        if token.Kind != TokenSemicolon {
            return fmt.Errorf("unparsed token %v", token.String())
        }
    }

    if !isExpression(expression) {
        return fmt.Errorf("did not parse an expression: %v", expression.Kind())
    }

    return nil
}

func doTestExpression(text string, test *testing.T){
    err := testExpression(text)
    if err != nil {
        test.Fatalf("could not '%v' expression: %v", text, err)
    }
}

func TestExpression(test *testing.T){
    doTestExpression("1 + 1;", test)
    doTestExpression("1 * 1 + 2;", test)
    doTestExpression("1 * (1 + 2);", test)
    doTestExpression("x * (y + z);", test)
    doTestExpression("foo();", test)
    doTestExpression("foo.bar();", test)
}

func TestSmallProgramParse(test *testing.T){
    text := `
// This file is part of www.nand2tetris.org
// and the book "The Elements of Computing Systems"
// by Nisan and Schocken, MIT Press.
// File name: projects/10/Square/Main.jack

// (derived from projects/09/Square/Main.jack, with testing additions)

/** Initializes a new Square Dance game and starts running it. */
class Main {
    static boolean test;    // Added for testing -- there is no static keyword
                            // in the Square files.
    function void main() {
      var SquareGame game;
      let game = SquareGame.new();
      do game.run();
      do game.dispose();
      return;
    }

    function void test() {  // Added to test Jack syntax that is not use in
        var int i, j;       // the Square files.
        var String s;
        var Array a;
        if (false) {
            let s = "string constant";
            let s = null;
            let a[1] = a[2];
        }
        else {              // There is no else keyword in the Square files.
            let i = i * (-j);
            let j = j / (-2);   // note: unary negate constant 2
            let i = i | j;
        }
        return;
    }
}
`
    ast, err := parse(strings.NewReader(text))
    if err != nil {
        test.Fatalf("could not parse: %v", err)
    }

    _ = ast
    /* TODO: verify the ast */
}
