/**
 * 自定义展示内容
 */

export function set(key, data) {
  localStorage.setItem(key, JSON.stringify(data));
}
export function get(key) {
  const text = localStorage.getItem(key);
  let rst;
  try {
    rst = JSON.parse(text);
  } catch (e) {
    rst = text || {};
  }
  return rst;
}
