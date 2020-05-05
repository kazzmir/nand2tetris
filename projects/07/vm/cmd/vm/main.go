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

type Translator struct {
    gensym uint64
}

func (translator *Translator) Gensym(name string) string {
    use := translator.gensym
    translator.gensym += 1
    return fmt.Sprintf("%v_%v", name, use)
}

type VMCommand interface {
    TranslateToAssembly(*Translator) []string
}

type PushConstant struct {
    VMCommand
    Constant uint64
}

func (constant *PushConstant) TranslateToAssembly(translator *Translator) []string {
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

func (add *Add) TranslateToAssembly(translator *Translator) []string {
    /* a = pop sp
     * b = pop sp
     * out = a + b
     * push out
     */

    return []string{
        "@SP",   // sp=sp-1
        "AM=M-1",
        "D=M",   // d=ram[sp]
        "@SP",
        "AM=M-1", // sp=sp-1
        "M=D+M", // ram[sp]=d+ram[sp]
        "@SP",
        "M=M+1",
    }
}

type Eq struct {
    VMCommand
}

func (eq *Eq) TranslateToAssembly(translator *Translator) []string {
    /* a = pop sp
     * b = pop sp
     * out = a == b
     * push out
     */
    notEqualLabel := translator.Gensym("eq")
    eqDone := translator.Gensym("eq")
    return []string{
        "@SP",
        "AM=M-1",
        "D=M",
        "@SP",
        "AM=M-1",
        "D=D-M",
        fmt.Sprintf("@%v", notEqualLabel),
        "D; JNE",
        "@SP",
        "A=M",
        "M=-1",
        fmt.Sprintf("@%v", eqDone),
        "0; JMP",
        fmt.Sprintf("(%v)", notEqualLabel),
        "@SP",
        "A=M",
        "M=0",
        fmt.Sprintf("(%v)", eqDone),
        "@SP",
        "M=M+1",
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
        case "eq":
            return &Eq{}, nil
        case "add":
            return &Add{}, nil
    }

    return nil, fmt.Errorf("unknown command %v", useParts[0])
}

func processVMLine(line string) (VMCommand, error) {
    processed := normalizeWhitespace(line)
    if len(processed) == 0 {
        return nil, nil
    }

    // fmt.Printf("Processing line '%v'\n", processed)

    command, err := parseLine(processed)
    if err != nil {
        return nil, err
    }

    /*
    assembly := command.TranslateToAssembly()
    for _, assemblyLine := range assembly {
        fmt.Printf("%v\n", assemblyLine)
    }
    */

    return command, nil
}

func replaceExtension(path string, what string) string {
    parts := strings.Split(path, ".")
    if len(parts) == 2 {
        return fmt.Sprintf("%s.%s", parts[0], what)
    }

    return path
}

func translate(path string) error {
    /* read each line of the file
     * for each line, translate it into the appropriate hack assembly commands
     * output the result to path.asm
     */

    translator := Translator{
        gensym: 0,
    }

    file, err := os.Open(path)
    if err != nil {
        return err
    }
    defer file.Close()

    output, err := os.Create(replaceExtension(path, "asm"))
    if err != nil {
        return err
    }
    defer output.Close()

    scanner := bufio.NewScanner(file)
    var sourceLine uint64
    for scanner.Scan() {
        line := scanner.Text()
        // fmt.Printf("%v: %v\n", i, line)
        sourceLine += 1

        command, err := processVMLine(line)
        if err != nil {
            return fmt.Errorf("Could not process line %v '%v': %v\n", sourceLine, line, err)
        }

        if command == nil {
            continue
        }

        output.WriteString(fmt.Sprintf("// %s\n", line))
        for _, asmLine := range command.TranslateToAssembly(&translator) {
            output.WriteString(asmLine)
            output.Write([]byte{'\n'})
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
