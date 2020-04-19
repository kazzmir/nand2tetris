// This file is part of www.nand2tetris.org
// and the book "The Elements of Computing Systems"
// by Nisan and Schocken, MIT Press.
// File name: projects/04/Mult.asm

// Multiplies R0 and R1 and stores the result in R2.
// (R0, R1, R2 refer to RAM[0], RAM[1], and RAM[2], respectively.)

// Put your code here.

// a = ram[0]
// b = ram[1]
// is_negative = b < 0
// out = 0
// while (b > 0){
//   out += a
// }
// 
// if is_negative {
//  out = -out
// }
// 
// 

// r2=0
@R2
M=0

// while r1 > 0
(LOOP)
@R1
D=M
@DONE
D; JEQ

// r2 = r2 + r0
@R0
D=M
@R2
M=M+D

// r1 = r1 - 1
@R1
M=M-1

// jump back to loop
@LOOP
0; JEQ

// halt
(DONE)
@DONE
0; JEQ
