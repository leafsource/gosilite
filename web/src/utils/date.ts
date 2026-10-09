// 获取常用时间
import dayjs from 'dayjs';

export const LAST_7_DAYS = [
  dayjs().subtract(7, 'day').format('YYYY-MM-DD'),
  dayjs().subtract(1, 'day').format('YYYY-MM-DD'),
];

/** 今天（YYYY-MM-DD 起止），用于日志类页面默认查询范围，防止全表扫描 */
export const todayRange = (): [string, string] => {
  const d = dayjs().format('YYYY-MM-DD');
  return [d, d];
};

export const LAST_30_DAYS = [
  dayjs().subtract(30, 'day').format('YYYY-MM-DD'),
  dayjs().subtract(1, 'day').format('YYYY-MM-DD'),
];
