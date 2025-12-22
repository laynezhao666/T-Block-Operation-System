import qs from 'qs';
export default {
  data() {
    return {
      showTitle: true,
      fullscreen: !!($('.is-hide-main-nav').length),
    };
  },
  mounted() {
    // window.addEventListener('resize', this.resizeTree);
    if ('_pn_' in qs.parse(location.href.split('?')[1])) {
      this.showTitle = false;
    }
  },
  methods: {
    toggle() {
      if (this.fullscreen) {
        // eslint-disable-next-line no-underscore-dangle
        window.__SwitchMainNavStatus(true);
        this.fullscreen = false;
      } else {
        const { search, href } = location;

        search.includes('_pn_') ? window.open(href) : window.open(`${href}?_pn_`);
      }
    },
  },
};
