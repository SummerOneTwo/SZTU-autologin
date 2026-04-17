# SZTU 校园网自动登录工具

深圳技术大学校园网自动登录工具，Windows 单文件 Go 版本。

## 功能

- 交互式主菜单与 `setup` 配置流程
- 立即登录与后台自动重连
- Windows 开机自启动管理

## 使用

```bash
# 交互式主菜单
sztu-autologin.exe

# 首次配置
sztu-autologin.exe setup

# 立即登录
sztu-autologin.exe login

# 后台运行（自动重连）
sztu-autologin.exe daemon

# 开机自启动管理
sztu-autologin.exe autostart on     # 启用
sztu-autologin.exe autostart off    # 禁用
sztu-autologin.exe autostart status # 查看状态

# 帮助
sztu-autologin.exe help
```

## 配置文件

配置保存在 `config.json`，可手动编辑：

```json
{
  "username": "学号",
  "password": "密码",
  "isp": "cucc",
  "ac_id": "17",
  "auto_reconnect": false,
  "check_interval": 300
}
```

### 运营商选项

- `cucc` - 中国联通
- `cmcc` - 中国移动
- `chinanet` - 中国电信

> **注意**: 暂时只支持宿舍区 (ac_id=17)，教学区设置暂未确定。

## 编译

```bash
go build -o sztu-autologin.exe
```

## 测试

```bash
go test ./...
```

如果本机默认 `GOCACHE` 目录权限异常，可临时指定项目内缓存目录：

```bash
$env:GOCACHE = (Join-Path (Get-Location) ".gocache")
go test ./...
```

## 行为说明

- 默认关闭自动重连，避免首次配置后未经确认即进入守护模式。
- 配置文件保存为 `0600` 权限，密码仍为明文存储。

## 从 Python 版本迁移

如果你之前使用 Python 版本：
1. 删除旧的 `config.json`（格式不兼容）
2. 运行 `sztu-autologin.exe setup` 重新配置
3. 删除旧的 Python 文件和 `autologin.key`
