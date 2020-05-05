package main

import (
    "os"
    "fmt"
    "bufio"
    "strings"
    "strconv"
)

func normalizeWhitespace(line string) string {
    commentStart := strings.Index(line, "//")
    if commentStart != -1 {
        line = line[0:commentStart]
    }

    return strings.TrimSpace(line)
}

type VMCommand interface {
    TranslateToAssembly() []string
}

type PushConstant struct {
    VMCommand
    Constant uint64
}

func (constant *PushConstant) TranslateToAssembly() []string {
    return []string{
        fmt.Sprintf("@%v", constant.Constant), // a = constant
        "D=A", // d = a
        "@SP", // a=0
        "A=M", // a = ram[0]
        "M=D", // ram[a] = D
        "@SP", // a = 0
        "M=M+1", // ram[a] = ram[a] + 1
    }
}

type Add struct {
    VMCommand
}

func (add *Add) TranslateToAssembly() []string {
    /* a = pop sp
     * b = pop sp
     * out = a + b
     * push out
     */

    /* FIXME: this needs to decrement sp first, then assign
     */
    return []string{
        "@SP",   // sp=sp-1
        "AM=M-1",
        "D=M",   // d=ram[sp]
        "@SP",
        "AM=M-1", // sp=sp-1
        "M=D+M", // ram[sp]=d+ram[sp]
    }
}

func parseLine(line string) (VMCommand, error) {
    parts := strings.Split(line, " ")
    var useParts []string
    for _, part := range parts {
        if len(part) > 0 {
            useParts = append(useParts, part)
        }
    }

    if len(useParts) == 0 {
        return nil, fmt.Errorf("no command given")
    }

    switch strings.ToLower(useParts[0]) {
        case "push":
            if len(useParts) > 1 {
                switch strings.ToLower(useParts[1]) {
                    case "constant":
                        if len(useParts) > 2 {
                            value, err := strconv.ParseInt(useParts[2], 10, 64)
                            if err != nil {
                                return nil, fmt.Errorf("push constant value must be an integer: %v", err)
                            }

                            return &PushConstant{
                                Constant: uint64(value),
                            }, nil
                        } else {
                            return nil, fmt.Errorf("push constant needs a value")
                        }
                }
            } else {
                return nil, fmt.Errorf("push command needs two arguments")
            }
        case "add":
            return &Add{}, nil
    }

    return nil, fmt.Errorf("unknown command %v", useParts[0])
}

func processVMLine(line string) error {
    processed := normalizeWhitespace(line)
    if len(processed) == 0 {
        return nil
    }

    // fmt.Printf("Processing line '%v'\n", processed)

    command, err := parseLine(processed)
    if err != nil {
        return err
    }

    assembly := command.TranslateToAssembly()
    for _, assemblyLine := range assembly {
        fmt.Printf("%v\n", assemblyLine)
    }

    return nil
}

func translate(path string) error {
    /* read each line of the file
     * for each line, translate it into the appropriate hack assembly commands
     * output the result to path.asm
     */

    file, err := os.Open(path)
    if err != nil {
        return err
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)
    var sourceLine uint64
    for scanner.Scan() {
        line := scanner.Text()
        // fmt.Printf("%v: %v\n", i, line)
        sourceLine += 1

        err = processVMLine(line)
        if err != nil {
            return fmt.Errorf("Could not process line %v '%v': %v\n", sourceLine, line, err)
        }
    }

    err = scanner.Err()
    if err != nil {
        return err
    }

    return nil
}

func main(){
    if len(os.Args) < 2 {
        fmt.Printf("Give a .vm file\n")
        return
    }

    err := translate(os.Args[1])
    if err != nil {
        fmt.Printf("Could not translate %v: %v\n", os.Args[1], err)
    } else {
        fmt.Printf("Translated %v\n", os.Args[1])
    }
}
