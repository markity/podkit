### PodKit

docker的拙劣模仿, 正在施工...预计使用以下机制:

- linux namespace隔离
    - ipc - OK
    - mnt - OK
    - pid -OK
    - net - NOT YET
    - user - NOT YET
- 使用pty设备交互 - OK
- veth网络 - NOT YET
- 断电安全 - OK