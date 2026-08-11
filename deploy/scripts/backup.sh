#!/bin/bash
# ===== EQS MySQL 备份脚本 =====
# P1-14：每日全量备份，保留 7 天，可选异地推送（COS/其他机器）
# 用法：bash backup.sh  （配合 cron 每日执行）
# 配置：BACKUP_DIR / DB_USER / DB_PASSWORD / DB_NAME / RETENTION_DAYS

set -e

# ===== 配置 =====
BACKUP_DIR="${BACKUP_DIR:-/opt/eqs/backup/mysql}"
DB_USER="${DB_USER:-eqs}"
DB_PASSWORD="${DB_PASSWORD:-EQS_DB_Pass_2026!}"
DB_NAME="${DB_NAME:-eqs}"
RETENTION_DAYS="${RETENTION_DAYS:-7}"
DATE=$(date +%Y%m%d_%H%M%S)

mkdir -p "$BACKUP_DIR"
BACKUP_FILE="$BACKUP_DIR/${DB_NAME}_${DATE}.sql.gz"

echo "=== EQS MySQL 备份开始 ==="

# 1. 全量备份（mysqldump，压缩；--no-tablespaces 避免 PROCESS 权限告警）
mysqldump -u"$DB_USER" -p"$DB_PASSWORD" --single-transaction --no-tablespaces --routines --triggers "$DB_NAME" | gzip > "$BACKUP_FILE"
echo "备份完成: $BACKUP_FILE ($(du -h "$BACKUP_FILE" | cut -f1))"

# 2. 校验备份文件完整性
gzip -t "$BACKUP_FILE" && echo "备份校验通过"

# 3. 清理过期备份
find "$BACKUP_DIR" -name "${DB_NAME}_*.sql.gz" -mtime +"$RETENTION_DAYS" -delete
echo "已清理超过 ${RETENTION_DAYS} 天的备份"

# 4. 可选异地推送（配置 REMOTE_HOST/REMOTE_DIR 时启用 rsync）
if [ -n "$REMOTE_HOST" ] && [ -n "$REMOTE_DIR" ]; then
  rsync -avz "$BACKUP_FILE" "${REMOTE_HOST}:${REMOTE_DIR}/" 2>/dev/null && echo "已同步到异地 $REMOTE_HOST" || echo "警告: 异地同步失败"
fi

echo "=== 备份完成 ==="
