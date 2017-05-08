#include <signal.h>
#include <stdio.h>
#include <string.h>
#include <sys/ptrace.h>
#include <unistd.h>

int main(int argc, char **argv) {
    if (argc < 2) {
        fputs("Arguments: program args...\n", stderr);
        return 1;
    }

    char *args[argc];
    memcpy(args, argv + 1, (argc - 1) * sizeof(char *));
    args[argc - 1] = NULL;

    if (ptrace(PTRACE_TRACEME)) {
        perror("ptrace failed");
    }
    raise(SIGSTOP);
    execvp(args[0], args);
    perror("execvp failed");
}
