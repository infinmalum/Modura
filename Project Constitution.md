# Project Constitution

你正在参与一个新的、独立设计的 Go + React 企业应用开发框架。

本项目不是 BladeX、SpringBlade、RuoYi、JeeSite 的语言移植版，也不追求逐文件、逐类、逐接口兼容。

BladeX / SpringBlade 源码仅作为成熟企业应用框架的需求、领域模型、权限模型、模块划分和历史设计经验参考。

核心目标是：

> 重新使用 Go、React 和现代工程实践设计企业应用框架，并系统性去除传统 Java/Spring 企业框架中因为历史、框架机制或过度抽象产生的复杂度。

项目采用 Apache License 2.0。

如果实际复制或修改第三方 Apache-2.0 代码，而不是仅参考设计思想，必须保留其许可证、版权声明和必要 NOTICE 信息。

---

# 一、项目定位

本项目定位为：

**Agent-native enterprise application framework for Go**

核心设计理念：

- Go-native
- Modular Monolith First
- Microservice Ready
- Explicit over Magical
- OpenAPI First
- Agent Native
- Minimal Ceremony
- Production Oriented

目标不是构建“Go 版若依”。

目标是重新回答：

> 如果今天从零设计一个面向 AI Coding Agent 和现代企业应用开发的基础框架，它应该是什么样？

---

# 二、核心技术栈

后端：

- Go
- Gin
- gRPC
- Protocol Buffers
- PostgreSQL
- Redis（仅在确有需求时）
- Casbin
- GORM / sqlx / sqlc，根据场景选择
- OpenTelemetry
- slog
- OpenAPI

前端：

- React
- TypeScript
- Vite
- React Router
- TanStack Query
- Ant Design
- 原生 fetch
- OpenAPI code generation
- Orval 或等价工具

AI / Agent：

- Codex 是主要开发 Agent
- 项目必须提供框架专用 Agent Skill / AGENTS.md
- 所有架构规范应尽量显式、稳定、可被 Agent 自动理解和检查

---

# 三、总体架构原则

默认采用：

Modular Monolith First

即：

cmd/
internal/
modules/
api/
proto/
admin/
skills/
docs/

业务首先组织为明确模块，而不是立即拆分为微服务。

示例：

modules/
identity/
organization/
authorization/
tenant/
audit/
system/
file/
workflow/

模块边界 != 部署边界。

一个模块默认可以运行在同一进程中。

只有出现明确理由时才拆为独立服务，例如：

- 独立扩缩容需求
- 独立资源消耗特征
- 强故障隔离需求
- 明确团队边界
- 独立生命周期
- 外部服务化价值

如果未来拆为服务，优先通过 gRPC 通信。

不要为了“微服务架构完整”而拆服务。

---

# 四、严禁机械复制 Java / Spring 架构

禁止把以下 Java/Spring 模式机械映射到 Go：

## 1. 无意义接口

禁止：

IUserService
UserServiceImpl

如果只有一个实现：

直接：

type UserService struct {}

只有存在真实多实现边界时才定义 interface，例如：

- LocalFileStore / S3FileStore
- LocalWorkflow / GRPCWorkflowClient
- Mock implementation
- 外部基础设施 Port

interface 应由消费者定义，而不是为了“分层完整”定义。

---

## 2. 禁止 Impl 后缀

禁止：

UserServiceImpl
UserRepositoryImpl
PermissionManagerImpl

---

## 3. 禁止 Java 风格继承体系

避免：

BaseService
BaseController
BaseEntity
AbstractManager
AbstractFactory
CommonBaseXXX

优先组合。

不要为了消除几行重复代码引入复杂抽象。

---

## 4. 禁止 DTO / BO / VO / PO 无限转换

只有边界语义真实不同才创建不同类型。

允许：

User
CreateUserRequest
UserResponse

禁止无意义：

UserEntity
UserPO
UserBO
UserDTO
UserVO
UserDO

在没有明确字段或语义差异时重复定义。

---

## 5. 禁止复制 Spring IoC 风格

使用显式构造函数注入。

例如：

repo := user.NewRepository(db)

service := user.NewService(
repo,
authz,
)

handler := user.NewHandler(service)

依赖关系必须容易通过代码阅读理解。

禁止隐藏依赖。

---

## 6. 禁止模仿 Spring AOP

优先：

- middleware
- decorator
- explicit function call
- event

不要为了实现日志、权限、数据范围等功能制造隐式 AOP 魔法。

---

# 五、错误处理原则

Go 的显式错误处理是设计的一部分。

允许并鼓励：

if err != nil {
return ...
}

优先使用：

- errors.New
- errors.Is
- errors.As
- fmt.Errorf("%w")

错误必须保留语义和上下文。

例如：

return fmt.Errorf("load user %d: %w", id, err)

HTTP / gRPC 错误转换必须发生在边界层。

Repository 不返回 HTTP status。

Domain/Application 不依赖 Gin。

禁止构建庞大的：

BaseException
BusinessException
ServiceException
GlobalExceptionFactory

式异常体系。

---

# 六、数据库原则

不要创建 Go 版 MyBatis。

不同场景允许不同数据访问方式：

简单 CRUD：

GORM 可以使用。

复杂查询：

优先考虑 sqlx / sqlc / 显式 SQL。

原则：

- SQL 应可理解
- 数据访问行为应显式
- 不引入无意义 repository ceremony
- transaction 边界必须清晰
- tenant / data scope 不允许靠不可见魔法偷偷注入

---

# 七、认证与授权

不要复制 Spring Security / Sa-Token 的结构。

认证 Authentication 和授权 Authorization 分离。

Authorization 以 Casbin 或类似策略引擎为核心。

概念优先：

subject
resource
action
tenant
scope

菜单不是权限本体。

菜单只是权限在 UI 层的一种投影。

推荐：

Authorization
Resource
Action
Policy

UI
Menu
Route
Button

API 与 UI 都引用统一 Permission / Resource 定义。

---

# 八、数据权限

禁止照搬：

@DataScope
AOP
ThreadLocal
SQL interceptor
隐式 SQL 拼接

优先显式模型。

例如：

scope, err := authz.Scope(
ctx,
actor,
ResourceAsset,
ActionRead,
)

items, err := repo.List(ctx, scope, filter)

Data Scope 可考虑：

- All
- Tenant
- Department
- DepartmentAndChildren
- Self
- Custom

具体实现应在设计阶段确定。

---

# 九、context.Context

context.Context 是请求生命周期、取消、deadline 和跨层 metadata 的标准机制。

必须向下传递。

禁止：

- 全局 request context
- ThreadLocal 风格实现
- 自定义隐式上下文容器

---

# 十、HTTP 与 gRPC

外部 Web API：

优先 Gin + HTTP/JSON。

服务间：

优先 gRPC。

浏览器不直接依赖内部 gRPC 服务。

典型：

React
-> Gin API / BFF
-> local module
-> gRPC service

如果模块尚无独立部署需要，应保持 local module。

---

# 十一、前端架构原则

前端目标：

简单、主流、强约束、Agent 友好。

不要创建复杂前端框架。

目录优先按 Feature，而不是按 Controller/Service/API 技术类型拆散：

src/
app/
features/
users/
roles/
tenants/
audit/
shared/
api/
generated/

普通 Server State：

TanStack Query。

URL State：

React Router。

局部 UI State：

useState。

表单：

Ant Design Form。

避免默认引入 Redux。

---

# 十二、禁止前端常见复杂化

禁止默认：

- Axios wrapper 多层封装
- Redux 管理 Server State
- useEffect 手工 fetch API
- 手写后端已经存在的 API 类型
- 重复定义 DTO
- 万能 utils 目录
- 自己重复实现 Ant Design 已经提供的基础组件
- 在组件内散落权限字符串
- 在多个地方重复处理 401/403/404

普通 REST API：

OpenAPI
-> generated TypeScript client
-> TanStack Query
-> React

HTTP transport：

原生 fetch。

AI / Agent streaming：

fetch + ReadableStream + SSE/NDJSON parser。

---

# 十三、OpenAPI First

Go 后端是 API contract 的主要来源。

OpenAPI 必须参与：

- API client generation
- TypeScript type generation
- 文档
- 测试

避免：

Go 定义一次
React 再手写一次

导致前后端类型漂移。

---

# 十四、Audit

所有重要写操作必须具有明确的审计设计。

审计至少考虑：

- actor
- tenant
- action
- resource
- resource id
- timestamp
- result
- trace id

不要默认使用 AOP + 注解实现审计。

优先显式事件或应用层行为。

---

# 十五、Observability

优先：

OpenTelemetry
slog

统一处理：

- trace id
- request id
- structured logging
- metrics
- gRPC trace
- HTTP trace

不要重新设计自己的 tracing 标准。

---

# 十六、AI Agent Native

本项目的一等用户之一是 Coding Agent。

因此：

- 模块结构必须规律
- 命名必须稳定
- 依赖必须显式
- 规范必须文档化
- 架构规则必须可检查
- generator 产生的代码必须可预测
- 避免大量反射和运行时魔法

Agent 应该能够：

1. 阅读一个模块理解全部重要依赖
2. 根据需求创建新模块
3. 正确注册权限
4. 创建数据库 migration
5. 创建 API
6. 更新 OpenAPI
7. 生成 React 页面
8. 添加测试
9. 自动运行测试和静态检查
10. 自己修复发现的问题

---

# 十七、BladeX / SpringBlade 参考原则

BladeX / SpringBlade 是重要设计参考，但禁止默认翻译代码。

分析 BladeX 模块时，必须先提取：

1. 它解决什么业务问题？
2. 核心实体是什么？
3. 数据关系是什么？
4. 权限规则是什么？
5. API surface 是什么？
6. 哪些跨模块依赖存在？
7. 哪些复杂度来自业务本身？
8. 哪些复杂度来自 Spring？
9. 哪些复杂度来自 MyBatis？
10. 哪些复杂度来自 Spring Cloud？
11. 哪些是历史兼容负担？
12. 如果今天使用 Go 从零设计，最小实现是什么？

之后才能提出 Go-native 实现。

不要逐类翻译。

不要保持 Java package hierarchy。

不要为了“和 BladeX 一致”保留不必要抽象。

---

# 十八、首阶段范围

优先完成企业应用核心：

1. Identity
   - User
   - Credential

2. Organization
   - Department
   - Position

3. Authorization
   - Role
   - Resource
   - Action
   - Policy
   - Data Scope

4. Tenant

5. System
   - Dictionary
   - Configuration

6. Audit

7. OpenAPI

8. React Admin Shell

暂不优先：

- 完整 BPM
- 复杂报表
- IoT
- 大数据
- 消息中间件
- Kubernetes Operator
- 分库分表
- 分布式事务
- 大规模微服务
- AI Gateway
- RAG 平台

除非明确收到任务，不得主动扩张 scope。

---

# 十九、每次开发前的决策规则

添加任何 abstraction、dependency 或 service 前必须回答：

1. 它解决了什么真实问题？
2. 标准库能否解决？
3. 现有组件能否解决？
4. 是否只有一个实现？
5. 是否真的需要 interface？
6. 是否真的需要独立 service？
7. 是否增加 Agent 理解成本？
8. 是否增加维护成本？
9. 删除它以后系统是否反而更清楚？

如果没有明确理由：

不要添加。

---

# 二十、最高原则

不要把 Java 屎山翻译成 Go 屎山。

不要为了“企业级”增加 ceremony。

不要为了“架构先进”增加微服务。

不要为了“可扩展”预测不存在的未来需求。

不要为了“复用”过早抽象。

优先：

简单
显式
稳定
可测试
可观察
可生成
可维护
Agent 友好

每次实现完成后，请主动检查：

> 如果一个熟悉 Go 但完全没见过这个项目的人阅读这段代码，他能否快速理解发生了什么？

如果答案是否定的，优先简化。
