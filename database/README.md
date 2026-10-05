# 数据库准备

Smart NESMA 使用 PostgreSQL，向量知识库功能还需要 pgvector 扩展。连接信息在 `server/config.example.yaml` 中提供结构示例；复制为 `server/config.yaml`，并设置自己的数据库主机、端口、数据库名称、账号和密码。

仓库不公开数据库备份。`*.dump` 可能包含项目、需求、用户、会话及业务数据，已列入忽略规则。若采用备份恢复方式部署，请自行提供经过授权的备份文件，恢复到专用数据库。现有应用的备份可能同时包含菜单、角色和权限等初始化数据；仅启动后端自动建表不保证业务环境已经完整初始化。

下面示例复用已有 Docker PostgreSQL 容器，不创建或覆盖现有容器。将 `<container>`、`<database-user>` 和 `<database-name>` 替换为自己的值，并先创建一个空的专用数据库：

```sh
docker cp database/s_nesma.dump <container>:/tmp/s_nesma.dump
docker exec <container> pg_restore --no-owner --no-privileges -U <database-user> -d <database-name> /tmp/s_nesma.dump
```

恢复前确保容器中的 PostgreSQL 已安装 pgvector，并由有权限的数据库管理员在目标数据库中执行：

```sql
CREATE EXTENSION IF NOT EXISTS vector;
```

`globals.sql` 属于 PostgreSQL 集群级全局对象备份，可能包含角色、密码哈希及权限，不能作为公开示例提交。不要将其导入共享数据库集群以覆盖现有账号或权限。请为应用单独创建数据库角色并按需要授予权限。

恢复后检查数据库连接及应用登录、菜单、项目列表；正式使用前更改备份中的默认账号密码和示例配置中的 JWT 签名密钥。
