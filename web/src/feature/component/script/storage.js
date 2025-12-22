import Cookies from 'js-cookie';

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

export function getCookie(key) {
  return Cookies.get(key) || '';
}
