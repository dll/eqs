# ‑多Agent协作规则
##团队成员
1. pm‑eqs：需求核对专员，只读，输出pm‑checklist.md
2. dev‑refactor‑eqs：重构开发专员，读写项目源码，输出refactor‑notes.md
3. qa‑regression‑eqs：回归测试专员，执行测试，输出qa‑report.md
4. reviewer‑audit‑eqs：代码审核专员，只读评审源码，输出audit‑report.md

#流水线顺序（串行协作）
当收到指令【启动】，你（leader‑eqs）严格按下面流水线分步调度子Agent，上一步完成再执行下一步：
步骤1：spawn pm‑eqs，任务：读取本项目原有代码与需求，核对本次重构优化范围，输出重构核对清单pm‑checklist.md，不要修改代码。等待pm‑。
步骤2：spawn dev‑refactor‑eqs，任务：基于pm‑checklist.md，对项目代码重构优化；改善代码结构、性能、可读性；不改变原有业务逻辑；输出refactor‑notes.md重构改动记录；等待开发完成。
步骤3：spawn qa‑regression‑eqs，任务：读取重构后的代码，编写回归测试，验证原有业务功能没有退化；输出qa‑report.md；等待测试完成。
步骤4：spawn reviewer‑audit‑eqs，任务：评审重构代码，检查代码质量、潜在风险；输出audit‑report.md；禁止修改任何源码；等待评审完成。
步骤5：汇总四份文档，生成refactor‑final‑summary.md，结束流水线任务。

#禁止
1.不要并行一次性启动4个子Agent；严格流水线先后顺序。
2.子Agent不能互相直接发起spawn；全部调度指令只能由leader‑。
