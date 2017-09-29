#include <signal.h>
#include <stdio.h>
#include <sys/ptrace.h>
#include <unistd.h>

int main(int argc, char **argv) {
    if (argc < 2) {
        fputs("Arguments: program args...\n", stderr);
        return 1;
    }
    if (ptrace(PTRACE_TRACEME)) {
        perror("ptrace failed");
    }
    raise(SIGSTOP);
    execvp(argv[1], argv + 1);
    perror("execvp failed");
}
