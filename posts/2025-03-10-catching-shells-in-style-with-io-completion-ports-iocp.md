---
title: Catching Shells in Style with I/O Completion Ports (IOCP)
---

I wanted to catch a shell (`windows/x64/shell_reverse_tcp`) on Windows and the versions of netcat available for Windows were a combination of old, deleted by AV, and hosted on untrustworthy websites. Other open-source options also all sucked, the code was generally shit, and in every case created unnecessary threads to handle proxying the input/output between stdin, stdout, and a socket.

So, I decided to write my own, because how hard can it be? My goal was that it had to be single threaded, because that actually requires more skill.

Windows I/O Completion Ports (IOCP) are the answer, you essentially fire read/write operations, and then use `GetQueuedCompletionStatus` to retrieve a completed I/O operation, and normally fire a new one.

There are 2 loops:
- Read from the socket, then write to stdout, repeat
- Read from stdin, then write to the socket, repeat

This essentially translate into a `while(1)` loop, calling `GetQueuedCompletionStatus`, and handling the completed I/O operation appropriately. If a socket read finished, start a write to stdout, if a write to stdout finished, start a read from the socket. If a stdin read finished, start a write to the socket, if a write to the socket finished, start a read from stdin. I should note, the current setup could be both reading and writing to the socket asynchronously, despite the program being single threaded the kernel is not, it will perform these operations asynchronously. It's also safe to read and write to a socket at the same time.

## Full Code
```c
#include <WinSock2.h>
#include <WS2tcpip.h>
#include <Windows.h>

#include <stdio.h>

#pragma comment(lib, "ws2_32.lib")

#define LISTEN_PORT 4444
#define BUFFER_SIZE 4096

#define ArraySize(x) (sizeof x / sizeof x[0])
#define STRINGIFY(x) #x

typedef enum IO_STATE {
	IO_STATE_INVALID,
	IO_STATE_SOCKET_READ,
	IO_STATE_SOCKET_WRITE,
	IO_STATE_FILE_READ,
	IO_STATE_FILE_WRITE,
} IO_STATE;

typedef struct IO_CONTEXT {
	OVERLAPPED overlapped;
	WSABUF     wsaBuf;
	DWORD      wsaRecvFlags;
	IO_STATE   ioState;
} IO_CONTEXT, *PIO_CONTEXT;

int main(void) {

	// Initiates use of the Winsock DLL
	WSADATA wsaData = { 0 };
	int err = WSAStartup(MAKEWORD(2, 2), &wsaData);
	if (err != ERROR_SUCCESS) {
		printf("WSAStartup failed with error: %d\n", err);
		return 1;
	}

	// Open standard input (stdin) with overlapped I/O
	HANDLE hStdIn = CreateFileA("CONIN$", GENERIC_READ, FILE_SHARE_READ, NULL, OPEN_EXISTING, FILE_FLAG_OVERLAPPED, NULL);
	if (hStdIn == INVALID_HANDLE_VALUE) {
		printf("CreateFileA failed to open CONIN$ with error: %d\n", GetLastError());
		return 1;
	}

	// Open standard output (stdout) with overlapped I/O
	HANDLE hStdOut = CreateFileA("CONOUT$", GENERIC_WRITE, FILE_SHARE_WRITE, NULL, OPEN_EXISTING, FILE_FLAG_OVERLAPPED, NULL);
	if (hStdOut == INVALID_HANDLE_VALUE) {
		printf("CreateFileA failed to open CONOUT$ with error: %d\n", GetLastError());
		return 1;
	}

	// Creates a socket that is bound to a specific transport service provider
	SOCKET listenSocket = socket(AF_INET, SOCK_STREAM, IPPROTO_TCP);
	if (listenSocket == INVALID_SOCKET) {
		printf("WSASocket failed with error: %d\n", WSAGetLastError());
		return 1;
	}

	SOCKADDR_IN serverAddr     = { 0 };
	serverAddr.sin_family      = AF_INET;
	serverAddr.sin_addr.s_addr = INADDR_ANY;
	serverAddr.sin_port        = htons(LISTEN_PORT);

	// Associates a local address with a socket
	if (bind(listenSocket, (SOCKADDR*)&serverAddr, sizeof(serverAddr)) == SOCKET_ERROR) {
		printf("bind failed with error: %d\n", WSAGetLastError());
		return 1;
	}

	// Places a socket in a state in which it is listening for an incoming connection
	if (listen(listenSocket, 1) == SOCKET_ERROR) {
		printf("listen failed with error: %d\n", WSAGetLastError());
		return 1;
	}

	printf("Server is listening on port %d\n", LISTEN_PORT);

	// Accept an incoming connection attempt on a socket
	SOCKADDR_IN clientAddr = { 0 };
	int clientAddrLen      = sizeof(clientAddr);

	SOCKET acceptSocket = accept(listenSocket, (SOCKADDR*)&clientAddr, &clientAddrLen);
	if (acceptSocket == INVALID_SOCKET) {
		printf("accept failed with error: %d\n", WSAGetLastError());
		return 1;
	}

	char clientIP[INET_ADDRSTRLEN] = { 0 };
	inet_ntop(AF_INET, &clientAddr.sin_addr, clientIP, INET_ADDRSTRLEN);
	printf("Connection accepted from %s:%d\n", clientIP, ntohs(clientAddr.sin_port));

	// No longer need server socket
	closesocket(listenSocket);

	// Creates an I/O completion port
	HANDLE hIOCompletionPort = CreateIoCompletionPort(INVALID_HANDLE_VALUE, NULL, 0, 1);
	if (hIOCompletionPort == NULL) {
		printf("CreateIoCompletionPort failed with error: %lu\n", GetLastError());
		return 1;
	}

	// Associate the acceptSocket with the completion port
	if (CreateIoCompletionPort((HANDLE)acceptSocket, hIOCompletionPort, 0, 0) == NULL) {
		printf("CreateIoCompletionPort failed to associate %s with error: %lu\n", STRINGIFY(acceptSocket), GetLastError());
		return 1;
	}

	// Associate the STD_INPUT_HANDLE with the completion port
	if (CreateIoCompletionPort(hStdIn, hIOCompletionPort, 0, 0) == NULL) {
		printf("CreateIoCompletionPort failed to associate %s with error: %lu\n", STRINGIFY(hStdIn), GetLastError());
		return 1;
	}

	// Associate the hStdOut with the completion port
	if (CreateIoCompletionPort(hStdOut, hIOCompletionPort, 0, 0) == NULL) {
		printf("CreateIoCompletionPort failed to associate %s with error: %lu\n", STRINGIFY(hStdOut), GetLastError());
		return 1;
	}

	char sockReadBuf[BUFFER_SIZE] = { 0 };
	char fileReadBuf[BUFFER_SIZE] = { 0 };

	IO_CONTEXT ioContextStdIn  = { .ioState = IO_STATE_FILE_READ  };
	IO_CONTEXT ioContextStdOut = { .ioState = IO_STATE_FILE_WRITE };
	
	IO_CONTEXT ioContextSocketRead   = { .ioState = IO_STATE_SOCKET_READ  };
	IO_CONTEXT ioContextSocketWrite  = { .ioState = IO_STATE_SOCKET_WRITE };

	PostQueuedCompletionStatus(hIOCompletionPort, 1, 0, &ioContextStdIn.overlapped);
	PostQueuedCompletionStatus(hIOCompletionPort, 1, 0, &ioContextSocketRead.overlapped);

	while (1) {

		DWORD       dwBytesTransferred = 0;
		ULONG_PTR   ulCompletionKey    = 0;
		PIO_CONTEXT ioContext          = NULL;

		BOOL bSuccess = GetQueuedCompletionStatus(
			hIOCompletionPort,
			&dwBytesTransferred,
			&ulCompletionKey,
			(LPOVERLAPPED*)&ioContext,
			INFINITE);

		if (!bSuccess) {
			printf("GetQueuedCompletionStatus failed with error: %lu\n", GetLastError());
			return 1;
		}

		if (dwBytesTransferred == 0) {
			printf("A send/recv/read/write has failed\n");
			return 2;
		}

		switch (ioContext->ioState) {
			case IO_STATE_SOCKET_READ: {
				// Finished reading from socket, fire a stdout write
				if (!WriteFile(hStdOut, sockReadBuf, dwBytesTransferred, NULL, &ioContextStdOut.overlapped)) {
					if (GetLastError() != ERROR_IO_PENDING) {
						printf("WriteFile failed with error: %lu\n", GetLastError());
						return 1;
					}
				}
				break;
			}
			case IO_STATE_FILE_WRITE: {
				// Finished writing to stdout, fire a new socket read
				ioContextSocketRead.wsaBuf.buf = sockReadBuf;
				ioContextSocketRead.wsaBuf.len = BUFFER_SIZE;

				if (WSARecv(acceptSocket, &ioContextSocketRead.wsaBuf, 1, NULL, &ioContextSocketRead.wsaRecvFlags, &ioContextSocketRead.overlapped, NULL) == SOCKET_ERROR) {
					if (WSAGetLastError() != WSA_IO_PENDING) {
						printf("WSARecv failed with error: %d\n", WSAGetLastError());
						return 1;
					}
				}
				break;
			}
			case IO_STATE_SOCKET_WRITE: {
				// Finished writing to socket, fire a new stdin read
				if (!ReadFile(hStdIn, fileReadBuf, BUFFER_SIZE, NULL, &ioContextStdIn.overlapped)) {
					if (GetLastError() != ERROR_IO_PENDING) {
						printf("ReadFile failed with error: %lu\n", GetLastError());
						return 1;
					}
				}
				break;
			}
			case IO_STATE_FILE_READ: {
				// Finished reading from stdin, fire a new socket write
				ioContextSocketWrite.wsaBuf.buf = fileReadBuf;
				ioContextSocketWrite.wsaBuf.len = dwBytesTransferred;

				if (WSASend(acceptSocket, &ioContextSocketWrite.wsaBuf, 1, NULL, 0, &ioContextSocketWrite.overlapped, NULL) == SOCKET_ERROR) {
					if (WSAGetLastError() != WSA_IO_PENDING) {
						printf("WSASend failed with error: %d\n", WSAGetLastError());
						return 1;
					}
				}
				break;
			}
			default: {
				printf("Invalid ioContext->ioState: %d\n", ioContext->ioState);
				break;
			}
		}
	}
	return 0;
}
```
