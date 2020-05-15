package main

import (
    "testing"
    "strings"
    // "fmt"
)

func TestLexer1(test *testing.T) {
    /*
    out, err := standardLexer(strings.NewReader(" this  while  "))
    fmt.Printf("out %+v error %v\n", out, err)
    */
}

func TestLexerThis(test *testing.T) {
    /* with only the 'this' lexeme this should produce two This tokens */
    doubleThis, err := lexer([]LexerStateMachine{
        buildLiteralMachine("this", TokenThis),
    }, strings.NewReader("thisthis"))

    if err != nil {
        test.Fatalf("parsing 'thisthis' should have not resulted in an error %v", err)
    }

    if len(doubleThis) != 2 {
        test.Fatalf("'thisthis' should have parsed into two tokens but was %v", len(doubleThis))
    }

    if doubleThis[0].Kind != TokenThis && doubleThis[1].Kind != TokenThis {
        test.Fatalf("'thisthis' did not parse into two this token: %v", doubleThis)
    }
}

func TestLexerIdentifier(test *testing.T){
    tokens, err := lexer([]LexerStateMachine{makeIdentifierMachine()}, strings.NewReader("anidentifier"))
    if err != nil {
        test.Fatalf("error was not nil: %v", err)
    }

    if len(tokens) != 1 {
        test.Fatalf("did not parse exactly one token %v", tokens)
    }

    if tokens[0].Kind != TokenIdentifier {
        test.Fatalf("did not parse an identifier %v", tokens)
    }
}

func TestLexerIdentifierThis(test *testing.T){
    machines := []LexerStateMachine{
        makeIdentifierMachine(),
        makeThisMachine(),
    }
    /* an identifier should parse because its a longer match then 'this' */
    tokens, err := lexer(machines, strings.NewReader("thisthis"))
    if err != nil {
        test.Fatalf("error was not nil: %v", err)
    }

    if len(tokens) != 1 {
        test.Fatalf("did not parse exactly one token %v", tokens)
    }

    if tokens[0].Kind != TokenIdentifier {
        test.Fatalf("did not parse an identifier %v", tokens)
    }

    if tokens[0].Value != "thisthis" {
        test.Fatalf("identifier value was not 'thisthis': %v", tokens[0].Value)
    }
}

func TestLexerIdentifierThis2(test *testing.T){
    machines := []LexerStateMachine{
        makeIdentifierMachine(),
        makeThisMachine(),
    }
    /* 'this' should match because it is higher precedence than identifier */
    tokens, err := lexer(machines, strings.NewReader("this"))
    if err != nil {
        test.Fatalf("error was not nil: %v", err)
    }

    if len(tokens) != 1 {
        test.Fatalf("did not parse exactly one token %v", tokens)
    }

    if tokens[0].Kind != TokenThis {
        test.Fatalf("did not parse a this token %v", tokens)
    }
}

func TestLexerIdentifierNumber(test *testing.T){
    machines := []LexerStateMachine{
        makeIdentifierMachine(),
        makeNumberMachine(),
    }
    tokens, err := lexer(machines, strings.NewReader("12"))
    if err != nil {
        test.Fatalf("error was not nil: %v", err)
    }

    if len(tokens) != 1 {
        test.Fatalf("did not parse exactly one token %v", tokens)
    }

    if tokens[0].Kind != TokenNumber {
        test.Fatalf("did not parse a number token %v", tokens)
    }
}

func TestPlus(test *testing.T){
    tokens, err := lexer([]LexerStateMachine{makePlusMachine()}, strings.NewReader("+"))
    if err != nil {
        test.Fatalf("error was not nil: %v", err)
    }

    if len(tokens) != 1 {
        test.Fatalf("did not parse exactly one token: %v", tokens)
    }

    if tokens[0].Kind != TokenPlus {
        test.Fatalf("did not parse +: %v", tokens)
    }
}

func TestPlusFull(test *testing.T){
    tokens, err := standardLexer(strings.NewReader("+"))
    if err != nil {
        test.Fatalf("error was not nil: %v", err)
    }

    if len(tokens) != 1 {
        test.Fatalf("did not parse exactly one token: %v", tokens)
    }

    if tokens[0].Kind != TokenPlus {
        test.Fatalf("did not parse +: %v", tokens)
    }
}

func TestLexerMath(test *testing.T){
    tokens, err := standardLexer(strings.NewReader("1 + 2"))
    if err != nil {
        test.Fatalf("error was not nil: %v", err)
    }

    tokens = removeWhitespaceTokens(tokens)

    if len(tokens) != 3 {
        test.Fatalf("did not parse three tokens: %v", tokens)
    }

    if tokens[0].Kind != TokenNumber {
        test.Fatalf("did not parse token[0] as number: %v", tokens[0])
    }

    if tokens[1].Kind != TokenPlus {
        test.Fatalf("did not parse token[1] as +: %v", tokens[1])
    }

    if tokens[2].Kind != TokenNumber {
        test.Fatalf("did not parse token[2] as number: %v", tokens[2])
    }
}
