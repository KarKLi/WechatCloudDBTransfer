# 微信小程序云数据库存储迁移工具

本工具为UOSP（Utopia Open Source Project）的组成部分。

使用Golang作为Server端，Python作为Client端，请先安装好相应的环境

Golang版本：1.12及以上

Python版本：3.7.4及以上

# Client端

**作者：Tiga江（Utopia前端组）**

Client端使用Server端预定的URL发起迁移请求，Client端默认的URL格式为：

```
https://<your_upload_URL>?path=<file_name>&id=<file_id>
```

你可以自行修改your_upload_URL为自己服务器网关设置的URL。

# Server端

**作者：Kark李（Utopia后端组）**

Server端监听由网关转发到指定端口的请求，你可以在Nginx里面进行location的设置，通过proxy_pass将请求转发至该端口，一个Nginx的conf例子（假设使用默认的20000端口）：

```
location /cloud-transfer/ {
            proxy_pass http://localhost:20000/;
}
```

则你发送的URL可以为（保持20000端口无其他应用占用）：

```
https://<your_domain>/cloud-transfer?path=<file_name>&id=<file_id>
```

如果你没有域名，那你可以：

```
http://<your_ip>/cloud-transfer?path=<file_name>&id=<file_id>
```

在使用之前，请将Server端const变量中的新旧版本小程序的各参数填写完毕，方可正常运行。

请注意：如果access_token未能成功获得，server端是无法启动的。

<center><b><i>Utopia © Copyright 2020. All rights reserved.</i></b></center>

<img src="https://raw.githubusercontent.com/KarKLi/WechatCloudDBTransfer/main/DCIM.png" style="float: left;" />

