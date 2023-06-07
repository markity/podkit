### PodKit

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