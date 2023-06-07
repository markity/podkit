### PodKit

### 当前状态

docker的拙劣模仿, 正在施工...预计使用以下机制:

- linux namespace隔离
    - ipc - OK
    - mnt - OK
    - pid -OK
    - net - - OK
    - user - TODO
- 使用pty设备交互 - OK
- veth网络 - OK
- 断电安全 - OK
- 使用特定用户和用户组隔离权限 - TODO
- 打包容器 - TODO

### 安装和卸载(安装卸载需要root)

安装:

```
0) 保证电脑上的gcc和go命令可用
1) git clone git@github.com:markity/podkit.git
2) cd podkit
3) make
```

卸载:

```
1) 重启电脑
2) cd podkit
3) make clean
```

### 基本使用(执行podkit需要root):

查看帮助:

```
# podkit --help
Podkit helps you better understand the mechanism of docker
It provides main functions to get you understand how docker works

Usage:
  podkit [command]

Available Commands:
  container   container command is used to manage containers
  help        Help about any command
  image       ls command is used to manage images

Flags:
  -h, --help   help for podkit

Use "podkit [command] --help" for more information about a command.
```

查看所有镜像:

```
# podkit image ls
busybox
ubuntu22.04
```

运行容器:

```
# podkit image start ubuntu22.04
Extracting /var/lib/podkit/images/ubuntu2204.tar
succeed: container id is 1
```

进入容器执行bash:

```
# podkit container exec -i 1 /bin/bash
root@container1:/# 
```

> -i的全称是interactive, 这条命令会为命令创建pty伪终端设备, 然后持续与命令进行交互, 直到命令结束。对bash来说, 它的stdin/out/err都是pts(pty slave device)

执行后台任务, 下面给个演示证明进程确实后台运行了:

```
# podkit container exec 1 /bin/sleep 100
ok, command now is running
$ podkit container exec -i 1 /bin/bash
root@container1:/# ps -ef
UID          PID    PPID  C STIME TTY          TIME CMD
root           1       0  0 10:55 ?        00:00:00 podkit_orphan_reaper
root          16       0  0 11:00 ?        00:00:00 /bin/sleep 100
root          17       0  0 11:00 pts/0    00:00:00 /bin/bash
root          20      17  0 11:00 pts/0    00:00:00 ps -ef
root@container1:/# 
```

> 后台运行和-i运行的区别之一就是不等待程序结束, 而是直接启动程序。另外一个区别就是后台命令的stdin/out/err都是/dev/null设备

停止容器:

```
# podkit container stop 1
conatiner 1 closed successfully
```

查看所有容器:

```
# podkit container ls
---stopped---:
1 ubuntu22.04 172.16.0.2
---running---:
(none)
```

重启停止的容器(只能重启停止的容器):

```
# podkit container restart 1
restarting...
container 1 restarted successfully
```

删除容器(只能删除停止的容器):

```
# podkit container stop 1
conatiner 1 closed successfully
$ podkit container remove 1
removed container 1 successfully
```