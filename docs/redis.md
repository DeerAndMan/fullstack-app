# redis 基础

创建时间 2025-04-12

------

# 安装

```bash
sudo apt-get update  
sudo apt install redis-server (debian 12)
sudo systemctl enable redis # 设置开机自启

brew install redis (mac)

# 配置文件中需要修改部分配置 /etc/redis/redis.conf
bind 0.0.0.0 -::1
protected-mode no
requirepass 自定义密码
```

# 启动

```bash
redis-server
# 后台启动
brew services start redis (mac)
# debian12 启动配置文件，适用于希望自定义启动流程或设置开机自启场景
redis-server /etc/redis/redis.conf
# 测试是否开启
redis-cli ping
# 添加密码
requirepass 后面加设置的密码 rediskey121qaz


# 重启
ps -ef | grep redis
sudo kill -9 $PID
redis-server /etc/redis/redis.conf
```

# 连接

```bash
# 本地连接
redis-cli -h 127.0.0.1 -p 6379

# 远程连接
redis-cli -h <远程服务器IP> -p <端口号> -a <密码>
redis-cli -h 38.38.251.129 -p 6379 # 无密码
```





# 基础操作

```bash
SET 'key name' "name value" # (key name 不需要引号)
GET 'ke name' # (key name 不需要引号)
```

