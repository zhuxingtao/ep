# websocker server 模型

参照公司生产环境的ws server实现的一个精简版websocket server, 主要用于展示各项功能,数据流转和架构设计
- 约定: 请求都带上 command号 标识不同功能
- hub: 控制 
     - ReadLoop: 从 kafka读取消息, 调度后,写入cli.downChan 发给客户端
     - WriteLoop: 从up读取消息, 写入kafka, 供上游服务,如直播,歌房等处理
     - ControlLoop: 处理client的挂载
- client: 与每个客户端一一对应
    - ReadLoop: 从客户端读取websocket消息, 写入upChan，供hub进行处理，保持cli.upChan=hub.up
    - WriteLoop: 从downchan获取要发送给客户端的消息, 进行发送 
  

