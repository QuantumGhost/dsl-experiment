# dsl-exp

我现在需要为 Dify 设计一个新的 DSL 语法，这个 DSL 无需考虑历史兼容性，但需要考虑如何让这个 DSL 最为自然，这个 DSL 需要支持下面的特性：

1. 支持 Memory （Global Memory 或者 Context Memory）
2. 支持多节点并行执行，并行执行的节点需要共享 memory。
3. 思考如何自然地用 DSL 表达 Workflow 的顺序关系
4. 思考如何增强 DSL 的能力，让它可以作为一个基础被用来实现 AI Agent。

在设计完这个 DSL 和上面 1 ，2 的实现方式之后，我需要你使用 Go 实现一个 minimal runtime，这个 minimal runtime 应该可以执行一个 DSL 输入。

为简化实现难度，这里的 minimal runtime 只需要支持 OpenAI 兼容的 LLM API 即可（你可以就以 OpenAI 的 API 作为实现基础）。

关于 DSL 的节点的思考：

- 一个最小的 DSL 应该包含哪些节点？
- 这些节点应该包含怎样的配置
