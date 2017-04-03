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
    execlp("cp", "cp", "b", "c", NULL);
    perror("execlp");
}

int main() {
    /* puts("chdir"); */
    if (chdir(".")) {
        perror("chdir");
    };

    /* puts("open"); */
    int fd = open(".", 0);
    if (fd < 0) {
        perror("open .");
    }

    /* puts("fchdir"); */
    if (fchdir(fd)) {
        perror("fchdir");
    };
    close(fd);

    fd = open("a", O_CREAT|O_WRONLY, -1);
    if (fd < 0) {
        perror("open a");
    }
    if (write(fd, "a\n", 2) < 0) {
        perror("write");
    }
    if (close(fd)) {
        perror("close");
    }

    int pid = fork();
    if (pid < 0) {
        perror("fork");
    } else if (pid == 0) {
        execlp("cp", "cp", "a", "b", NULL);
        perror("child execlp");
    } else {
        wait(NULL);
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
