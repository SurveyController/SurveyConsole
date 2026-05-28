# SurveyConsole

SurveyController 的无图形化界面实现 — 轻量、高性能的命令行问卷自动化工具。

> 本项目仅供获得授权的学习与测试使用。请勿用于污染第三方问卷数据。

## 支持平台

- **问卷星 (WJX)** - 完整支持
- **腾讯问卷 (QQ)** - 完整支持
- **Credamo 见数** - 完整支持

## 功能特性

- 问卷结构自动解析 (3 个平台)
- 多题型答案生成（单选、多选、矩阵、量表、文本、滑块、排序）
- 概率分布配置
- 运行时分布修正 (Distribution Tracking)
- 一致性规则引擎 (Answer Rules)
- 心理计量优化 (Psychometric Plan)
- AI 主观题作答
- 并发提交引擎
- 代理池管理（随机 IP）
- 二维码解析
- Excel 报告导出
- JSON 配置导入/导出
- CLI 命令行界面

## 快速开始

```bash
# 克隆
git clone https://github.com/SurveyController/SurveyConsole.git
cd SurveyConsole

# 构建
go build -o surveyconsole ./cmd/surveycontroller

# 运行测试
go test ./...
```

## 使用方法

### 解析问卷

```bash
# 问卷星
./surveyconsole parse -url "https://www.wjx.cn/vm/xxxxx.aspx"

# 腾讯问卷
./surveyconsole parse -url "https://wj.qq.com/s2/123456/abc/"

# Credamo 见数
./surveyconsole parse -url "https://www.credamo.com/s/xxxxx"
```

### 创建配置

```bash
./surveyconsole config -create -url "https://www.wjx.cn/vm/xxxxx.aspx" -output config.json
```

### 运行提交任务

```bash
# 基本用法
./surveyconsole run -url "https://www.wjx.cn/vm/xxxxx.aspx" -target 10 -threads 3

# 使用配置文件
./surveyconsole run -config config.json

# 启用随机 IP
./surveyconsole run -config config.json -random-ip -proxy-source custom -custom-proxy "http://api.example.com/proxy"
```

### 二维码解析

```bash
./surveyconsole qr -image qrcode.png
```

### 导出报告

```bash
./surveyconsole export -config config.json -output report.xlsx
```

## 命令行参数

| 命令 | 说明 | 主要参数 |
|------|------|----------|
| `run` | 运行提交任务 | `-config`, `-url`, `-target`, `-threads`, `-random-ip` |
| `parse` | 解析问卷结构 | `-url` |
| `config` | 配置管理 | `-create`, `-url`, `-output` |
| `qr` | 解析二维码 | `-image` |
| `export` | 导出 Excel 报告 | `-config`, `-output` |

## 配置文件格式

```json
{
  "url": "https://www.wjx.cn/vm/xxxxx.aspx",
  "survey_provider": "wjx",
  "target": 10,
  "threads": 3,
  "submit_interval": [0, 0],
  "answer_duration": [10, 20],
  "random_ip_enabled": false,
  "proxy_source": "default",
  "question_entries": [
    {
      "question_type": "single",
      "probabilities": [0.25, 0.25, 0.25, 0.25],
      "option_count": 4
    }
  ],
  "answer_rules": [
    {
      "condition_question_num": 1,
      "condition_mode": "selected",
      "condition_option_indices": [0],
      "target_question_num": 2,
      "action_mode": "must_not_select",
      "target_option_indices": [1]
    }
  ]
}
```

## 项目结构

```
SurveyConsole/
├── cmd/surveycontroller/     # CLI 入口
├── internal/
│   ├── config/               # 配置管理
│   ├── models/               # 数据模型
│   ├── providers/            # 问卷平台适配器
│   │   ├── wjx/             # 问卷星
│   │   ├── tencent/         # 腾讯问卷
│   │   └── credamo/         # Credamo 见数
│   ├── engine/               # 执行引擎
│   ├── network/              # 网络层
│   │   ├── httpclient/       # HTTP 客户端
│   │   └── proxy/            # 代理管理
│   ├── questions/            # 答案生成 & 心理计量
│   ├── io/                   # 二维码 & Excel
│   └── logging/              # 日志
├── tests/                    # 测试
└── configs/                  # 示例配置
```

## 高级功能

### 一致性规则 (Answer Rules)

定义条件规则，当某题选择了特定选项时，自动约束其他题的答案：

```json
{
  "answer_rules": [
    {
      "condition_question_num": 1,
      "condition_mode": "selected",
      "condition_option_indices": [0],
      "target_question_num": 3,
      "action_mode": "must_not_select",
      "target_option_indices": [2, 3]
    }
  ]
}
```

### 心理计量优化 (Psychometric Plan)

使用 IRT 模型生成心理测量学上有效的答案，自动达到目标 Cronbach's Alpha 信度系数。

### 运行时分布修正

根据实际提交的选项分布，动态调整概率以趋近目标分布。

## 运行内核方向

后续版本会支持三种可选运行内核：

- `hybrid`：默认模式。优先保证浏览器兼容性；当平台适配器明确支持安全复用请求时，启用 HTTP 快速路径。
- `browser`：纯浏览器模式，优先追求兼容性。
- `http`：纯 HTTP 模式；当所选平台不支持时，必须清晰报错，不自动降级。

## License

MIT
