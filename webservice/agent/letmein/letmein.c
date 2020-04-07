#define _GNU_SOURCE
#include <errno.h>
#include <sched.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <fcntl.h>
#include <sys/types.h>
#include <sys/wait.h>
#include <sys/mount.h>
#include <signal.h>
#include <unistd.h>
#define STACK_SIZE (1024 * 1024)

static char child_stack[STACK_SIZE];

char* const container_args[] = {
   "/bin/bash",
   NULL
};

int child_main() {
      system("mount -t proc proc /proc");
      execv(container_args[0], container_args);      // Execute a command in namspace
      return 1;
}

main(int argc, char* argv[]) {
    int i;
    char nspath[1024];
    char *namespaces[] = { "ipc", "uts", "net", "pid", "mnt" };

    if (geteuid()) { fprintf(stderr, "%s\n", "abort: you want to run this as root"); exit(1); }

    if (argc != 3) { fprintf(stderr, "%s\n", "abort: you must provide a PID as the sole argument"); exit(2); }

    for (i=0; i<4; i++) {
        sprintf(nspath, "/proc/%s/ns/%s", argv[1], namespaces[i]);
        int fd = open(nspath, O_RDONLY);

        if (setns(fd, 0) == -1) {
            fprintf(stderr, "setns on %s namespace failed: %s\n", namespaces[i], strerror(errno));
        } else {
            fprintf(stdout, "setns on %s namespace succeeded\n", namespaces[i]);
        }

        close(fd);
    }

    sprintf(nspath, "/proc/%s/ns/%s", argv[2], namespaces[4]);
    int fd = open(nspath, O_RDONLY);

    if (setns(fd, 0) == -1) {
        fprintf(stderr, "setns on %s namespace failed: %s\n", namespaces[4], strerror(errno));
    } else {
        fprintf(stdout, "setns on %s namespace succeeded\n", namespaces[4]);
    }

    close(fd);

    system("mount -t proc proc /proc");
    char *cmd[] = {"/bin/sh", NULL};
    execvp("/bin/sh", cmd);      // Execute a command in namspace
    // int child_pid = clone(child_main, child_stack+STACK_SIZE, CLONE_NEWNS | SIGCHLD, NULL);
    // waitpid(child_pid, NULL, 0);
    // printf("Parent - container stopped!\n");
}
