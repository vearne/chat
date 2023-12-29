#### account

|field|type|illustrate|detail|
|:---|:---|:---|:---|
|id|int|primary key||
|nickname|string|nick name|nickname can be repeated|
|status|int|status|0:created  1: in use 2: destroyed|
|broker|string|broker that user is logged into|When delivering a message, it is required|
|token|string|account unique identifier|Required when reconnecting|
|created_at|date|create time||
|modified_at|date|modify time||

#### session

|field|type|illustrate|detail|
|:---|:---|:---|:---|
|id|int|primary key||
|status|int|status|0:created  1: in use 2: destroyed|
|created_at|date|create time||
|modified_at|date|modify time||

#### session_account

|field|type|illustrate|detail|
|:---|:---|:---|:---|
|id|int|primary key||
|session_id|int|sender ID||
|account_id|int|account ID||


#### outbox

|field|type|illustrate|detail|
|:---|:---|:---|:---|
|id|int|primary key||
|sender_id|int|sender ID||
|session_id|int|session ID||
|status|int|status|0:deleted 1:  normal|
|msg_type|int|message type| 0: dialogue 1: signal|
|content|string|content||
|created_at|date|create time||

#### inbox

|field|type|illustrate|detail|
|:---|:---|:---|:---|
|id|int|primary key||
|sender_id|int|sender ID||
|msg_id|int|message ID||
|receiver_id|int|receiver ID||


#### view ack
|field|type|illustrate|detail|
|:---|:---|:---|:---|
|id|int|primary key||
|session_id|int|session ID||
|account_id|int|account ID||
|msg_id|int|message ID||
|created_at|date|create time||
