<img width="250" height="167" alt="1769270281843" src="https://github.com/user-attachments/assets/9e110ef7-b84e-41a7-b131-5eb460a82689" align="center"/>
# Dango：轻量级,去插件化,键盘友好的动态博客
# 帮助文档
# 一.简介
### 1.技术栈
>   Dango博客是基于Golang官方net/http(http1/1.1/2),quic-go(http3)为服务器，goldmark为解析器,以后端goldmark+前端katex插件的混合式解析+ast识别视频解析语法简单的静态博客系统

### 2.特色功能

#### 2.1 集成功能,无插件系统🚫🧩
>  go是强静态类型的语言,做个插件系统，做个插件系统可能会影响复杂性和轻量性🤔，而且看一下wordpress的插件系统，需要动态代码执行, 插件可能会导致安全性受到威胁，而且go对动态代码执行实际上不是很好，官方的库(go有插件系统需要特殊标记// +build plugin，再使用-buildmode=plugin生成so的文件,plugin+反射调用reflect)，以及第三方库(yaegi)对复杂代码的解释器行为不完善故不采用，现有功能已经比较完善，实现了大多数情况下一些博客需要装插件才可以使用的功能，现在基本上我能想到的功能都有了设置项😄
#### 2.2 开箱即用的美观设置,手机电脑界面差分📦
> 默认的毛玻璃风格设计,与背景图片百搭，只要简单更新背景即可一定程度上实现定制化,默认使用差分化移动手机端不同背景，模态框设计
#### 2.3 支持键盘操作⌨️
>主要是聚焦模式,我仿照了一些Linux工具软件(如vim的插入模式,archinstall的空格选定），还有对管理员模态框进行了自动聚焦设置

#### 2.4 轻量,单二进制文件📄
>启动总占用106MB左右1G内存即可高效运行,代码使用内存池优化,二进制文件小于20M(15MB)，使用特定的启动参数即可启动程序

### 3.思路杂谈🤔
> 之前学了golang,简单地学了一些go每日一库,想着做点小东西玩玩，突然想到go还很少有人做可以动态上传的带后台的个人博客系统,而主流的同语言博客(Hugo),又是要命令行操作的，有没有一种更直观的的方式呢,有的兄弟，想wordpress那样使用后台上传，于是我就设计了后台上传文章并常驻在管理员面板上
>关于界面设计,因为我并不是很擅长css，就大概以壁纸美化为聚焦，参考了哲风壁纸的毛玻璃设计在网上找了下代码库,缝好了界面,对于电脑端的文章界面参考了obs的设计风格
>关于取名,其实这个项目的名字,Dango出自key社剧情向作品《团子大家族》,其中讲述探讨了家庭的概念，我取这个名字大概也是因为项目是go语言写的，并且追求界面的统一，美观聚合。

# 二.快速部署
## 1.运行☁️
#### 1.1 最基础的运行
```go
chmod +x myblog-gogogo && ./myblog-gogogo
```
该行为会以默认启动参数在8080端口启动一个http服务器,不启动http服务
> ⚠️ **警告**：如果出现Server error: HTTP server error: listen tcp :8080: bind: address already in use,sh是因为端口占用请使用其他端口，或通过lsof -i 8080;kill `<pid>`来解除占用
#### 1.2 带参数运行
感谢flag库的接口,通过短标签即可识别参数进行运行,一下是常用参数
```sh
  -db-conn string
    	Database connection string (default "./db/data/blog.db")
    	数据库连接地址
  -db-driver string
    	Database driver (sqlite3, mysql, postgres) (default "sqlite3")
    	数据库连接类型,提供数据库连接类型,目前只为sqlite3初始化提供初始化,其他语言只提供接口
  -enable-tls
    	Enable TLS (HTTPS/HTTP3)
    	是否启用http/3,http/2如果启用则必须要配合-tls-cert,-tls-key参数
  -jwt-secret string
    	JWT secret key (leave empty to auto-generate or load from ./data/jwt-secret)
    	jwt密钥不填写会自动生成
  -kafka-brokers string
    	Kafka brokers (comma-separated, leave empty to disable)
    	卡夫卡队列，支持高并发,没有java Kafka服务,服务器内存小于4G不建议启动此选项
  -kafka-group-id string
    	Kafka consumer group ID (default "myblog-consumer-group")
    	卡夫卡集群ID需要配合前面的选项
  -log-level string
    	Log level (debug, info, warn, error) (default "info")
    	日志等级默认info级
  -port string
    	Port to listen on (default "8080")
    	监听的端口
  -tls-cert string
    	Path to TLS certificate file (absolute path)
    	https证书,需要从一些服务商网站请求获取，在本地测试应该填写fullchain.pem
  -tls-key string
    	Path to TLS private key file (absolute path)
    	privkey.pem 或一些key文件
```
### 1.3 长期运行到服务(以通过let's Encrypt certbot为例)
推荐使用systemd服务管理,端口443⭐
推荐写一个脚本管理:
1./etc/systemd/system/myblog.service
```
[Unit]
Description=Dango Blog Service
After=network.target

[Service]
ExecStart=/usr/local/bin/start_Dangoblog.sh
Restart=always
User=root
WorkingDirectory=/path/to/your/file
Environment=PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
Environment=GODEBUG=http2server=0

[Install]
WantedBy=multi-user.target
```
2./usr/local/bin/start_Dangoblog.sh
```sh
#!/bin/bash
/path/to/your/myblog-gogogo -port 443 -enable-tls -tls-cert /etc/letsencrypt/live/你的域名/fullchain.pem -tls-key /etc/letsencrypt/live/你的域名/privkey.pem
```
### 1.4 端口转发或透明代理(可选)

启动你的nginx,或apache服务,以nginx 为例：
####  1.4.1 最简单的条转

```
server {
    listen 80;
    server_name yoursitename yoursitename;

    # 核心：返回301重定向状态码，引导浏览器访问HTTPS
    return 301 https://$server_name$request_uri;  
} 
```

#### 1.4.2 透明代理
Nginx透明代理是一种在不修改客户端配置的情况下，将流量重定向到代理服务器的技术。以下是几种常见的Nginx透明代理实现方案：

##### 1. **透明HTTP/HTTPS代理**

##### 基本HTTP透明代理配置
```nginx
http {
    # 开启透明代理支持
    proxy_set_header Host $http_host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
    
    # 禁用缓存，避免影响动态内容
    proxy_buffering off;
    proxy_request_buffering off;
    
    # 配置上游服务器
    upstream backend_servers {
        server 192.168.1.100:80;
        server 192.168.1.101:80;
        keepalive 32;
    }
    
    server {
        listen 80 default_server;
        listen [::]:80 default_server;
        
        # 透明代理所有HTTP请求
        location / {
            proxy_pass http://backend_servers;
            proxy_http_version 1.1;
            proxy_set_header Connection "";
            
            # 保持连接设置
            proxy_connect_timeout 5s;
            proxy_send_timeout 10s;
            proxy_read_timeout 30s;
        }
    }
}
```

#### 1.4.3 **使用TPROXY进行TCP透明代理**

##### 系统配置（Linux）
```bash
# 启用IP转发
echo "net.ipv4.ip_forward = 1" | sudo tee -a /etc/sysctl.conf
echo "net.ipv4.conf.all.route_localnet = 1" | sudo tee -a /etc/sysctl.conf
sudo sysctl -p

# 设置iptables规则（将80端口流量重定向到Nginx的透明代理端口）
sudo iptables -t mangle -N DIVERT
sudo iptables -t mangle -A PREROUTING -p tcp -m socket --transparent -j DIVERT
sudo iptables -t mangle -A DIVERT -j MARK --set-mark 1
sudo iptables -t mangle -A DIVERT -j ACCEPT

sudo iptables -t mangle -A PREROUTING -p tcp --dport 80 -j TPROXY \
  --tproxy-mark 0x1/0x1 --on-port 8080
```

#### 1.4.4 Nginx TCP透明代理配置
```nginx
stream {
    # 启用TPROXY
    proxy_protocol on;
    
    # 设置解析器（用于DNS解析）
    resolver 8.8.8.8 1.1.1.1 valid=300s;
    resolver_timeout 5s;
    
    # TCP透明代理服务器
    server {
        listen 8080 transparent;  # 启用透明模式
        proxy_pass $upstream;
        proxy_bind $remote_addr transparent;
        
        # 连接参数
        proxy_connect_timeout 5s;
        proxy_timeout 1h;
        proxy_buffer_size 16k;
    }
}
```

#### 1.4.5 **透明SOCKS5代理**
##### Nginx Stream模块配置
```nginx
stream {
    upstream socks_backend {
        server 192.168.1.100:1080;  # SOCKS5后端服务器
    }
    
    server {
        listen 1080;
        proxy_pass socks_backend;
        proxy_connect_timeout 10s;
        proxy_timeout 1h;
    }
}
```

##### 结合iptables实现透明SOCKS代理
```bash
# 设置透明代理规则
sudo iptables -t nat -A OUTPUT -p tcp -d 192.168.1.0/24 -j ACCEPT
sudo iptables -t nat -A OUTPUT -p tcp -j DNAT --to-destination 127.0.0.1:1080
sudo iptables -t nat -A PREROUTING -p tcp -j DNAT --to-destination 127.0.0.1:1080
```

#### 1.4.6 **完整的透明网关配置**

##### 网络拓扑
```
客户端 → 网关(Nginx) → 互联网
```

##### Nginx透明网关配置
```nginx
events {
    worker_connections 1024;
}

http {
    # 基础配置
    log_format transparent '$remote_addr - $remote_user [$time_local] '
                          '"$request" $status $body_bytes_sent '
                          '"$http_referer" "$http_user_agent" '
                          'proxy: "$proxy_host" "$upstream_addr"';
    
    access_log /var/log/nginx/transparent_access.log transparent;
    error_log /var/log/nginx/transparent_error.log;
    
    # 上游DNS解析
    resolver 8.8.8.8 1.1.1.1 valid=300s;
    resolver_timeout 5s;
    
    # 透明代理服务器
    server {
        listen 3128 transparent;
        listen 8080 transparent;
        
        # 动态上游解析
        set $target_host $http_host;
        if ($target_host = "") {
            set $target_host $proxy_host;
        }
        
        location / {
            # 使用变量实现动态代理
            proxy_pass http://$target_host;
            
            # 保持原始请求头
            proxy_set_header Host $http_host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
            proxy_set_header X-Original-URI $request_uri;
            
            # 性能优化
            proxy_http_version 1.1;
            proxy_set_header Connection "";
            proxy_buffering off;
            
            # 超时设置
            proxy_connect_timeout 10s;
            proxy_send_timeout 30s;
            proxy_read_timeout 60s;
            
            # 响应头修改
            proxy_hide_header X-Powered-By;
            proxy_hide_header Server;
            add_header X-Transparent-Proxy "nginx";
        }
    }
}

stream {
    # HTTPS透明代理
    server {
        listen 443 transparent;
        ssl_preread on;  # 启用SNI读取
        
        # 根据SNI动态代理
        proxy_pass $ssl_preread_server_name:443;
        proxy_ssl_name $ssl_preread_server_name;
        proxy_ssl_server_name on;
        
        proxy_connect_timeout 5s;
        proxy_timeout 1h;
    }
}
```

#### 1.4.7 **安装与配置步骤**

##### 安装所需模块
```bash
# 安装Nginx（包含stream模块）
sudo apt install nginx nginx-extras  # Ubuntu/Debian
sudo yum install nginx nginx-mod-stream  # CentOS/RHEL

# 编译安装（自定义模块）
./configure --with-stream --with-stream_ssl_module --with-stream_ssl_preread_module
make && sudo make install
```

##### 系统网络配置
```bash
# 1. 启用IP转发
sudo echo "net.ipv4.ip_forward = 1" >> /etc/sysctl.conf
sudo echo "net.ipv6.conf.all.forwarding = 1" >> /etc/sysctl.conf
sudo sysctl -p

# 2. 设置iptables规则（透明重定向）
# 将80端口流量重定向到Nginx的8080端口
sudo iptables -t nat -A PREROUTING -p tcp --dport 80 -j REDIRECT --to-port 8080
sudo iptables -t nat -A PREROUTING -p tcp --dport 443 -j REDIRECT --to-port 8443

# 3. 保存iptables规则
sudo apt install iptables-persistent  # Ubuntu
sudo netfilter-persistent save

# CentOS/RHEL
sudo service iptables save
sudo chkconfig iptables on
```

#### 1.4.8 **透明代理测试**

##### 验证配置
```bash
# 测试Nginx配置
sudo nginx -t

# 检查端口监听
sudo netstat -tlnp | grep nginx

# 测试代理功能
curl -x http://代理服务器IP:3128 http://example.com
curl --proxy-insecure -x https://代理服务器IP:8443 https://example.com
```

##### 客户端配置
```bash
# 无需特殊配置，系统会自动路由
# 或者设置系统代理（可选）
export http_proxy="http://代理服务器IP:3128"
export https_proxy="http://代理服务器IP:3128"
```
#### 1.4.9 **高级功能**
##### 基于域名的分流
```nginx
map $host $backend {
    default 192.168.1.100:80;
    
    ~^api\.   192.168.1.101:8080;
    ~^static\. 192.168.1.102:80;
    ~^cdn\.   192.168.1.103:80;
}

server {
    listen 8080 transparent;
    
    location / {
        proxy_pass http://$backend;
        # ... 其他配置
    }
}
```

#### 透明代理监控
```nginx
server {
    listen 8081;
    
    location /status {
        stub_status on;
        access_log off;
        allow 127.0.0.1;
        deny all;
    }
    
    location /metrics {
        # Prometheus监控端点
        content_by_lua_block {
            ngx.say("# HELP nginx_connections Active connections")
            ngx.say("# TYPE nginx_connections gauge")
            ngx.say("nginx_connections " .. ngx.var.connections_active)
        }
    }
}
```
### 1.5 常见问题🤔
#### 1.5.1 服务器端口占用

```bash
lsof -i :8080 或者netstat -tulpn | grep :8080
```
#### 1.5.2 因为可用内存小于150M而在启动时oom被killed
##### 1.排查相关的服务占用
```bash
最基础的命令
ps aux 
top
或者排查前十占用最高的用户
ps aux --sort=-%mem | head -11
```
##### 2.结束并禁用你不需要的服务
```bash
kill <pid>
pkill <name>
systemctl stop xxx
systemctl disable xxx
```
##### 3.如果还是不够，请分配swap
```bash
# 创建 8GB swap 文件
sudo fallocate -l 8G /swapfile
# 或使用 dd（fallocate 不适用某些文件系统）
sudo dd if=/dev/zero of=/swapfile bs=1M count=8192
# 设置权限
sudo chmod 600 /swapfile
# 格式化为 swap
sudo mkswap /swapfile
# 启用
sudo swapon /swapfile
# 永久生效
echo '/swapfile none swap sw 0 0' | sudo tee -a /etc/fstab
```
#### 1.5.3 启动成功后登录显示会话不存在
这种情况可能发生再更新版本重新部署时,是有可能因为ecc失效需要手动清理jwt-key,并重启服务
```bash
cd your/path/to/binary && rm ./data/jwt-secret
```

# 三.日常使用⚙️
### 1.认识界面
![[1769236168597.jpg]]
手机端如上
其他界面同理
### 2.特色功能
#### 2.1快捷键
##### 2.1.1 聚焦模式(适用于电脑)
在大多数电脑端界面的大多数界面你可以使用i来进入聚焦模式q和esc退出快捷键,聚焦模式有点类似于yazi的内容区域切换
##### 2.2.2 全部快捷键功能列表
###### 🔢 导航快捷键

| 快捷键 | 功能                 |
| :----- | :------------------- |
| 1      | 跳转到主页           |
| 2      | 跳转到文章页         |
| 3      | 跳转到归档页         |
| 4      | 跳转到关于页         |
| 5      | 打开个人中心模态框   |
| 6      | 跳转到 Markdown 编辑器 |

###### ⚙️ 功能快捷键

| 快捷键    | 功能      |
| :----- | :------ |
| l      | 打开登录模态框 |
| /      | 显示快捷键帮助 |
| Escape | 关闭所有模态框 |

###### 应用内快捷键

###### 🎵 音乐播放器快捷键

| 快捷键 | 功能             |
| :----- | :--------------- |
| Space  | 播放/暂停        |
| ←      | 上一首           |
| →      | 下一首           |
| ↑      | 音量+            |
| ↓      | 音量-            |
| m      | 静音/取消静音    |
| p      | 打开/关闭播放列表 |

##### 📝 Markdown 编辑器快捷键
*（仅在编辑器内生效）*

| 快捷键 | 功能                 |
| :----- | :------------------- |
| Ctrl+S | 保存                 |
| Ctrl+D | 下载 Markdown 文件   |
| Ctrl+B | 插入粗体 **text**    |
| Ctrl+I | 插入斜体 *text*      |
| Ctrl+K | 插入行内代码 `code`  |
| Ctrl+M | 插入数学公式 $$...$$ |
| Ctrl+G | 插入 Mermaid 流程图  |
| Tab    | 插入2个空格          |

###### 📄 文章页面快捷键

| 快捷键    | 功能     |     |
| :----- | :----- | --- |
| s      | 切换侧边栏  |     |
| t      | 切换阅读模式 |     |
| h      | 切换标题栏  |     |
| f      | 全屏模式   |     |
| Escape | 退出全屏   |     |

##### 🎯 聚焦模式快捷键

###### 文章页面聚焦模式

| 快捷键   | 功能       |     |
| :---- | :------- | --- |
| i     | 进入文本聚焦模式 |     |
| q     | 退出文本聚焦模式 |     |
| ← →   | 切换面板     |     |
| ↑ ↓   | 导航       |     |
| Enter | 激活       |     |
| u     | 展开/折叠    |     |

###### 归档页面聚焦模式

| 快捷键     | 功能         |
| :------ | :--------- |
| i       | 进入聚焦模式     |
| q       | 退出聚焦模式     |
| ← → ↑ ↓ | 导航         |
| Enter   | 进入子菜单/激活   |
| Escape  | 返回上一级或暂时退出 |
|         |            |

###### 关于页面聚焦模式

| 快捷键 | 功能         |
| :----- | :----------- |
| i      | 进入聚焦模式 |
| q      | 退出聚焦模式 |
| ↑ ↓    | 导航卡片     |
| Enter  | 查看卡片内容 |

##### 👨‍💼 管理员面板快捷键

###### 标签页切换

| 快捷键 | 标签页     |
| :----- | :--------- |
| 1      | 文章管理   |
| 2      | 用户管理   |
| 3      | 评论管理   |
| 4      | 分类管理   |
| 5      | 标签管理   |
| 6      | 统计分析   |
| 7      | 关于页面   |
| 8      | 文件管理   |
| 9      | 附件管理   |
| 0      | 系统设置   |

###### 通用快捷键

| 快捷键 | 功能                    |
| :----- | :---------------------- |
| i      | 进入聚焦模式            |
| q      | 退出聚焦模式            |
| Escape | 关闭模态框/退出聚焦模式 |
| ← →    | 切换标签页              |
| r      | 刷新当前标签页          |
| n      | 新建项目                |
| u      | 上传项目                |
| f      | 打开搜索                |

###### 表格操作快捷键

| 快捷键 | 功能                             |
| :----- | :------------------------------- |
| ↑ ↓    | 选择行                           |
| Enter  | 打开选中项                       |
| e      | 编辑选中项                       |
| d      | 删除选中项                       |
| v      | 查看选中项（文章）               |
| a      | 附加/添加/批准（文章/分类/评论） |
| p      | 发布选中项（文章）               |

###### 模态框快捷键

| 快捷键    | 功能        |
| :----- | :-------- |
| Enter  | 提交表单/确认   |
| s      | 保存        |
| y      | 确认操作      |
| c      | 取消/关闭     |
| Tab    | 循环导航      |
| Space  | 切换单选框/复选框 |
| Escape | 关闭模态框     |

###### 文件管理器专用

| 快捷键       | 功能      |
| :-------- | :------ |
| Enter     | 打开选中文件  |
| Backspace | 返回上一级目录 |
| r         | 刷新/重命名  |
| Delete    | 删除选中文件  |

###### 系统设置快捷键

| 快捷键       | 功能              |
| :-------- | :-------------- |
| 1         | 外观设置            |
| 2         | 音乐设置            |
| 3         | 模板设置            |
| 4         | 文章标题设置          |
| 5         | 切换界面提示设置        |
| 6         | 外部链接设置          |
| 7         | 赞助设置            |
| Tab       | 表单控件导航          |
| Shift+Tab | 反向导航            |
| Space     | 切换复选框           |
| ↑ ↓       | 下拉框切换           |
| s         | 保存当前区块（输入框中也可用） |
| r         | 重置为默认           |
| q         | 退出聚焦模式（输入框中也可用） |
| ?         | 显示设置快捷键帮助             |
 
✦ 注意：部分快捷键在输入框中输入时不会触发，除非特别标注（如系统设置中的
  s、q、数字键）。移动端设备上不显示快捷键提示。
#### 2.2 附件系统
📎 附件系统完整功能列表

##### 支持的文件类型

| 类型 | 扩展名 | 文件类型标识 |
| :--- | :--- | :--- |
| 图片 | jpg, jpeg, png, gif, bmp, svg, webp | Image |
| 文档 | pdf, doc, docx, xls, xlsx, ppt, pptx | Document |
| 视频 | mp4, webm, flac | Video |
| 音频 | mp3, flac | Audio |
| 压缩包 | zip, rar, 7z, tar, gz, tar.gz | Archive |

###### 🔒 安全特性

1. **文件类型验证**: 基于扩展名和内容检测
2. **文件安全扫描**: 检测恶意文件
3. **大小限制**: 最大 500MB（可配置）
4. **权限控制**: 三级可见性系统
5. **路径安全**: 自动生成唯一文件名，防止覆盖

##### 📊 管理功能

##### 管理员面板功能
- 查看所有附件列表
- 按文章、可见性筛选
- 上传附件
- 下载附件
- 删除附件
- 批量删除
- 更新可见性
- 控制是否在文章中显示

##### 系统配置

| 配置项 | 默认值 | 说明 |
| :--- | :--- | :--- |
| attachment_default_visibility | public | 默认可见性 |
| attachment_max_size | 500MB | 最大附件大小 |
| attachment_allowed_types | jpg,jpeg,png,gif,mp4,mp3... | 允许的文件类型 |

##### 🎯 文章集成

###### 文章页面附件显示
- 自动加载文章关联的附件
- 只显示 visibility=public 且 show_in_passage=true 的附件
- 支持按日期加载附件（用于无 passage_id 的旧文章）
- 显示附件图标、名称、大小、类型

##### 附件展示样式
- **图片附件**: 显示缩略图
- **视频/音频**: 显示播放图标
- **文档/压缩包**: 显示文件类型图标

##### 🔄 同步功能

###### SyncToDB 功能
- 扫描 attachments 目录
- 将未记录的文件同步到数据库
- 自动提取原始文件名
- 设置默认可见性为 public
- 默认不在文章中显示
# 四.高级功能
请见开发者文档
