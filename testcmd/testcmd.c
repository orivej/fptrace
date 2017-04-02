#include <fcntl.h>
#include <stdio.h>
#include <sys/stat.h>
#include <sys/types.h>
#include <sys/wait.h>
#include <unistd.h>

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

    /* puts("execlp"); */
    execlp("cp", "cp", "a", "b", NULL);
    perror("execlp");
}
