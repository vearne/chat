
#### 用户 account
|字段|类型|说明|备注|
|:---|:---|:---|:---|
|id|int|主键||
|nickname|string|昵称|可以重复|
|status|int|状态|0:创建 1:使用中 2:销毁|
|broker|string|用户登录的broker信息|消息投递时，需要|
|token|string|用户唯一性标识|重连时需要|
|created_at|date|创建时间||
|modified_at|date|更新时间||


#### 会话 session  
|字段|类型|说明|备注|
|:---|:---|:---|:---|
|id|int|会话ID||
|status|int|状态|0:创建 1:使用中 2:销毁|
|created_at|date|创建时间||
|modified_at|date|更新时间||

#### session_account
|字段|类型|说明|备注|
|:---|:---|:---|:---|
|id|int|主键||
|session_id|int|会话ID||
|account_id|int|账号ID||


#### 发件箱 outbox
|字段|类型|说明|备注|
|:---|:---|:---|:---|
|id|int|主键||
|sender_id|int|发出者ID||
|session_id|int|会话ID||
|status|int|状态|0:删除 1: 正常|
|msg_type|int|类型| 0: dialogue 1: signal|
|content|string|内容||
|created_at|date|创建时间||

#### 收件箱 inbox 
|字段|类型|说明|备注|
|:---|:---|:---|:---|
|id|int|主键||
|sender_id|int|发出者ID||
|msg_id|int|消息ID||
|receiver_id|int|接收者ID||


#### view ack
|字段|类型|说明|备注|
|:---|:---|:---|:---|
|id|int|主键||
|session_id|int|会话ID||
|account_id|int|账号ID||
|msg_id|int|消息ID||
|created_at|date|创建时间||
