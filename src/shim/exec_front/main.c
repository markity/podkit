#define _GNU_SOURCE

#include <unistd.h>
#include <fcntl.h>
#include <unistd.h>
#include <sys/mount.h>
#include <string.h>
#include <stdio.h>
#include <sys/stat.h>
#include <fcntl.h>
#include <stdlib.h>
#include <sched.h>

// 开启容器, 这里stdout是管道, 告知父进程有没有此命令
// stdin是pty从设备
int main(int argc, char **argv) {
    char *containerIDStr = argv[1];
    char *runCmd = argv[2];

    // 切换namsespcae, pid namespace由父进程负责切换
    char tmpStr[64];
    sprintf(tmpStr, "/var/lib/podkit/container/%s/proc/1/ns/ipc", containerIDStr);
    int ipcNSFD = open(tmpStr, O_RDONLY, 0);
    sprintf(tmpStr, "/var/lib/podkit/container/%s/proc/1/ns/mnt", containerIDStr);
    int mntNSFD = open(tmpStr, O_RDONLY, 0);
    sprintf(tmpStr, "/var/lib/podkit/container/%s/proc/1/ns/uts", containerIDStr);
    int utsNSFD = open(tmpStr, O_RDONLY, 0);
    sprintf(tmpStr, "/var/lib/podkit/container/%s/proc/1/ns/net", containerIDStr);
    int netNSFD = open(tmpStr, O_RDONLY, 0);
    setns(ipcNSFD, 0);
    setns(mntNSFD, 0);
    setns(utsNSFD, 0);
    setns(netNSFD, 0);

    sprintf(tmpStr, "/var/lib/podkit/container/%s", containerIDStr);
    chroot(tmpStr);
    chdir("/");

    // 首先切换root查询是否有这个命令
    if (access(runCmd, X_OK) != 0) {
        char writeByte = 1;
        write(1, &writeByte, 1);
        return 0;
    }

    char writeByte = 0;
    write(1, &writeByte, 1);

    close(1);
    close(2);

    // stdin/out/err都是pty从设备
    dup(0);
    dup(0);

    setsid();
    ioctl(0, TIOCSCTTY, NULL);

    char *arg[argc - 1];
    // arg[0] = argv[2]
    // arg[1] = argv[3]
    arg[0] = runCmd;
    for (int i = 1; i < argc - 2; i++) {
        arg[i] = argv[i + 2];
    }
    arg[argc - 2] = NULL;
    execv(runCmd, arg);
}