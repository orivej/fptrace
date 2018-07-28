#include "seccomp.h"

#include <linux/seccomp.h>
#include <signal.h>
#include <stdbool.h>
#include <stdio.h>
#include <string.h>
#include <sys/prctl.h>
#include <sys/ptrace.h>
#include <unistd.h>

int main(int argc, char **argv) {
    int sep;
    for (sep = 1; sep < argc && strcmp(argv[sep], "--") != 0; sep++);
    if (sep >= argc - 1) {
        fputs("Arguments: [-seccomp] -- program args...\n", stderr);
        return 1;
    }
    bool withSeccomp = sep > 1 && strcmp(argv[sep - 1], "-seccomp") == 0;

    if (withSeccomp) {
        if (prctl(PR_SET_NO_NEW_PRIVS, 1, 0, 0, 0) < 0) {
            perror("no_new_privs failed");
        }
        if (prctl(PR_SET_SECCOMP, SECCOMP_MODE_FILTER, &seccomp_program) < 0) {
            perror("seccomp failed");
        }
    }
    if (ptrace(PTRACE_TRACEME)) {
        perror("ptrace failed");
    }
    raise(SIGSTOP);
    execvp(argv[sep + 1], argv + sep + 1);
    fprintf(stderr, "execvp '%s'", argv[sep + 1]);
    perror(" failed");
}
