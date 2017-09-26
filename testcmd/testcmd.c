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
    int dirfd = open("testcmd", O_CLOEXEC);
    if (dirfd < 0) {
        perror("open testcmd");
    }

    int fd = openat(dirfd, "../a", O_CREAT|O_WRONLY, -1);
    if (fd < 0) {
        perror("open a");
    }
    int fd2 = dup(fd);
    if (write(fd2, "a\n", 2) < 0) {
        perror("write");
    }
    if (rename("a", "b") < 0) {
        perror("rename a b");
    }
    if (renameat(AT_FDCWD, "b", dirfd, "../c") < 0) {
        perror("rename b c");
    }

    int pipefd[2];
    if (pipe(pipefd)) {
        perror("pipe");
    }
    char *pipe_r_path, *pipe_w_path;
    if (asprintf(&pipe_r_path, "/dev/fd/%d", pipefd[0]) < 0) {
        perror("asprintf pipe_r");
    }
    if (asprintf(&pipe_w_path, "/proc/self/fd/%d", pipefd[1]) < 0) {
        perror("asprintf pipe_w");
    }
    int pipe_r = open(pipe_r_path, O_RDONLY);
    if (pipe_r < 0) {
        perror("open pipe_r");
    }
    int pipe_w = open(pipe_w_path, O_WRONLY);
    if (pipe_w < 0) {
        perror("open pipe_w");
    }

    int pid = fork();
    if (pid < 0) {
        perror("fork");
    } else if (pid == 0) {
        char buf;
        if (read(pipefd[0], &buf, 1) != 1 || read(pipe_r, &buf, 1) != 1) {
            perror("read pipe_r");
        }

        execlp("cp", "cp", "c", "b", NULL);
        perror("child execlp");
    } else {
        if (write(pipefd[1], "a", 1) != 1 || write(pipe_w, "b", 1) != 1) {
            perror("write pipe_w");
        }

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
