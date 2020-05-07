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

/* the temp segment starts at ram 5 */
const TempStart = 5
/* the pointer segment starts at ram 3 */
const PointerStart = 3

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

type PopLocal struct {
    VMCommand
    Index int
}

func popToSegment(segment string, index int) []string {
    /* ram[local+index] = sp--
     *
     * store ram[sp-1] in r13
     * compute local+index, store in r14
     * store r14 into ram[r13]
     */
    return []string{
        "@SP",
        "AM=M-1",
        "D=M", // d = ram[sp]

        "@R13",
        "M=D", // ram[r13] = d

        fmt.Sprintf("@%v", index),
        "D=A",
        fmt.Sprintf("@%v", segment),
        "D=D+M", // ram[local+index]
        "@R14",
        "M=D",  // ram[r14] = local+index

        "@R13", // a = r13
        "D=M",

        "@R14",
        "A=M",
        "M=D",
    }
}

func pushToSegment(segment string, index int) []string {
    return []string {
        fmt.Sprintf("@%v", segment),
        "D=M",
        fmt.Sprintf("@%v", index),
        "D=D+A", // d = segment+index
        "A=D",
        "D=M",  // d = ram[segment+index]

        "@SP",
        "A=M",
        "M=D", // ram[sp] = d
        "@SP",
        "M=M+1", // sp++
    }
}

func (local *PopLocal) TranslateToAssembly(translator *Translator) []string {
    return popToSegment("LCL", local.Index)
}

type PopArgument struct {
    VMCommand
    Index int
}

func (argument *PopArgument) TranslateToAssembly(translator *Translator) []string {
    return popToSegment("ARG", argument.Index)
}

type PopThis struct {
    VMCommand
    Index int
}

func (this *PopThis) TranslateToAssembly(translator *Translator) []string {
    return popToSegment("THIS", this.Index)
}

type PopThat struct {
    VMCommand
    Index int
}

func (that *PopThat) TranslateToAssembly(translator *Translator) []string {
    return popToSegment("THAT", that.Index)
}

type PopTemp struct {
    VMCommand
    Index int
}

func (temp *PopTemp) TranslateToAssembly(translator *Translator) []string {
    index := TempStart + temp.Index
    return []string{
        "@SP",
        "AM=M-1",
        "D=M",
        fmt.Sprintf("@%v", index),
        "M=D",
    }
}

type PopPointer struct {
    VMCommand
    Index int
}

func (pointer *PopPointer) TranslateToAssembly(translator *Translator) []string {
    index := PointerStart + pointer.Index
    return []string{
        "@SP",
        "AM=M-1",
        "D=M",
        fmt.Sprintf("@%v", index),
        "M=D",
    }
}

type PushLocal struct {
    VMCommand
    Index int
}

func (local *PushLocal) TranslateToAssembly(translator *Translator) []string {
    return pushToSegment("LCL", local.Index)
}

type PushTemp struct {
    VMCommand
    Index int
}

func (temp *PushTemp) TranslateToAssembly(translator *Translator) []string {
    index := TempStart + temp.Index
    return []string{
        fmt.Sprintf("@%v", index),
        "D=M",
        "@SP",
        "A=M",
        "M=D",
        "@SP",
        "M=M+1",
    }
}

type PushThis struct {
    VMCommand
    Index int
}

func (this *PushThis) TranslateToAssembly(translator *Translator) []string {
    return pushToSegment("THIS", this.Index)
}

type PushThat struct {
    VMCommand
    Index int
}

func (that *PushThat) TranslateToAssembly(translator *Translator) []string {
    return pushToSegment("THAT", that.Index)
}

type PushArgument struct {
    VMCommand
    Index int
}

func (argument *PushArgument) TranslateToAssembly(translator *Translator) []string {
    return pushToSegment("ARG", argument.Index)
}

type PushPointer struct {
    VMCommand
    Index int
}

func (pointer *PushPointer) TranslateToAssembly(translator *Translator) []string {
    index := PointerStart + pointer.Index
    return []string{
        fmt.Sprintf("@%v", index),
        "D=M",
        "@SP",
        "A=M",
        "M=D",
        "@SP",
        "M=M+1",
    }
}

type PushStatic struct {
    VMCommand
    Index int
}

func (static *PushStatic) TranslateToAssembly(translator *Translator) []string {
    return []string{
        fmt.Sprintf("@static.%v", static.Index),
        "D=M",
        "@SP",
        "A=M",
        "M=D",
        "@SP",
        "M=M+1",
    }
}

type PopStatic struct {
    VMCommand
    Index int
}

func (static *PopStatic) TranslateToAssembly(translator *Translator) []string {
    return []string{
        "@SP",
        "AM=M-1",
        "D=M",
        fmt.Sprintf("@static.%v", static.Index),
        "M=D",
    }
}

func getPushPopParts(parts []string) (string, int, error) {
    if len(parts) == 3 {
        where := parts[1]
        number := parts[2]

        value, err := strconv.ParseInt(number, 10, 64)
        if err != nil {
            return "", 0, fmt.Errorf("push/pop value must be an integer: %v", err)
        }

        return where, int(value), nil
    } else {
        return "", 0, fmt.Errorf("push/pop needs 3 parts, but only given %v: %v", len(parts), parts)
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
            where, index, err := getPushPopParts(useParts)
            if err != nil {
                return nil, err
            }

            switch where {
                case "constant": return &PushConstant{Constant: uint64(index)}, nil
                case "local": return &PushLocal{Index: index}, nil
                case "that": return &PushThat{Index: index}, nil
                case "this": return &PushThis{Index: index}, nil
                case "argument": return &PushArgument{Index: index}, nil
                case "temp": return &PushTemp{Index: index}, nil
                case "pointer": return &PushPointer{Index: index}, nil
                case "static": return &PushStatic{Index: index}, nil
            }

            return nil, fmt.Errorf("Unknown push command '%v'", where)
        case "pop":
            where, index, err := getPushPopParts(useParts)
            if err != nil {
                return nil, err
            }
            switch where {
                case "local": return &PopLocal{Index: index}, nil
                case "argument": return &PopArgument{Index: index}, nil
                case "this": return &PopThis{Index: index}, nil
                case "that": return &PopThat{Index: index}, nil
                case "temp": return &PopTemp{Index: index}, nil
                case "pointer": return &PopPointer{Index: index}, nil
                case "static": return &PopStatic{Index: index}, nil
            }
            return nil, fmt.Errorf("Unknown memory area '%v'", where)
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
