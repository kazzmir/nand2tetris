package main

import (
    "testing"
    "strings"
)

func TestClassParse(test *testing.T){
    text := "class Foo { }"
    tokens := make(chan Token)

    var err error
    go func(){
        err = standardLexer(strings.NewReader(text), tokens)
    }()

    if err != nil {
        test.Fatalf("could not lex %v: %v", text, err)
    }

    class, err := parse(tokens)
    if err != nil {
        test.Fatalf("could not parse %v: %v", text, err)
    }

    if class.Kind() != ASTKindClass {
        test.Fatalf("did not get a class ast: %v", class)
    }
}
