package main

import (
    "os"
    "io"
    "fmt"
    "bufio"
    "path/filepath"
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
    CurrentFile string
    CurrentFunction string
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
}

func (gt *Gt) TranslateToAssembly(translator *Translator) []string {
    return generateComparison(translator, "JLE")
}

type Neg struct {
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
    Index int
}

func (argument *PopArgument) TranslateToAssembly(translator *Translator) []string {
    return popToSegment("ARG", argument.Index)
}

type PopThis struct {
    Index int
}

func (this *PopThis) TranslateToAssembly(translator *Translator) []string {
    return popToSegment("THIS", this.Index)
}

type PopThat struct {
    Index int
}

func (that *PopThat) TranslateToAssembly(translator *Translator) []string {
    return popToSegment("THAT", that.Index)
}

type PopTemp struct {
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
    Index int
}

func (local *PushLocal) TranslateToAssembly(translator *Translator) []string {
    return pushToSegment("LCL", local.Index)
}

type PushTemp struct {
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
    Index int
}

func (this *PushThis) TranslateToAssembly(translator *Translator) []string {
    return pushToSegment("THIS", this.Index)
}

type PushThat struct {
    Index int
}

func (that *PushThat) TranslateToAssembly(translator *Translator) []string {
    return pushToSegment("THAT", that.Index)
}

type PushArgument struct {
    Index int
}

func (argument *PushArgument) TranslateToAssembly(translator *Translator) []string {
    return pushToSegment("ARG", argument.Index)
}

type PushPointer struct {
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
    Index int
}

func (static *PushStatic) TranslateToAssembly(translator *Translator) []string {
    return []string{
        fmt.Sprintf("@static.%v.%v", translator.CurrentFile, static.Index),
        "D=M",
        "@SP",
        "A=M",
        "M=D",
        "@SP",
        "M=M+1",
    }
}

type PopStatic struct {
    Index int
}

func (static *PopStatic) TranslateToAssembly(translator *Translator) []string {
    return []string{
        "@SP",
        "AM=M-1",
        "D=M",
        fmt.Sprintf("@static.%v.%v", translator.CurrentFile, static.Index),
        "M=D",
    }
}

type Label struct {
    Name string
}

func (label *Label) TranslateToAssembly(translator *Translator) []string {
    return []string {
        fmt.Sprintf("(%v)", label.Name),
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

type IfGoto struct {
    Name string
}

func (ifgoto *IfGoto) TranslateToAssembly(translator *Translator) []string {
    /* if-goto X
     * pop a; if a != 0: jump X
     */
    return []string {
        "AM=M-1",
        "D=M",
        fmt.Sprintf("@%v", ifgoto.Name),
        "D; JNE",
    }
}

type Goto struct {
    Name string
}

func (this *Goto) TranslateToAssembly(translator *Translator) []string {
    return []string {
        fmt.Sprintf("@%v", this.Name),
        "0; JMP",
    }
}

type Function struct {
    Name string
    Locals int
}

func (function *Function) TranslateToAssembly(translator *Translator) []string {
    /* modifies the translator */
    translator.CurrentFunction = function.Name

    out := []string {
        fmt.Sprintf("(%v)", function.Name),
    }

    for i := 0; i < function.Locals; i++ {

        local := []string {
            "@SP",
            "A=M",
            "M=0",
            "@SP",
            "M=M+1",
        }

        out = append(out, local...)
    }

    return out
}

type Return struct {
}

func (ret *Return) TranslateToAssembly(translator *Translator) []string {
    return []string {
        /* frame = lcl, ret = *(frame-5) */
        "@LCL",
        "D=M", // d = LCL
        "@R13",
        "M=D", // save LCL in r13
        "@5",
        "A=D-A", // 5 = (return address, that, this, arg, lcl)
        "D=M", // d=*(lcl-5), which is the return address
        "@R14",
        "M=D", // r14 = return address

        /* *ARG = pop() */
        "@SP",
        "AM=M-1",
        "D=M", // d = popped value
        "@ARG",
        "A=M",
        "M=D", // *arg = d

        "@ARG",
        "D=M+1",
        "@SP",
        "M=D", // set sp to arg+1

        /* that = *(frame-1) */
        "@1",
        "D=A",
        "@R13",
        "A=M-D",
        "D=M",
        "@THAT",
        "M=D",

        /* this = *(frame-2) */
        "@2",
        "D=A",
        "@R13",
        "A=M-D",
        "D=M",
        "@THIS",
        "M=D",

        /* arg = *(frame-3) */
        "@3",
        "D=A",
        "@R13",
        "A=M-D",
        "D=M",
        "@ARG",
        "M=D",

        /* lcl = *(frame-4) */
        "@4",
        "D=A",
        "@R13",
        "A=M-D",
        "D=M",
        "@LCL",
        "M=D",

        /* goto ret */
        "@R14",
        "A=M",
        "0; JMP",
    }
}

type Call struct {
    Name string
    Arguments int
}

func (call *Call) TranslateToAssembly(translator *Translator) []string {
    returnAddress := translator.Gensym(fmt.Sprintf("%v_return", translator.CurrentFunction))

    return []string {
        /* push return address */
        fmt.Sprintf("@%v", returnAddress),
        "D=A",
        "@SP",
        "A=M",
        "M=D",
        "@SP",
        "M=M+1",

        /* push lcl */
        "@LCL",
        "D=M",
        "@SP",
        "A=M",
        "M=D",
        "@SP",
        "M=M+1",

        /* push arg */
        "@ARG",
        "D=M",
        "@SP",
        "A=M",
        "M=D",
        "@SP",
        "M=M+1",

        /* push this */
        "@THIS",
        "D=M",
        "@SP",
        "A=M",
        "M=D",
        "@SP",
        "M=M+1",

        /* push that */
        "@THAT",
        "D=M",
        "@SP",
        "A=M",
        "M=D",
        "@SP",
        "M=M+1",

        /* arg = sp-n-5 */
        "@SP",
        "D=M",
        fmt.Sprintf("@%v", call.Arguments),
        "D=D-A",
        "@5",
        "D=D-A",
        "@ARG",
        "M=D",

        /* lcl = sp */
        "@SP",
        "D=M",
        "@LCL",
        "M=D",

        /* goto f */
        fmt.Sprintf("@%v", call.Name),
        "0; JMP",

        fmt.Sprintf("(%v)", returnAddress),
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
        case "function":
            if len(useParts) == 3 {
                locals, err := strconv.Atoi(useParts[2])
                if err != nil {
                    return nil, fmt.Errorf("Expected a number for the locals '%v': %v", useParts[2], err)
                }

                return &Function{
                    Name: useParts[1],
                    Locals: locals,
                }, nil
            } else {
                return nil, fmt.Errorf("Expected a name and number of locals for function")
            }
        case "return":
            return &Return{}, nil
        case "call":
            if len(useParts) == 3 {
                name := useParts[1]
                arguments, err := strconv.Atoi(useParts[2])
                if err != nil {
                    return nil, fmt.Errorf("Expected a number of arguments for call '%v': %v", useParts[2], err)
                }

                return &Call{Name: name, Arguments: arguments}, nil
            } else {
                return nil, fmt.Errorf("Call needs a function name and number of arguments")
            }
        case "label":
                if len(useParts) == 2 {
                    return &Label{Name: useParts[1]}, nil
                } else {
                    return nil, fmt.Errorf("Missing label name")
                }
        case "if-goto":
            if len(useParts) == 2 {
                return &IfGoto{Name: useParts[1]}, nil
            } else {
                return nil, fmt.Errorf("Missing label name")
            }
        case "goto":
            if len(useParts) == 2 {
                return &Goto{Name: useParts[1]}, nil
            } else {
                return nil, fmt.Errorf("Missing label name")
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

    return nil, fmt.Errorf("unknown command '%v'", useParts[0])
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

    return fmt.Sprintf("%v.asm", path)
}

func bootstrapCode() []string {
    return []string {
        "call Sys.init 0",
    }
}

func removeExtension(path string) string {
    dot := strings.Index(path, ".")
    if dot != -1 {
        return path[0:dot]
    }
    return path
}

func className(path string) string {
    return removeExtension(filepath.Base(path))
}

func translateVMFile(output io.Writer, path string, translator *Translator) error {
    file, err := os.Open(path)
    if err != nil {
        return err
    }
    defer file.Close()

    translator.CurrentFile = className(path)

    io.WriteString(output, fmt.Sprintf("// %v", path))
    output.Write([]byte{'\n'})

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

        io.WriteString(output, fmt.Sprintf("// %s\n", line))
        for _, asmLine := range command.TranslateToAssembly(translator) {
            io.WriteString(output, asmLine)
            output.Write([]byte{'\n'})
        }
    }

    err = scanner.Err()
    if err != nil {
        return err
    }

    return nil
}

func writeBootstrapCode(output io.Writer, translator *Translator) error {
    translator.CurrentFunction = "Sys.init"
    /* initialize SP to 256 */
    io.WriteString(output, "@256\n")
    io.WriteString(output, "D=A\n")
    io.WriteString(output, "@SP\n")
    io.WriteString(output, "M=D\n")

    /*
    io.WriteString(output, "@Sys.init\n")
    io.WriteString(output, "0; JMP\n")
    */

    for _, line := range bootstrapCode() {
        command, err := processVMLine(line)
        if err != nil {
            return fmt.Errorf("Error in bootstrap code '%v': %v", line, err)
        }

        if command == nil {
            return fmt.Errorf("Did not produce a command for bootstrap code line '%v'", line)
        }

        for _, asmLine := range command.TranslateToAssembly(translator) {
            io.WriteString(output, asmLine)
            output.Write([]byte{'\n'})
        }
    }

    return nil
}

func isFile(path string) bool {
    stat, err := os.Stat(path)
    if err != nil {
        return false
    }

    return !stat.IsDir()
}

func isDir(path string) bool {
    stat, err := os.Stat(path)
    if err != nil {
        return false
    }

    return stat.IsDir()
}

func findVMFiles(root string) ([]string, error) {
    var out []string

    err := filepath.Walk(root, func (path string, info os.FileInfo, err error) error {
        if strings.HasSuffix(path, ".vm") {
            out = append(out, path)
        }

        return nil
    })

    return out, err
}

func translate(path string) error {
    /* read each line of the file
     * for each line, translate it into the appropriate hack assembly commands
     * output the result to path.asm
     */

    translator := Translator{
        gensym: 0,
    }

    var vmFiles []string

    resolved, err := filepath.EvalSymlinks(path)
    if err != nil {
        return err
    }

    if isFile(resolved) {
        vmFiles = []string{resolved}
    } else if isDir(resolved) {
        vmFiles, err = findVMFiles(resolved)
        if err != nil {
            return err
        }
    } else {
        return fmt.Errorf("Not a file or directory?")
    }

    fmt.Printf("Translating files %v\n", vmFiles)

    output, err := os.Create(replaceExtension(path, "asm"))
    if err != nil {
        return err
    }
    defer output.Close()

    err = writeBootstrapCode(output, &translator)
    if err != nil {
        return err
    }

    for _, vmFile := range vmFiles {
        err = translateVMFile(output, vmFile, &translator)
        if err != nil {
            return err
        }
    }

    return nil
}

func main(){
    if len(os.Args) < 2 {
        fmt.Printf("Give a .vm file or directory with .vm files in it\n")
        return
    }

    err := translate(os.Args[1])
    if err != nil {
        fmt.Printf("Could not translate %v: %v\n", os.Args[1], err)
    } else {
        fmt.Printf("Translated %v\n", os.Args[1])
    }
}
