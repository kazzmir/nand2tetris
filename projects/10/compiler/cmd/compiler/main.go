package main

import (
    _ "os"
    "fmt"
    "strings"
)

func compile(path string) error {
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
    tokens, err := standardLexer(strings.NewReader("1 + 2"))
    fmt.Printf("Tokens %v error %v\n", tokens, err)
}

func main(){
    test()

    // TestL()
    /*
    if len(os.Args) == 1 {
        fmt.Printf("Give a directory of .jack files or a list of .jack files")
        return
    } else {
        err := compileAll(os.Args[1:])
        if err != nil {
            fmt.Printf("Error: %v\n", err)
        }
    }
    */
}
