export function getText(data, unit = '') {
  if (data === null) {
    return '--';
  } else {
    return `${data}${unit}`;
  }
}
