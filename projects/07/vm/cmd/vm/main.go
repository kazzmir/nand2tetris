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

type Sub struct {
    VMCommand
}

func (sub *Sub) TranslateToAssembly(translator *Translator) []string {
    /* sp -> y
     *    -> x
     * out = x-y
     * push out
     */
    return []string {
        "@SP",
        "AM=M-1",
        "D=M",    // y
        "@SP",
        "AM=M-1",
        "M=M-D",
        "@SP",
        "M=M+1",
    }
}

type Lt struct {
    VMCommand
}

func generateComparison(translator *Translator, jumpFalse string) []string {
    /* a = pop sp
     * b = pop sp
     * out = b CMP a
     * push out
     */

    falseBranch := translator.Gensym("cmp_false")
    done := translator.Gensym("cmp_done")

    return []string{
        "@SP",
        "AM=M-1",
        "D=M",
        "@SP",
        "AM=M-1",
        "D=M-D", // b-a
        // d<0, then jump to m=-1 (true)
        // d>=0, then jump to m=0 (false)
        fmt.Sprintf("@%v", falseBranch),
        fmt.Sprintf("D; %v", jumpFalse),
        "@SP",
        "A=M",
        "M=-1",
        fmt.Sprintf("@%v", done),
        "0; JMP",
        fmt.Sprintf("(%v)", falseBranch),
        "@SP",
        "A=M",
        "M=0",
        fmt.Sprintf("(%v)", done),
        "@SP",
        "M=M+1",
    }

}

func (lt *Lt) TranslateToAssembly(translator *Translator) []string {
    /* sp -> b
     *    -> a
     * a-b is true if a<b and false if a>=b
     *
     */
    return generateComparison(translator, "JGE")
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

    return generateComparison(translator, "JNE")
}

type Gt struct {
    VMCommand
}

func (gt *Gt) TranslateToAssembly(translator *Translator) []string {
    return generateComparison(translator, "JLE")
}

type Neg struct {
    VMCommand
}

func (neg *Neg) TranslateToAssembly(translator *Translator) []string {
    return []string {
        "@SP",
        "AM=M-1",
        "M=-M",
        "@SP",
        "M=M+1",
    }
}

type Not struct {
    VMCommand
}

func (not *Not) TranslateToAssembly(translator *Translator) []string {
    return []string {
        "@SP",
        "AM=M-1",
        "M=!M",
        "@SP",
        "M=M+1",
    }
}

type And struct {
    VMCommand
}

func (and *And) TranslateToAssembly(translator *Translator) []string {
    return []string {
        "@SP",
        "AM=M-1",
        "D=M",
        "@SP",
        "AM=M-1",
        "M=D&M",
        "@SP",
        "M=M+1",
    }
}

type Or struct {
    VMCommand
}

func (or *Or) TranslateToAssembly(translator *Translator) []string {
    return []string {
        "@SP",
        "AM=M-1",
        "D=M",
        "@SP",
        "AM=M-1",
        "M=D|M",
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
        case "lt":
            return &Lt{}, nil
        case "gt":
            return &Gt{}, nil
        case "eq":
            return &Eq{}, nil
        case "add":
            return &Add{}, nil
        case "sub":
            return &Sub{}, nil
        case "neg":
            return &Neg{}, nil
        case "and":
            return &And{}, nil
        case "or":
            return &Or{}, nil
        case "not":
            return &Not{}, nil
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
            return fmt.Errorf("Could not process line %v '%v': %v", sourceLine, line, err)
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
