export function getMozuId() {
  // eslint-disable-next-line no-underscore-dangle
  const mozu = window.__GetFrameDataByKey('curMozuData');
  // eslint-disable-next-line radix
  return (mozu && mozu.id && parseInt(mozu.id)) || 326;
}
export function getMozuName() {
  // eslint-disable-next-line no-underscore-dangle
  const mozu = window.__GetFrameDataByKey('curMozuData');
  return (mozu && mozu.name) || '';
}
export function getEndTime() {
  const date = new Date();
  const year = date.getFullYear();
  const month = (date.getMonth() + 1) < 10 ? `0${date.getMonth() + 1}` : (date.getMonth() + 1);
  const day = date.getDate() < 10 ? `0${date.getDate()}` : date.getDate();
  const hour = date.getHours() < 10 ? `0${date.getHours()}` : date.getHours();
  const getMinutes = date.getMinutes() < 10 ? `0${date.getMinutes()}` : date.getMinutes();
  return `${year}${month}${day}${hour}${getMinutes}00`;
}
export function getStartTime() {
  const dateTime = new Date().getTime() - (3600000 * 24);
  const date = new Date(dateTime);
  const year = date.getFullYear();
  const month = (date.getMonth() + 1) < 10 ? `0${date.getMonth() + 1}` : (date.getMonth() + 1);
  const day = date.getDate() < 10 ? `0${date.getDate()}` : date.getDate();
  const hour = date.getHours() < 10 ? `0${date.getHours()}` : date.getHours();
  const getMinutes = date.getMinutes() < 10 ? `0${date.getMinutes()}` : date.getMinutes();
  return `${year}${month}${day}${hour}${getMinutes}00`;
}
export function getNowTime() {
  const dateTime = new Date().getTime();
  const date = new Date(dateTime);
  const year = date.getFullYear();
  const month = (date.getMonth() + 1) < 10 ? `0${date.getMonth() + 1}` : (date.getMonth() + 1);
  const day = date.getDate() < 10 ? `0${date.getDate()}` : date.getDate();
  const hour = date.getHours() < 10 ? `0${date.getHours()}` : date.getHours();
  const getMinutes = date.getMinutes() < 10 ? `0${date.getMinutes()}` : date.getMinutes();
  const getSeconds = date.getSeconds() < 10 ? `0${date.getSeconds()}` : date.getSeconds();
  return `${year}-${month}-${day} ${hour}:${getMinutes}:${getSeconds}`;
}
export function sampling(interval = 1, list = []) {
  // interval间隔 list抽样数组 数组长度length
  const arr = [];
  for (let i = 0; i < list.length; i = i + interval) {
    arr.push(list[i]);
  }
  return arr;
}
