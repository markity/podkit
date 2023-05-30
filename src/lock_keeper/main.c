#define _GNU_SOURCE

#include <sys/file.h>
#include <unistd.h>
#include <signal.h>

// TODO: MORE ELEGANT

// 用来做重启电脑的running_info.json的数据结构更新的
// 因为断电后有些running_info.json里面有些running的container
// 但是其实它们并没有running, 因此我设计了这个进程来做lock_keeper
// 每次运行podkit命令都会先拿lock, 然后拿reboot_lock, 拿不到说明没重启
// 拿到了就说明重启了, 此时启动lock_keeper一直占着文件锁
// 我觉得这个思路有点臃肿, 所以这里TODO: MORE ELEGANT

int main() {
    // 先创建守护进程
    setsid();
    int fd = open("/var/lib/podkit/lock_check_reboot", O_RDONLY, 0);
    flock(fd, LOCK_EX);
    char writeByte = 1;
    write(1, &writeByte, 1);
    sigset_t set;
    sigfillset(&set);
    int sig;
    while(1) {
        // TODO: MORE ELEGANT: 有没有更好的永久阻塞的方法呢?
        sigwait(&set, &sig);
    }
}
