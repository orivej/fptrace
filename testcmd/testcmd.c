#define _GNU_SOURCE
#include <fcntl.h>
#include <sched.h>
#include <stdio.h>
#include <stdlib.h>
#include <sys/stat.h>
#include <sys/types.h>
#include <sys/wait.h>
#include <unistd.h>

int execer(void *arg) {
    execlp("cp", "cp", "../b", "../a", NULL);
    perror("execlp");
}

int main() {
    /* puts("chdir"); */
    if (chdir(".")) {
        perror("chdir");
    };

    /* puts("open"); */
    int dirfd = open("testcmd", 0);
    if (dirfd < 0) {
        perror("open testcmd");
    }

    int fd = openat(dirfd, "../a", O_CREAT|O_WRONLY, -1);
    if (fd < 0) {
        perror("open a");
    }
    if (write(fd, "a\n", 2) < 0) {
        perror("write");
    }
    if (rename("a", "b") < 0) {
        perror("rename a b");
    }
    if (close(fd)) {
        perror("close");
    }
    if (renameat(AT_FDCWD, "b", dirfd, "../c") < 0) {
        perror("rename b c");
    }

    int pid = fork();
    if (pid < 0) {
        perror("fork");
    } else if (pid == 0) {
        execlp("cp", "cp", "c", "b", NULL);
        perror("child execlp");
    } else {
        wait(NULL);
    }

    if (fchdir(dirfd)) {
        perror("fchdir");
    };
    if (close(dirfd)) {
        perror("close dir");
    }

    /* execer(NULL); */

    void *stack = malloc(4096000);
    int flags = CLONE_SIGHAND|CLONE_FS|CLONE_VM|CLONE_FILES|CLONE_THREAD;
    pid = clone(execer, stack+4096000, flags, NULL);
    if (pid < 0) {
        perror("clone");
    }
    pause();
}
