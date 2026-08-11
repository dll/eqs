#!/bin/bash
# ===== EQS MySQL 备份脚本 =====
# P1-14：每日全量备份，保留 7 天，可选异地推送（COS/其他机器）
# 用法：bash backup.sh  （配合 cron 每日执行）
# 配置：BACKUP_DIR / DB_USER / DB_PASSWORD / DB_NAME / RETENTION_DAYS

# pipefail：否则 mysqldump 失败仍会写出"有效"的空 gzip
set -eo pipefail

# ===== 配置 =====
# 凭据来源（二选一，不内置默认值）：
#   1) 环境变量 DB_PASSWORD（经 MYSQL_PWD 传递，不落进程列表）
#   2) MySQL 凭据文件 MYSQL_DEFAULTS_FILE（默认 /root/.my.cnf，权限 600，cron 场景推荐）
BACKUP_DIR="${BACKUP_DIR:-/opt/eqs/backup/mysql}"
DB_USER="${DB_USER:-eqs}"
MYSQL_DEFAULTS_FILE="${MYSQL_DEFAULTS_FILE:-/root/.my.cnf}"
DB_NAME="${DB_NAME:-eqs}"
RETENTION_DAYS="${RETENTION_DAYS:-7}"
DATE=$(date +%Y%m%d_%H%M%S)

mkdir -p "$BACKUP_DIR"
BACKUP_FILE="$BACKUP_DIR/${DB_NAME}_${DATE}.sql.gz"

echo "=== EQS MySQL 备份开始 ==="

# 1. 全量备份（mysqldump，压缩；--no-tablespaces 避免 PROCESS 权限告警）
#    优先用环境变量（MYSQL_PWD 避免口令出现在进程列表），否则回退凭据文件
if [ -n "${DB_PASSWORD:-}" ]; then
  MYSQL_PWD="$DB_PASSWORD" mysqldump -u"$DB_USER" --single-transaction --no-tablespaces --routines --triggers "$DB_NAME" | gzip > "$BACKUP_FILE"
elif [ -f "$MYSQL_DEFAULTS_FILE" ]; then
  mysqldump --defaults-extra-file="$MYSQL_DEFAULTS_FILE" --single-transaction --no-tablespaces --routines --triggers "$DB_NAME" | gzip > "$BACKUP_FILE"
else
  echo "错误: 未配置数据库凭据（请设置 DB_PASSWORD 或提供 $MYSQL_DEFAULTS_FILE）" >&2
  exit 1
fi
echo "备份完成: $BACKUP_FILE ($(du -h "$BACKUP_FILE" | cut -f1))"

# 2. 备份含完整业务数据，收紧为仅 root 可读（防止其他用户读取）
chmod 600 "$BACKUP_FILE"

# 3. 校验备份文件完整性
gzip -t "$BACKUP_FILE"

# 3. 清理过期备份（-print 只报告真正删掉的文件）
find "$BACKUP_DIR" -name "${DB_NAME}_*.sql.gz" -mtime +"$RETENTION_DAYS" -print -delete

# 4. 可选异地推送（配置 REMOTE_HOST/REMOTE_DIR 时启用 rsync）
if [ -n "$REMOTE_HOST" ] && [ -n "$REMOTE_DIR" ]; then
  rsync -avz "$BACKUP_FILE" "${REMOTE_HOST}:${REMOTE_DIR}/" 2>/dev/null && echo "已同步到异地 $REMOTE_HOST" || echo "警告: 异地同步失败"
fi

echo "=== 备份完成 ==="
