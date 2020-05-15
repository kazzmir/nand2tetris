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
