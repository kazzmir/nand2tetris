package main

import (
    "os"
    "fmt"
    // "strings"

    "runtime/pprof"
)

func compile(path string) error {
    file, err := os.Open(path)
    if err != nil {
        return err
    }
    defer file.Close()

    tokens := make(chan Token, 1000)

    go func(){
        err = standardLexer(file, tokens)
    }()

    var count uint64
    for token := range tokens {
        _ = token
        count += 1
    }

    if err != nil {
        return err
    }

    // tokens = removeWhitespaceTokens(tokens)
    fmt.Printf("Lexed %v tokens\n", count)
    return nil
}

func compileAll(paths []string) error {
    for _, path := range paths {
        err := compile(path)
        if err != nil {
            return err
        }
    }

    return nil
}

func test(){
    /*
    tokens, err := standardLexer(strings.NewReader("1 + 2"))
    fmt.Printf("Tokens %v error %v\n", tokens, err)
    */
}

func main(){
    cpu, _ := os.Create("cpu.prof")
    pprof.StartCPUProfile(cpu)
    // test()

    // TestL()
    if len(os.Args) == 1 {
        fmt.Printf("Give a directory of .jack files or a list of .jack files")
        return
    } else {
        err := compileAll(os.Args[1:])
        if err != nil {
            fmt.Printf("Error: %v\n", err)
        }
    }

    pprof.StopCPUProfile()
    cpu.Close()
}
