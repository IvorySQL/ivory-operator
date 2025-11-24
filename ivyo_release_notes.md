# IvorySQL Operator Release Notes

## IvorySQL Operator 1.0
- 实例管理（订阅、连接、重启、关闭、自定义配置、调整CPU mem和磁盘、删除）
- 用户/数据库管理（创建用户、权限控制、修改密码、删除用户、删除数据库）
- 备份恢复（备份、按时间点恢复至新/旧实例）
- 数据库管理工具pgadmin
- 高可用（故障恢复、主备切换）
- 数据库监控

## IvorySQL Operator 1.1
- 实例管理（自定义tls、初始化sql）
- 连接池（添加连接池、连接到连接池、配置TLS）
- 备份恢复（设置自动备份计划和保留策略、备份至云存储、恢复单个数据库、加密备份、Ipv6支持、备用实例）
- 数据库插件支持（Postgis、TimescaleDB、pg_cron、pgAudit、pgBouncer、wal2json）
- 高可用（同步复制、亲和性、Pod拓扑分布约束）
- 其他功能（暂停operator的调协、轮换 TLS 证书）

## IvorySQL Operator 2.0 
- 数据库版本支持**ivorysql 3.0**
- 组件版本升级
	- pgbackrest 2.47
	- pgadmin4 8.0
	- pgbouncer 1.21
	- pgexporter 0.15
	- postgis 3.4
	- pgaudit 16
	- pg_cron 1.6.2
	- pgnodemx 1.6
	- wal2json 2.5
## IvorySQL Operator 4.0
- 数据库版本支持**ivorysql 4.5**
- 组件版本升级
    - pgBackrest 4.5
    - PgBouncer 1.23.0
    - Patroni 4.0.4
    - pgAdmin4 8.14.0
    - pgExporter 0.17.0
- 插件版本升级
	- PostGIS 3.4.0
	- pgaudit 17.0
	- pg_cron 1.6.5
	- timescaledb 2.17.2
	- wal2json 2.6
	- pgnodemx 1.7
## IvorySQL Operator 5.0
- 数据库版本支持**ivorysql 5.0**
- 组件版本升级
    - pgBackrest 2.56
    - PgBouncer 1.23.0
    - Patroni 4.0.7
    - pgAdmin4 9.9
    - pgExporter 0.17.0
- 插件版本升级
	- PostGIS 3.5.4
	- pgaudit 18.0
	- pg_cron 1.6.7
	- wal2json 2.6
	- pgnodemx 1.7
