package main

import (
    "os"
    "fmt"
    "bufio"
    "time"
    // "strings"

    _ "runtime/pprof"
    _ "runtime"
)

func lex(path string) error {
    file, err := os.Open(path)
    if err != nil {
        return err
    }
    defer file.Close()

    tokens := make(chan Token, 1000)

    start := time.Now()

    go func(){
        err = standardLexer(file, tokens)
    }()

    var count uint64
    for token := range tokens {
        _ = token
        count += 1
        // fmt.Printf("%+v\n", token)
    }

    if err != nil {
        return err
    }

    end := time.Now()

    // tokens = removeWhitespaceTokens(tokens)
    fmt.Printf("Lexed %v tokens in %v\n", count, end.Sub(start))
    return nil
}

func compile(path string) error {
    file, err := os.Open(path)
    if err != nil {
        return err
    }
    defer file.Close()

    start := time.Now()

    ast, err := parse(file)
    if err != nil {
        return err
    }

    end := time.Now()

    fmt.Printf("Parsed %v in %v\n", path, end.Sub(start))

    /*
    fmt.Printf("%v\n", ast.ToSExpression())
    */

    output, err := os.Create("out.vm")
    if err != nil {
        return err
    }
    defer output.Close()

    buffer := bufio.NewWriter(output)
    defer buffer.Flush()

    start = time.Now()

    err = GenerateCode(ast, buffer)
    if err != nil {
        return err
    }

    end = time.Now()
    fmt.Printf("Codegen %v in %v\n", path, end.Sub(start))

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
    // cpu, _ := os.Create("cpu.prof")
    // pprof.StartCPUProfile(cpu)
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

    /*
    memory, _ := os.Create("memory.prof")
    runtime.GC()
    pprof.WriteHeapProfile(memory)
    memory.Close()

    pprof.StopCPUProfile()
    cpu.Close()
    */
}
