// This file is part of www.nand2tetris.org
// and the book "The Elements of Computing Systems"
// by Nisan and Schocken, MIT Press.
// File name: projects/04/Fill.asm

// Runs an infinite loop that listens to the keyboard input.
// When a key is pressed (any key), the program blackens the screen,
// i.e. writes "black" in every pixel;
// the screen should remain fully black as long as the key is pressed. 
// When no key is pressed, the program clears the screen, i.e. writes
// "white" in every pixel;
// the screen should remain fully clear as long as no key is pressed.

// Put your code here.

// pressed=kbd != 0
// start:
// while {
//   is_pressed = kbd != 0
//   if is_pressed != pressed {
//     break
//   }
// }
//
// if is_pressed > 0 {
//   fill=1
// } else {
//   fill=0
// }
// 
// for pixel = 0; pixel < 8192; pixel++ {
//   screen[pixel] = fill
// }
// goto start

@KBD
D=M
@pressed
M=D

(CHECK_KEY)
@KBD
D=M
@pressed
D=D-M
@CHECK_KEY
D; JEQ

@KBD
D=M
@pressed
M=D

// if keyboard>0 {
//   fill = 1
// } else {
//   fill = 0
// }
@KBD
D=M
@NO_PRESS
D; JEQ

@fill
M=0
@fill
M=M-1

@BLIT
0; JEQ
(NO_PRESS)
@fill
M=0

(BLIT)
// index=0
@8191
// @3
D=A
@index
M=D

(BLIT_LOOP)
@index
D=M
@BLIT_END
D; JLT

@index
D=M
@SCREEN
D=D+A
@screen_pointer
M=D
@fill
D=M
@screen_pointer
A=M
M=D

@index
M=M-1
@BLIT_LOOP
0; JEQ

(BLIT_END)

// go back to start of program
@CHECK_KEY
0; JEQ
