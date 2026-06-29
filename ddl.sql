-- `create database`

CREATE DATABASE IF NOT EXISTS tbos DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;

USE tbos;

-- `tbos`.t_alarm_active definition

CREATE TABLE IF NOT EXISTS `t_alarm_active`
(
    `id`             bigint(10)   NOT NULL AUTO_INCREMENT COMMENT '告警自增ID',
    `alarm_id`       bigint(20)   NOT NULL COMMENT '告警唯一ID',
    `occur_time`     datetime     NOT NULL COMMENT '告警发生时间',
    `level`          varchar(255) NOT NULL COMMENT '告警级别',
    `mozu_id`        int(11)      NOT NULL COMMENT '模组ID',
    `rid`            bigint(20)   NOT NULL COMMENT '告警策略ID',
    `alarm_name`     varchar(255) NOT NULL COMMENT '告警类型',
    `content`        varchar(255) NOT NULL COMMENT '告警内容',
    `fingerprint`    varchar(100) NOT NULL COMMENT '告警指纹',
    `analyze_result` longtext     NOT NULL COMMENT '告警分析结果',
    `device_gid`     varchar(100) NOT NULL COMMENT '告警设备GID',
    `device_number`  varchar(255) NOT NULL COMMENT '告警设备编号',
    `device_type_zh` varchar(255) DEFAULT NULL COMMENT '设备类型中文',
    `box_name`       varchar(100) NOT NULL COMMENT '方仓名，搜索用',
    `room_name`      varchar(100) NOT NULL COMMENT '房间名，搜索用',
    `create_at`      datetime     NOT NULL COMMENT '创建时间',
    `update_time`    datetime     NOT NULL COMMENT '告警最新检出时间',
    `status`         int(9)       NOT NULL DEFAULT '0' COMMENT '告警状态： 0 正常 1 挂起',
    `event_status`   int(9)       NOT NULL DEFAULT '0' COMMENT '告警事件状态 1.未转单 2已转单 3 结单',
    `op_user`        varchar(100) DEFAULT NULL COMMENT '最近一次操作人',
    `op_reason`      varchar(255) DEFAULT NULL COMMENT '最近一次操作原因',
    PRIMARY KEY (`id`),
    UNIQUE KEY `alarm_id` (`alarm_id`) USING BTREE,
    UNIQUE KEY `fingerprint` (`fingerprint`) USING BTREE,
    KEY              `mozu_level_IDX` (`mozu_id`, `status`, `level`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8 COMMENT ='活动告警库';


-- `tbos`.t_alarm_history definition

CREATE TABLE IF NOT EXISTS `t_alarm_history`
(
    `id`                     bigint(10)   NOT NULL AUTO_INCREMENT COMMENT '告警自增ID',
    `alarm_id`               bigint(20)   NOT NULL COMMENT '告警ID',
    `level`                  varchar(255) NOT NULL COMMENT '告警级别',
    `occur_time`             datetime     NOT NULL COMMENT '告警发生时间',
    `rid`                    bigint(20)   NOT NULL COMMENT '告警策略ID',
    `mozu_id`                int(11)      NOT NULL COMMENT '模组ID',
    `alarm_name`             varchar(255) NOT NULL COMMENT '告警类型',
    `content`                varchar(500) NOT NULL COMMENT '告警内容',
    `analyze_result`         longtext     NOT NULL COMMENT '告警分析结果',
    `fingerprint`            varchar(100) NOT NULL COMMENT '告警指纹',
    `device_gid`             varchar(100) NOT NULL COMMENT '告警设备GID',
    `device_number`          varchar(255) NOT NULL COMMENT '告警设备编号',
    `device_type_zh`         varchar(255) DEFAULT NULL COMMENT '设备类型中文',
    `box_name`               varchar(100) NOT NULL COMMENT '方仓名，搜索用',
    `room_name`              varchar(100) NOT NULL COMMENT '房间名，搜索用',
    `restore_time`           datetime     NOT NULL COMMENT '告警恢复时间',
    `restore_analyze_result` longtext     NOT NULL COMMENT '恢复分析结果',
    `create_at`              datetime     NOT NULL COMMENT '创建时间',
    `active_create_at`       datetime     NOT NULL COMMENT '活动告警创建时间',
    `op_user`                varchar(100) DEFAULT NULL COMMENT '最近一次操作人',
    `op_reason`              varchar(255) DEFAULT NULL COMMENT '最近一次操作原因',
    PRIMARY KEY (`id`),
    UNIQUE KEY `alarm_id` (`alarm_id`) USING BTREE,
    KEY                      `fingerprint` (`fingerprint`) USING BTREE,
    KEY                      `mozu_occur_time_IDX` (`mozu_id`, `occur_time`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8;


-- `tbos`.t_alarm_mworker definition

CREATE TABLE IF NOT EXISTS `t_alarm_mworker`
(
    `id`            bigint(20)   NOT NULL AUTO_INCREMENT COMMENT '主键id',
    `set_id`        int(9)       NOT NULL COMMENT '片区Id',
    `worker_id`     int(9)       NOT NULL COMMENT '分布式唯一workerId',
    `occupy_status` int(9)       NOT NULL COMMENT 'id占用状态 0: 未占用 1:已占用',
    `uuid`          varchar(255) NOT NULL COMMENT '占用的Pod的唯一标识',
    `pod_ip`        varchar(100) NOT NULL COMMENT '占用的Pod的ip地址',
    `heartbeat`     datetime     NOT NULL COMMENT '最近一次注册心跳的时间',
    `create_at`     datetime     NOT NULL COMMENT '创建时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `WorkerIdx` (`worker_id`) USING BTREE,
    UNIQUE KEY `UidIdx` (`uuid`) USING BTREE
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8 COMMENT ='告警管理Pod分布式workerID获取';


-- `tbos`.t_alarm_strategy definition

CREATE TABLE IF NOT EXISTS `t_alarm_strategy`
(
    `id`                     bigint(20)   NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `device_gid`             varchar(127) NOT NULL DEFAULT '' COMMENT '设备GID',
    `rid`                    bigint(20)   NOT NULL DEFAULT '0' COMMENT '策略ID',
    `rid_version`            varchar(127) NOT NULL DEFAULT '' COMMENT '策略版本',
    `rid_type`               tinyint(4)   NOT NULL DEFAULT '0' COMMENT '策略类型,0:实时,1:延时',
    `mozu_id`                int(11)      NOT NULL DEFAULT '0' COMMENT '所属模组ID',
    `alarm_name`             varchar(127) NOT NULL DEFAULT '' COMMENT '告警名称',
    `alarm_expression`       text         NOT NULL COMMENT '告警表达式',
    `alarm_expression_str`   text         NOT NULL COMMENT '告警表达式(中文)',
    `restore_expression`     text         NOT NULL COMMENT '恢复表达式',
    `restore_expression_str` text         NOT NULL COMMENT '恢复表达式(中文)',
    `expression_map`         text         NOT NULL COMMENT '表达式映射',
    `alarm_level`            varchar(8)   NOT NULL DEFAULT '' COMMENT '告警级别',
    `content_template`       text         NOT NULL COMMENT '告警内容模版',
    `owner`                  varchar(32)  NOT NULL DEFAULT '' COMMENT '告警负责人',
    `compute_cost`           int(11)      NOT NULL DEFAULT '0' COMMENT '计算复杂度',
    `create_at`              datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '记录创建时间',
    `update_at`              datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '记录上次更新时间',
    PRIMARY KEY (`id`),
    KEY                      `t_alarm_strategy_device_gid_IDX` (`device_gid`) USING BTREE,
    KEY                      `t_alarm_strategy_rid_IDX` (`rid`, `rid_version`) USING BTREE,
    KEY                      `mozu_rid_type_IDX` (`mozu_id`, `rid_type`)
) ENGINE = InnoDB
  AUTO_INCREMENT = 5352338
  DEFAULT CHARSET = utf8 COMMENT ='告警策略配置信息，每条告警挂载在具体的设备上';


-- `tbos`.t_collector_device definition

CREATE TABLE IF NOT EXISTS `t_collector_device`
(
    `id`                   bigint(20)   NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `device_gid`           varchar(255) NOT NULL DEFAULT '' COMMENT '设备GID',
    `device_number`        varchar(255) NOT NULL DEFAULT '' COMMENT '设备编码',
    `device_sn`            varchar(127) NOT NULL DEFAULT '' COMMENT '设备SN',
    `device_code`          varchar(127) NOT NULL DEFAULT '' COMMENT '设备代号',
    `device_name`          varchar(127) NOT NULL DEFAULT '' COMMENT '设备名称',
    `device_type_en`       varchar(127) NOT NULL DEFAULT '' COMMENT '采集设备类型英文',
    `device_type_zh`       varchar(127) NOT NULL DEFAULT '' COMMENT '采集设备类型中文',
    `collector_type`       tinyint(4)   NOT NULL DEFAULT '0' COMMENT '采集类型,1:Tbox,2: Tbox下子设备，3：厂商采集器，4：厂商采集器子设备',
    `channel_type`         varchar(16)           DEFAULT '' COMMENT '通道类型',
    `channel_id`           varchar(127)          DEFAULT '' COMMENT '通道地址',
    `channel_link`         text COMMENT '通道详细信息',
    `active_status`        tinyint(4)            DEFAULT '0' COMMENT '激活状态',
    `template_name`        varchar(127)          DEFAULT '' COMMENT '模版名称',
    `template_info`        varchar(255) NOT NULL DEFAULT '' COMMENT '模版信息',
    `parent_device_number` varchar(255)          DEFAULT '' COMMENT '父级设备编号',
    `mozu_id`              int(11)      NOT NULL DEFAULT '0' COMMENT '所属模组ID',
    `extend`               text COMMENT '扩展字段',
    `create_at`            datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '记录创建时间',
    `update_at`            datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '记录上次更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `t_collector_device_unique` (`device_gid`, `mozu_id`),
    KEY                    `t_collect_strategy_device_number_IDX` (`device_number`) USING BTREE,
    KEY                    `t_collect_strategy_template_name_IDX` (`template_name`) USING BTREE,
    KEY                    `t_collector_device_parent_device_number_IDX` (`parent_device_number`) USING BTREE
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8 COMMENT ='所有采集设备对应的采集策略';


-- `tbos`.t_collector_template definition

CREATE TABLE IF NOT EXISTS `t_collector_template`
(
    `id`               bigint(20)                                                   NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `template_name`    varchar(127) NOT NULL DEFAULT '' COMMENT '模版名称',
    `mozu_id`          int(11)                                                      NOT NULL DEFAULT '0' COMMENT '所属模组ID',
    `device_type_en`   varchar(127) NOT NULL DEFAULT '' COMMENT '设备类型英文',
    `device_type_zh`   varchar(127)          DEFAULT '' COMMENT '设备类型中文',
    `manufacturer`     varchar(127) NOT NULL DEFAULT '' COMMENT '设备制造商',
    `device_model_en`  varchar(127) NOT NULL DEFAULT '' COMMENT '设备型号',
    `protocol_type`    varchar(32)  NOT NULL DEFAULT '' COMMENT '协议类型',
    `protocol_version` varchar(32)  NOT NULL DEFAULT '' COMMENT '协议版本',
    `protocol_extend`  text COMMENT '协议扩展信息',
    `create_at`        datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '记录创建时间',
    `update_at`        datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '记录上次更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `t_collector_template_unique` (`template_name`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8 COMMENT ='设备采集模版定义';


-- `tbos`.t_collector_template_point definition

CREATE TABLE IF NOT EXISTS `t_collector_template_point`
(
    `id`             bigint(20)    NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `template_name`  varchar(127)  NOT NULL DEFAULT '' COMMENT '模版名称',
    `mozu_id`        int(11)       NOT NULL DEFAULT '0' COMMENT '所属模组ID',
    `sub_device`     varchar(127)  NOT NULL DEFAULT '' COMMENT '子设备名称',
    `point_name_en`  varchar(255)  NOT NULL DEFAULT '' COMMENT '测点名称英文',
    `point_name_zh`  varchar(255)  NOT NULL DEFAULT '' COMMENT '测点名称中文',
    `point_type`     varchar(16)   NOT NULL DEFAULT '' COMMENT '测点类型',
    `point_rw`       varchar(16)   NOT NULL DEFAULT '' COMMENT '测点读写分类',
    `point_standard` tinyint(4)    NOT NULL DEFAULT '0' COMMENT '是否标准测点',
    `delta_def`      varchar(1024) NOT NULL DEFAULT '' COMMENT '变化定义规则',
    `verify_def`     varchar(1024) NOT NULL DEFAULT '' COMMENT '校验规则',
    `exp_def`        text COMMENT '表达式定义规则',
    `prot_def`       text COMMENT '协议定义规则',
    `val_def`        varchar(1024) NOT NULL DEFAULT '' COMMENT '值定义规则',
    `simulator`      varchar(255)  NOT NULL DEFAULT '' COMMENT '模拟定义规则',
    `create_at`      datetime      NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '记录创建时间',
    `update_at`      datetime      NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '记录上次更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `t_collector_template_point_unique` (`template_name`, `sub_device`, `point_name_en`),
    KEY              `t_collector_template_point_template_name_IDX` (`template_name`) USING BTREE
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8 COMMENT ='设备采集模版中的测点定义';


-- `tbos`.t_device_entity definition

CREATE TABLE IF NOT EXISTS `t_device_entity`
(
    `id`                         bigint(20)   NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `device_gid`                 varchar(127) NOT NULL DEFAULT '' COMMENT '设备GID',
    `device_number`              varchar(255) NOT NULL DEFAULT '' COMMENT '设备编码',
    `device_number_route`        varchar(255) NOT NULL DEFAULT '' COMMENT '路由编码',
    `device_number_show`         varchar(255) NOT NULL DEFAULT '' COMMENT '展示编码',
    `device_name`                varchar(127) NOT NULL DEFAULT '' COMMENT '设备名称',
    `mozu_id`                    int(11)      NOT NULL DEFAULT '0' COMMENT '所属模组ID',
    `mozu_name`                  varchar(64)  NOT NULL DEFAULT '' COMMENT '模组名称',
    `idc_area`                   varchar(64)  NOT NULL DEFAULT '' COMMENT '机房区域',
    `func_room`                  varchar(64)  NOT NULL DEFAULT '' COMMENT '方仓/功能间',
    `parent_device_number`       varchar(255) NOT NULL DEFAULT '' COMMENT '父级设备编码',
    `application_type_en`        varchar(127) NOT NULL DEFAULT '' COMMENT '应用类型英文',
    `application_type_zh`        varchar(127) NOT NULL DEFAULT '' COMMENT '应用类型中文',
    `belong_application_type_en` varchar(127) NOT NULL DEFAULT '' COMMENT '所属应用类型',
    `create_at`                  datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '记录创建时间',
    `update_at`                  datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '记录上次更新时间',
    `device_type_en`             varchar(127) NOT NULL DEFAULT '' COMMENT '设备种类英文',
    `device_type_zh`             varchar(127) NOT NULL DEFAULT '' COMMENT '设备种类中文',
    PRIMARY KEY (`id`),
    UNIQUE KEY `t_device_entity_unique` (`device_gid`, `mozu_id`),
    KEY                          `t_device_entity_parent_device_gid_IDX` (`parent_device_number`) USING BTREE,
    KEY                          `t_device_entity_device_number_IDX` (`device_number`) USING BTREE
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8 COMMENT ='设备实体信息表';


-- `tbos`.t_device_point definition

CREATE TABLE IF NOT EXISTS `t_device_point`
(
    `id`                bigint(20)   NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `device_gid`        varchar(255) NOT NULL DEFAULT '' COMMENT '设备GID',
    `device_number`     varchar(255) NOT NULL DEFAULT '' COMMENT '设备编码',
    `belong_collector`  varchar(255) NOT NULL DEFAULT '' COMMENT '归属采集器',
    `point_name_en`     varchar(255) NOT NULL DEFAULT '' COMMENT '测点名称英文',
    `point_name_zh`     varchar(255) NOT NULL DEFAULT '' COMMENT '测点名称中文',
    `point_key`         varchar(255) NOT NULL DEFAULT '' COMMENT '测点标识',
    `point_type`        tinyint(4) NOT NULL DEFAULT '0' COMMENT '测点类型',
    `point_rw`          varchar(16)  NOT NULL DEFAULT '' COMMENT '测点读写分类',
    `point_level`       varchar(16)  NOT NULL DEFAULT '' COMMENT '测点级别',
    `expression`        text COMMENT '测点表达式',
    `expression_map`    text COMMENT '测点映射(设备GID)',
    `expression_map_zh` text COMMENT '测点映射(设备编号)',
    `value_type`        varchar(16)  NOT NULL DEFAULT '' COMMENT '测点值类型',
    `value_valid_range` varchar(255) NOT NULL DEFAULT '' COMMENT '测点值有效范围',
    `value_unit`        varchar(32)  NOT NULL DEFAULT '' COMMENT '测点值单位',
    `value_precision`   varchar(16)  NOT NULL DEFAULT '' COMMENT '测点值精度',
    `value_enum`        varchar(255) NOT NULL DEFAULT '' COMMENT '值枚举映射',
    `mozu_id`           int(11)      NOT NULL DEFAULT '0' COMMENT '所属模组ID',
    `create_at`         datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '记录创建时间',
    `update_at`         datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '记录上次更新时间',
    `point_category`    tinyint(4)   NOT NULL DEFAULT '0' COMMENT '测点类型: 0:未定义,1:全采集,2:标准+采集,3:全标准',
    PRIMARY KEY (`id`),
    UNIQUE KEY `t_device_point_unique` (`device_gid`, `point_name_en`),
    KEY                 `t_device_point_device_number_IDX` (`device_number`) USING BTREE,
    KEY                 `t_device_point_point_key_IDX` (`point_key`) USING BTREE,
    KEY                 `t_device_point_belong_collector_IDX` (`belong_collector`) USING BTREE
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8 COMMENT ='设备测点信息，每个设备上的测点列表及测点计算方案';


-- `tbos`.t_mozu_info definition

CREATE TABLE IF NOT EXISTS `t_mozu_info`
(
    `id`                 int(11)     NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `mozu_id`            int(11)     NOT NULL COMMENT '模组ID',
    `mozu_name`          varchar(32) NOT NULL DEFAULT '' COMMENT '模组名称',
    `mozu_code`          varchar(32) NOT NULL DEFAULT '' COMMENT '模组编码',
    `mozu_type`          int(11)     NOT NULL DEFAULT '0' COMMENT '模组类型',
    `belong_building`    varchar(32) NOT NULL DEFAULT '' COMMENT '所属楼栋',
    `belong_campus`      varchar(32) NOT NULL DEFAULT '' COMMENT '所属园区',
    `belong_campus_code` varchar(32) NOT NULL DEFAULT '' COMMENT '所属园区编码',
    `publish_version`    varchar(32) NOT NULL DEFAULT '' COMMENT '下发版本',
    `alarm_version`      varchar(32) NOT NULL DEFAULT '' COMMENT '下发版本',
    `create_at`          datetime    NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '记录创建时间',
    `update_at`          datetime    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '记录上次更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `t_mozu_info_unique` (`mozu_id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4 COMMENT ='模组信息表';