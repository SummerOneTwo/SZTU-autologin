# SZTU 校园网自动登录工具

深圳技术大学校园网自动登录工具，Go 重构版本。

## 功能

- 交互式配置账号密码
- 立即登录
- 后台自动检测重连
- Windows 开机自启动

## 使用

```bash
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
  "area": "dormitory",
  "ac_id": "17",
  "auto_reconnect": true,
  "check_interval": 300
}
```

### 运营商选项

- `cucc` - 中国联通
- `cmcc` - 中国移动
- `chinanet` - 中国电信

### 区域选项

- `dormitory` - 宿舍区 (ac_id=17)
- `teaching` - 教学区 (ac_id=1)

## 编译

```bash
go build -o sztu-autologin.exe
```

## 体积

编译后约 2-3MB，无额外依赖。

## 从 Python 版本迁移

如果你之前使用 Python 版本：
1. 删除旧的 `config.json`（格式不兼容）
2. 运行 `sztu-autologin.exe setup` 重新配置
3. 删除旧的 Python 文件和 `autologin.key`
