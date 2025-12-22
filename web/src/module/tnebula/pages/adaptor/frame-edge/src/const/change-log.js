import * as _ from 'lodash';
import dayjs from 'dayjs';

const ctx = require.context('./change-logs/', true, /\.json$/);

export const changeLogs = _.orderBy(ctx.keys().map(ctx), log => -dayjs(log.timestamp).toDate().getTime());

if (changeLogs.length > 0) {
  // 最新的的版本设置
  changeLogs[0].type = 'primary';
}

export const latestVersion = changeLogs[0]?.title;
