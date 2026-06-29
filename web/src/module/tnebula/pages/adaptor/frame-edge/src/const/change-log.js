import * as _ from 'lodash';
import dayjs from 'dayjs';

let changeLogs = [];
let latestVersion = '';

try {
  const ctx = require.context('./change-logs/', true, /\.json$/);
  changeLogs = _.orderBy(ctx.keys().map(ctx), log => -dayjs(log.timestamp).toDate().getTime());

  if (changeLogs.length > 0) {
    // 最新的的版本设置
    changeLogs[0].type = 'primary';
  }

  latestVersion = changeLogs[0]?.title || '';
} catch (e) {
  // change-logs 目录为空或不存在时不影响应用启动
  console.warn('change-logs not found, skipping:', e.message);
}

export { changeLogs, latestVersion };
