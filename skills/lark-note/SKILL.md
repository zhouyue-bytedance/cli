---
name: lark-note
version: 1.0.0
description: "飞书会议纪要（Note）直查：已知 note_id 时查询纪要详情、展示类型（普通纪要 / 三合一纪要）、关联文档 token，以及三合一纪要的原始逐字记录（unified transcript）。1. 用户已经持有 note_id 并想查纪要元信息、纪要类型、纪要/逐字稿文档 token 时使用本技能。2. 三合一（unified）纪要的逐字稿不是独立文档，必须用 note +transcript 按 note_id 拉取。3. 本技能只接受 note_id 入口；只有 meeting_id / calendar_event_id / minute_token 等会议线索时，请先用 lark-vc 定位 note_id；要读纪要正文 / 逐字稿文档请用 lark-doc。"
metadata:
  requires:
    bins: ["lark-cli"]
  cliHelp: "lark-cli note --help"
---

# note (v1)

**CRITICAL — 开始前 MUST 先用 Read 工具读取 [`../lark-shared/SKILL.md`](../lark-shared/SKILL.md)，其中包含认证、权限处理。**

Note 域负责**已知 `note_id`** 时的纪要直查。它不反查会议、日程或妙记，也不读取 Docx 正文——那些分别属于 `lark-vc`、`lark-minutes`、`lark-doc`。

## 核心概念

- **Note（会议纪要）**：会议结束后生成的纪要实体，通过 `note_id` 标识。
- **展示类型（`note_display_type`）**：区分纪要形态，取值 `unknown` / `normal` / `unified`。
  - `normal`（普通纪要）：纪要正文和逐字稿是两份独立的飞书文档，分别对应 `note_doc_token`、`verbatim_doc_token`。
  - `unified`（三合一纪要）：纪要正文、AI 产物、逐字记录合并呈现；**逐字稿不再是独立文档**，要用 `note +transcript` 按 `note_id` 拉取原始记录。
- **文档 token**：`note_doc_token`（AI 智能纪要主文档）、`verbatim_doc_token`（普通纪要逐字稿文档）、`shared_doc_tokens`（会中共享文档）。拿到 token 后读正文交给 [lark-doc](../lark-doc/SKILL.md)。

## 触发规则

| 用户表达 | 命令 |
|---------|------|
| 已知 `note_id`，查纪要详情 / 纪要类型 / 关联文档 token | `note +detail --note-id NOTE_ID` |
| 已知 `note_id`，查三合一（unified）原始记录 / 逐字稿 | `note +transcript --note-id NOTE_ID` |
| 已知 `note_id`，读纪要正文 | 先 `note +detail` 拿 `note_doc_token`，再调 `docs +fetch --api-version v2 --doc <note_doc_token>` |

## 路由规则（拿到 detail 后按 `note_display_type` 决策）

| 条件 | Agent 后续动作 |
|------|---------------|
| 用户要纪要正文 / 总结 / 待办 / 章节 | `docs +fetch --api-version v2 --doc <note_doc_token>` |
| `note_display_type=normal`，用户要逐字稿 / 谁说了什么 | `docs +fetch --api-version v2 --doc <verbatim_doc_token>` |
| `note_display_type=unified`，用户要逐字稿 / 原始记录 / 谁说了什么 | `note +transcript --note-id <note_id>` |

> **判别键是 `note_display_type`，不是 `verbatim_doc_token` 是否为空。** unified 纪要的 `verbatim_doc_token` 也可能有值，但 unified 的逐字稿应统一走 `note +transcript`（输出更结构化）。

## 禁止规则

- 不处理 `meeting_id` —— 那是 [lark-vc](../lark-vc/SKILL.md) 的入口。
- 不处理 `calendar_event_id` —— 那是 [lark-vc](../lark-vc/SKILL.md) 的入口。
- 不处理 `minute_token` —— 那是 [lark-vc](../lark-vc/SKILL.md)（纪要产物索引）/ [lark-minutes](../lark-minutes/SKILL.md)（妙记基础信息与媒体）的入口。
- 不读取 Docx 正文 —— 拿到文档 token 后交给 [lark-doc](../lark-doc/SKILL.md)。
- 不从纪要正文或 `doc_token` 反推 `note_id`。

## Shortcuts（推荐优先使用）

Shortcut 是对常用操作的高级封装（`lark-cli note +<verb> [flags]`）。

| Shortcut | 说明 |
|----------|------|
| [`+detail`](references/lark-note-detail.md) | Get note detail (display type, document tokens) by note_id |
| [`+transcript`](references/lark-note-transcript.md) | Fetch the unified (three-in-one) note transcript and save it to a file |

- 使用 `+detail` 命令时，必须阅读 [references/lark-note-detail.md](references/lark-note-detail.md)。
- 使用 `+transcript` 命令时，必须阅读 [references/lark-note-transcript.md](references/lark-note-transcript.md)。

## 权限表

| 方法 | 所需 scope |
|------|-----------|
| `+detail` | `vc:note:read` |
| `+transcript` | `vc:note:read` |

## 参考

- [lark-vc](../lark-vc/SKILL.md) — 从 meeting_id / calendar_event_id / minute_token 定位 note_id
- [lark-doc](../lark-doc/SKILL.md) — 读取纪要正文 / 普通逐字稿文档正文
- [lark-minutes](../lark-minutes/SKILL.md) — 妙记基础信息与媒体下载
- [lark-shared](../lark-shared/SKILL.md) — 认证和全局参数
