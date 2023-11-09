# 设备与接入层交互协议
websocket + 自定义协议(JSON)


```
    client                      broker

# 请求创建一个临时账号
CRT_ACCOUNT_REQ       -->
                      <--     CRT_ACCOUNT_RESP
# 请求匹配一个聊天对象                    
MATCH_REQ             -->
                      <--     MATCH_RESP
# 发出一条对话消息
DIALOGUE_REQ          -->     
                      <--     DIALOGUE_RESP

# 收到一条对话消息                   
                      <--     PUSH_DIALOGUE_REQ
PUSH_DIALOGUE_RESP    -->     
                    
# 收到一条系统消息(聊天对象下线等等)
                      <--     PUSH_SIGNAL_REQ
PUSH_SIGNAL_RESP      -->     

# Ping
PING_REQ              -->     
                      <--     PING_RESP

# 告知对方消息已经被已方阅读                    
VIEWED_ACK_REQ        -->
                      <--     VIEWED_ACK_RESP
  
# 收到已方消息已经被对方阅读               
                      <--     PUSH_VIEWED_ACK_REQ
PUSH_VIEWED_ACK_RESP  -->     

# 连接意外断开后，尝试重新连接
RECONNECT_REQ         -->     
                      <--     RECONNECT_RESP          

```

## 时序图
![](../img/seq.jpeg)