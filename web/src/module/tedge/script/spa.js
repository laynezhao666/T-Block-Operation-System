import Vue from 'vue';
import entry from 'common/script/vue-entry';
import element from '@tencent/TNWeb-ui';

Vue.use(element);

Vue.filter('formatNum', (value) => {
  if (value === '****') {
    return value;
  }
  const num = (value || 0).toString();
  let index = num.length;
  if (num.indexOf('.') > -1) {
    index = num.indexOf('.');
  }
  let int = num.substr(0, index);
  const floor = num.substr(index);
  let result = '';
  while (int.length > 3) {
    result = `,${int.slice(-3)}${result}`;
    int = int.slice(0, int.length - 3);
  }
  if (int) {
    result = int + result + floor;
  }
  return result;
});

export default (EntryComp) => {
  entry(EntryComp, {
  });
};

export * from 'common/script/vue-entry';
