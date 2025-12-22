import simpleEntry from '@@/script/spa';
import page from 'feature/warning/warning-strategy/index.vue';

export default simpleEntry({
  render: h => h(page, {
    props: {
      rights: 0b11011,
    },
  }),
});

export * from '@@/script/spa';
