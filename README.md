# SZTU-autologin
SZTU深圳技术大学校园网自动登录脚本 v2.0

觉得好用不妨点个star😋

## 功能特性

- ✅ 支持宿舍区和教学区切换
- ✅ 账号密码便捷配置
- ✅ 命令行交互式菜单
- ✅ 定时检测网络状态，掉网自动重连
- ✅ Windows任务计划开机自启
- ✅ 后台静默运行
- ✅ 完整的日志系统

## 快速开始

### 方式一：使用发布版本（推荐）

1. 从 [Releases](../../releases) 下载最新的 `SZTU-Autologin.exe`
2. 双击运行，按提示完成配置
3. 在菜单中启用开机自启

### 方式二：从源码运行

1. 克隆仓库，使用 `uv` 管理依赖：

```bash
uv sync
uv run main.py
```

### 方式三：打包为可执行文件

```bash
uv add pyinstaller
uv run pyinstaller -F -w --name SZTU-Autologin main.py
```

## 使用说明

首次运行会进入配置向导，按提示输入：
1. 学号
2. 密码
3. 选择运营商（联通/移动/电信）
4. 选择区域（宿舍区/教学区）

主菜单选项：
- [1] 修改账号密码
- [2] 切换区域
- [3] 手动登录/注销
- [4] 开关自动检测重连
- [5] 开关开机自启
- [6] 查看日志
- [7] 测试当前连接
- [0] 退出

## 更新日志

详见 [CHANGELOG.md](./CHANGELOG.md)

## 致谢

本项目基于 [shadow-aaa/SZTU-autologin](https://github.com/shadow-aaa/SZTU-autologin) 重构而来。
