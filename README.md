
<img src="https://raw.githubusercontent.com/vearne/chat/master/img/logo.png" height="200px" align="right" />

# chat
[![golang-ci](https://github.com/vearne/chat/actions/workflows/golang-ci.yml/badge.svg)](https://github.com/vearne/chat/actions/workflows/golang-ci.yml)

This chat system is designed to automatically match strangers for chat, so no registration is required

[中文 README](./README_zh.md)

## Online service
[chat.vearne.cc](http://chat.vearne.cc/)

Notice: If no other match is available,
you can open multiple windows and chat with yourself.
## Quick Start with Docker compose
Switch to directory [docker_compose](https://github.com/vearne/chat/tree/master/docker_compose)
```
cd docker_compose
```

### start

```
docker-compose up -d
```

### stop
```
docker-compose down
```
Then you can open a browser and visit
http://localhost/

### Interface
Notice: If no other match is available, 
you can open multiple windows and chat with yourself.
![chat](./img/chat_window.jpg)

### Architecture
![Architecture](./img/arch.png)

### Database Table Design
[database](./docs/mysql.md)

### Websocket Command
[command](./docs/command.md)

### Supporting projects
[chat-ui](https://github.com/vearne/chat-ui)

### Thanks
>"If I have been able to see further, it was only because I stood on the shoulders of giants."   by Isaac Newton

