---
title: 'CTF Challenge: RISC It for the Biscuit'
---

I have developed a RISC-V 64bit (RV64IM) virtual machine you can download [here](/assets/files/RISC%20It%20for%20the%20Biscuit.7z). I decided to keep it simple, and didn't add an MMU, so it uses the host process memory addresses. I doubt thats going to cause any issues. There is a `flag.txt` the in same directory as the VM that you need to read by providing the VM with a binary to execute.

You can compile C code for the virtual machine like so:
```c
#include <stdint.h>

extern void syscall_exit(void);
extern void print_char_string(char* str);
extern void print_u64_hex(uint64_t val);
extern void print_u64(uint64_t val);
extern void print_new_line(void);

// Return the n-th prime (n >= 1). For n == 1, returns 2.
uint64_t find_nth_prime(uint64_t n) {
    if (n == 0) {
        return 0;  // or handle as error
    }

    uint64_t t0 = 1;      // current number being tested
    uint64_t t3 = 0;      // count of primes found

    while (1) {
        t0 += 1;          // next candidate
        int t2 = 0;       // 0 = assume prime, 1 = not prime

        // special case: 2 is prime
        if (t0 == 2) {
            t3 += 1;
            if (t3 == n) {
                return t0;
            }
            continue;
        }

        // check divisors from 2 up to t4*t4 > t0
        uint64_t t4 = 2;
        while (1) {
            uint64_t t1 = t4 * t4;
            if (t1 > t0) {
                break;          // no divisor found
            }
            if (t0 % t4 == 0) { // found a divisor
                t2 = 1;
                break;
            }
            t4 += 1;
        }

        if (t2 == 0) {     // still marked as prime
            t3 += 1;
            if (t3 == n) {
                return t0; // t0 is the n-th prime
            }
        }
    }
}

void run(void) {
    uint64_t nth_prime_number_to_find = 10000;
    uint64_t nth_prime = find_nth_prime(nth_prime_number_to_find);
    print_char_string("The "); 
    print_u64(nth_prime_number_to_find);
    print_char_string("th prime number is: ");
    print_u64(nth_prime);
    print_new_line();
    syscall_exit();
}
```

This uses a couple syscall's the VM exposes, those are defined like so:
```asm
.section .text
.globl _start, syscall_exit, print_char_string, print_u64_hex, print_u64, print_new_line

_start:
    call run

syscall_exit:
    li a7, 0x5d
    ecall

print_char_string:
    li a7, 100
    ecall
    ret

print_u64_hex:
    li a7, 101
    ecall
    ret

print_u64:
    li a7, 102
    ecall
    ret

print_new_line:
    li a7, 103
    ecall
    ret
```

Lastly, you need a linker script like:
```ld
ENTRY(_start)

SECTIONS {
    . = 0x80000000;

    .text : {
        *(.text*)
        *(.rodata*)
    }

    .data : {
        *(.data*)
    }

    .bss : {
        *(.bss*)
        *(COMMON)
    }
}
```

To build the example, use Developer Command Prompt for VS or Developer PowerShell for VS. 
In the Visual Studio Installer, you will need C++ Clang tools for Windows.
```bash
clang -target riscv64 -march=rv64im -mcmodel=medany -nostdlib -Wl,-T,link.ld stub.s prime.c -o prime.o
llvm-objcopy -O binary prime.o prime.bin
# Look at the disassembly
llvm-objdump --disassemble prime.o
```

You should see:
```text
PS Release> .\RV64VMv4.exe .\prime.bin
The 10000th prime number is: 104729
```

When running the VM, the exit code indicates the following:
```text
SUCCESS(0): It exited successful using the exit syscall.
INVALID_INSTRUCTION(1): You give the VM bad machine code.
INVALID_SYSCALL(2): You called something other than the 5 syscalls in the example.
```

## Submitting a Solution 
You can give the binary to acebond on Discord, and I'll run it on the latest version of Windows 11 25H2 and give you the output, and add you to the list of solvers if it prints the flag.

## Smart People Who Have Solved It
:(
