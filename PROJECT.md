# ballast-watch：船舶压载水处理监测系统

> 垂直领域：船舶压载水管理系统（BWTS）的处理过程监测与排放合规判定。
> 系统服务于船上机舱与环保值班人员，记录压载舱、处理单元、采样点和水质传感器的实时数据，评估处理周期是否满足排放条件，并保留可追溯的操作与告警记录。
> 业务只涉及处理过程、设备状态、采样与合规评估，不涉及订单、库存、账务、预约或权限后台。

## 1. 业务实体

| 实体 | 关键字段 | 业务含义 |
|---|---|---|
| Vessel | id, imo_number, name, flag, ballast_mode, status | 船舶与当前压载模式 |
| BallastTank | id, vessel_id, code, capacity_m3, salinity_band, status | 船上的压载舱 |
| TreatmentUnit | id, vessel_id, serial, method(uv/electrolysis), status, last_service_at | 紫外或电解处理单元 |
| SamplingPoint | id, tank_id, code, parameter, min_value, max_value, alarm_duration_sec | 采样点和参数阈值 |
| Sensor | id, point_id, serial, parameter, status, calibration_due_at | 水质/流量/设备传感器 |
| WaterReading | id, point_id, sensor_id, value, measured_at, raw | 采集读数，SQLite 持久化 |
| TreatmentCycle | id, vessel_id, tank_id, started_at, ended_at, mode, status | 一次压载水处理周期 |
| DischargeEvent | id, vessel_id, tank_id, cycle_id, started_at, ended_at, status | 排放开始、暂停、完成状态 |
| ComplianceRule | id, vessel_id, parameter, comparator, threshold, duration_sec, level | 合规与告警规则 |
| ComplianceAlert | id, rule_id, vessel_id, point_id, level, status, opened_at | 超标或处理故障告警 |
| Calibration | id, sensor_id, performed_at, due_at, result, technician | 传感器校准记录 |
| ComplianceStatus | id, vessel_id, cycle_id, state, reason, changed_at | 合规状态机历史 |

## 2. 业务模块

1. **船舶与压载舱**：登记船舶、压载舱和处理单元，维护机舱设备状态。
2. **采样采集**：网关批量接收流量、盐度、浊度、余氯、UV 强度等读数，后台 worker 写入 SQLite。
3. **处理周期**：开始、暂停、恢复、完成处理周期；完成前需满足处理单元在线与水质阈值要求。
4. **排放控制**：排放事件状态机 `prepared -> running -> paused -> completed/rejected`，排放前检查最近窗口合规率。
5. **合规引擎**：阈值、持续时长、设备在线状态、校准有效期共同决定告警与合规状态。
6. **校准管理**：校准失败的传感器不可用于排放判定；校准成功可恢复 active。
7. **报告与导出**：按船舶/压载舱生成采样趋势、处理周期摘要和监管 CSV。
8. **实时页面**：内嵌 web 页面展示船舶、处理单元、采样读数、告警和当前合规状态。

## 3. HTTP 接口

| 方法 | 路径 |
|---|---|
| POST/GET | `/api/v1/vessels` |
| GET | `/api/v1/vessels/{id}` |
| POST/GET | `/api/v1/tanks` |
| GET | `/api/v1/tanks/{id}/status` |
| POST/GET | `/api/v1/treatment-units` |
| POST/GET | `/api/v1/sampling-points` |
| POST/GET | `/api/v1/sensors` |
| POST/GET | `/api/v1/readings` |
| GET | `/api/v1/readings/query` |
| GET | `/api/v1/readings/stats` |
| POST | `/api/v1/cycles` |
| PUT | `/api/v1/cycles/{id}/pause` |
| PUT | `/api/v1/cycles/{id}/resume` |
| PUT | `/api/v1/cycles/{id}/complete` |
| GET | `/api/v1/cycles` |
| POST | `/api/v1/discharges` |
| PUT | `/api/v1/discharges/{id}/pause` |
| PUT | `/api/v1/discharges/{id}/complete` |
| GET | `/api/v1/discharges` |
| POST/GET | `/api/v1/rules` |
| GET | `/api/v1/alerts` |
| PUT | `/api/v1/alerts/{id}/ack` |
| PUT | `/api/v1/alerts/{id}/resolve` |
| POST/GET | `/api/v1/sensors/{id}/calibrations` |
| GET | `/api/v1/reports/trend` |
| GET | `/api/v1/reports/summary` |
| GET | `/api/v1/reports/export` |
| GET | `/api/v1/dashboard` |
| GET | `/healthz` |

## 4. 技术与规模约束

- Go 1.22，单体单进程，标准库 `net/http`。
- `modernc.org/sqlite` 文件模式，数据库默认 `ballast.db`，禁止 `:memory:`。
- 直接第三方依赖不超过 10 个；只使用 SQLite 驱动。
- 内部包不超过 12 个：config / model / store / service / handler / collector / alerter / report / util。
- 采集器使用 goroutine/channel；告警评估采用 worker；所有关键数据落盘。
- 非测试 Go 文件至少 50 个、有效 Go 代码至少 5000 行。
- web 页面依附 Go 后端，代码不计入 Go 规模统计。

## 5. 缺陷设计素材

- 读数去重键为 `(sampling_point_id, measured_at 秒)`，批量上报必须整批校验。
- 排放前要检查最近处理窗口、传感器校准有效期、处理单元状态和水质阈值。
- 合规状态流转为 `idle -> treating -> compliant -> ready_to_discharge -> discharging -> completed/rejected`。
- 取消处理周期或排放时，所有循环都必须停止继续写库。
- 告警需要持续时长判定，同一规则/采样点的未决告警不能重复创建。
- 传感器校准失败必须将传感器标为 fault，任何 fault 传感器读数不能参与放行判定。
