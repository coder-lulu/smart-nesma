# 演示数据库

本目录提供与 README 中 17 张截图对应的“蜂巢工作室 · 智慧园区演示”数据，包含 1 个项目、2 个建设周期、2 个需求版本、13 条层级需求、1 条人工评估和 3 条知识条目，以及应用所需的菜单、角色与权限数据。评估为人工样例，没有实际 AI 计算结果的保证。

- `s_nesma.dump`：PostgreSQL custom-format 演示备份。
- `s_nesma.sql`：同一份演示库的可检索 SQL，便于查看数据和恢复。

两种格式任选其一。公开文件仅含演示数据，不包含原始业务备份或真实运行凭证；不要再把实际业务备份加入本目录。

## 恢复准备

使用 PostgreSQL **17** 与匹配版本的 **pgvector**。可以复用已有 Docker PostgreSQL 容器。自行创建独立应用角色和归其所有的空数据库，不覆盖已有业务库。由数据库管理员在该目标库中预先执行：

```sql
CREATE EXTENSION IF NOT EXISTS vector;
```

两种备份均已去除扩展创建和扩展注释语句。`globals.sql` 不发布，也不需要导入；它属于集群级对象备份，可能包含角色、密码哈希及权限。应用角色应在自己的环境中创建。

将下列占位符替换为实际容器名、应用角色和空目标数据库名。所用角色需要目标库的建表及写入权限；认证使用自己配置的凭证。

## 从 dump 恢复

```sh
docker cp database/s_nesma.dump <container>:/tmp/s_nesma.dump
docker exec <container> pg_restore --no-owner --no-privileges --exit-on-error --single-transaction -U <database-user> -d <database-name> /tmp/s_nesma.dump
```

## 从 SQL 恢复

SQL 已去除对象所有者与权限恢复语句，不创建数据库或集群角色：

```sh
docker cp database/s_nesma.sql <container>:/tmp/s_nesma.sql
docker exec <container> psql -X -v ON_ERROR_STOP=1 --single-transaction -U <database-user> -d <database-name> -f /tmp/s_nesma.sql
```

## 配置与登录

将 `server/config.example.yaml` 复制为本地 `server/config.yaml`，设置自己的 PostgreSQL 主机、端口、数据库名、账号和密码，再按主 README 启动应用。

演示账号为 `admin`，密码为 `123456`。恢复后修改此密码，并为 `jwt.signing-key` 设置独立随机密钥。检查登录、菜单、项目及需求列表。外部 AI 需要自行配置，演示库不包含服务密钥；演示评估数值不能直接作为正式计量结论。
